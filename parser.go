package publiccode

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
	"fmt"

	"gopkg.in/yaml.v3"
	"github.com/go-playground/validator/v10"
	publiccodeValidator "github.com/italia/publiccode-parser-go/validators"
)

// Parser is a helper class for parsing publiccode.yml files.
type Parser struct {
	PublicCode PublicCode

	// LocalBasePath is a filesystem path pointing to the directory where the
	// publiccode.yml is located. It's used as a base for relative paths. If
	// left empty, RemoteBaseURL will be used.
	LocalBasePath string

	// RemoteBaseURL is the URL pointing to the raw directory where the publiccode.yml
	// file is located. It's used for validating abolute URLs and as a base for
	// relative paths. If left empty, absolute URLs will not be validated and
	// no remote validation of files with relative paths will be performed. If
	// not left empty, publiccode.yml keys with relative paths will be turned
	// into absolute URLs.
	// (eg: https://raw.githubusercontent.com/gith002/Medusa/master)
	RemoteBaseURL string

	// DisableNetwork disables all network tests (URL existence and Oembed). This
	// results in much faster parsing.
	DisableNetwork bool

	// Domain will have domain specific settings, including basic auth if provided
	// this will avoid strong quota limit imposed by code hosting platform
	Domain Domain
}

// Domain is a single code hosting service.
type Domain struct {
	// Domains.yml data
	Host        string   `yaml:"host"`
	UseTokenFor []string `yaml:"use-token-for"`
	BasicAuth   []string `yaml:"basic-auth"`
}

// ParseInDomain wrapper func to be in domain env
func (p *Parser) ParseInDomain(in []byte, host string, utf []string, ba []string) error {
	p.Domain = Domain{
		Host:        host,
		UseTokenFor: utf,
		BasicAuth:   ba,
	}

	return p.Parse(in)
}

func getNodes(key string, node *yaml.Node) (*yaml.Node, *yaml.Node) {
	for i := 0; i < len(node.Content); i += 2 {
		childNode := *node.Content[i]

		if childNode.Value == key {
			return &childNode, node.Content[i + 1]
		}
	}

	return nil, nil
}

func getPositionInFile(key string, node yaml.Node) (int, int) {
	var n *yaml.Node = &node

	keys := strings.Split(key, ".")
	for _, path := range keys[:len(keys) - 1] {
		_, n = getNodes(path, n)

		// This should not happen, but let's be defensive
		if (n == nil) {
			return 0, 0
		}
	}

	parentNode := n

	n, _ = getNodes(keys[len(keys) - 1], n)

	if (n != nil) {
		return n.Line, n.Column
	} else {
		return parentNode.Line, parentNode.Column
	}
}

// getKeyAtLine returns the key name at line "line" for the YAML document
// represented at parentNode.
func getKeyAtLine(parentNode yaml.Node, line int, path string) string {
	var key = path

	for i, currNode := range parentNode.Content {
		// If this node is a mapping and the index is odd it means
		// we are not looking at a key, but at its value. Skip it.
		if parentNode.Kind == yaml.MappingNode && i%2 != 0 && currNode.Kind == yaml.ScalarNode {
			continue
		}

		// This node is a key of a mapping type
		if parentNode.Kind == yaml.MappingNode && i%2 == 0 {
			if path == "" {
				key = currNode.Value
			} else {
				key = fmt.Sprintf("%s.%s", path, currNode.Value)
			}
		}

		// We want the scalar node (ie. key) not the mapping node which
		// doesn't have a tag name even if it has the same line number
		if currNode.Line == line && parentNode.Kind == yaml.MappingNode && currNode.Kind == yaml.ScalarNode {
			return key
		}

		if currNode.Kind != yaml.ScalarNode {
			if k := getKeyAtLine(*currNode, line, key); k != "" {
				return k
			}
		}
	}

	return ""
}

func toValidationError(errorText string, node yaml.Node) ValidationError {
	r := regexp.MustCompile(`^(line ([0-9]+): )`)
	matches := r.FindStringSubmatch(errorText)

	line := 0
	if (len(matches) > 1) {
		line, _ = strconv.Atoi(matches[2])
		errorText = strings.ReplaceAll(errorText, matches[1], "")
	}

	// Transform unmarshalling errors messages to a user friendlier message
	r = regexp.MustCompile("^cannot unmarshal")
	if r.MatchString(errorText) {
		errorText = "wrong type for this field"
	}

	key := getKeyAtLine(node, line, "")

	return ValidationError{
		Key: key,
		Description: errorText,
		Line: line,
		Column: 1,
	}
}

// Parse loads the yaml bytes and tries to parse it. Return an error if fails.
func (p *Parser) Parse(in []byte) error {
	var ve ValidationErrors

	if !utf8.Valid(in) {
		ve = append(ve,newValidationError("", "Invalid UTF-8"))
		return ve
	}

	// First, decode the YAML into yaml.Node so we can access line and column
	// numbers.
	var node yaml.Node

	d := yaml.NewDecoder(bytes.NewReader(in))
	d.KnownFields(true)
	d.Decode(&node)

	node = *node.Content[0]

	// Decode the YAML into a PublicCode structure, so we get type errors
	d = yaml.NewDecoder(bytes.NewReader(in))
	d.KnownFields(true)
	if err := d.Decode(&p.PublicCode); err != nil {
		switch err.(type) {
			case *yaml.TypeError:
				for _, errorText := range err.(*yaml.TypeError).Errors {
					ve = append(ve, toValidationError(errorText, node))
				}
			default:
				ve = append(ve, newValidationError("", err.Error()))
		}
	}

	validate := publiccodeValidator.New()

	err := validate.Struct(p.PublicCode)
	if err != nil {
		tagMap := map[string]string{
			"gt": "must be more than",
			"oneof": "must be one of the following:",
			"email": "must be a valid email",
			"date": "must be a date with format 'YYYY-MM-DD'",
			"umax": "must be less or equal than",
			"umin": "must be more or equal than",
			"is_category_v0_2": "must be a valid category",
			"is_scope_v0_2": "must be a valid scope",
			"iso3166_1_alpha2_lowercase": "must be a valid lowercase ISO 3166-1 alpha-2 two-letter country code",
			"bcp47": "must be a valid BCP 47 language",
		}
		for _, err := range err.(validator.ValidationErrors) {
			var sb strings.Builder

			tag, ok := tagMap[err.ActualTag()]
			if !ok {
				tag = err.ActualTag()
			}

			sb.WriteString(tag)

			// condition parameters, e.g. oneof=red blue -> red blue
			if err.Param() != "" {
				sb.WriteString(" " + err.Param())
			}

			// TODO: find a cleaner way
			key := strings.Replace(err.Namespace(), "PublicCode.", "", 1)
			m := regexp.MustCompile(`\[([[:alpha:]]+)\]`)
			key = m.ReplaceAllString(key, ".$1")

			line, column := getPositionInFile(key, node)

			ve = append(ve, ValidationError{
				Key: key,
				Description: sb.String(),
				Line: line,
				Column: column,
			})
		}
	}

	err = p.validateFields()
	if err != nil {
		for _, err := range err.(ValidationErrors) {
			err.Line, err.Column = getPositionInFile(err.Key, node)

			ve = append(ve, err)
		}
	}

	if (len(ve) == 0) {
		return nil
	}

	return ve
}

// ParseFile loads a publiccode.yml file from a given file path.
func (p *Parser) ParseFile(file string) error {
	// Read data.
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	return p.Parse(data)
}

// ParseRemoteFile loads a publiccode.yml file from its raw URL.
func (p *Parser) ParseRemoteFile(url string) error {
	// Read data.
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return p.Parse(data)
}

// NewParser initializes a new Parser object and returns it.
func NewParser() *Parser {
	var p Parser

	return &p
}

// ToYAML converts parser.PublicCode into YAML again.
func (p *Parser) ToYAML() ([]byte, error) {
	// Make a copy and set the latest versions
	pc2 := p.PublicCode
	pc2.PubliccodeYamlVersion = Version
	pc2.It.CountryExtensionVersion = ExtensionITVersion
	return yaml.Marshal(pc2)
}
