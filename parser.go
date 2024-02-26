package publiccode

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/alranel/go-vcsurl/v2"
	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"

	publiccodeValidator "github.com/italia/publiccode-parser-go/v4/validators"
	urlutil "github.com/italia/publiccode-parser-go/v4/internal"
)

// ParserConfig defines the configuration options for the Parser
type ParserConfig struct {
	// DisableNetwork disables all network tests (fe. URL existence and Oembed). This
	// results in much faster parsing.
    DisableNetwork bool

	// Domain will have domain specific settings, including basic auth if provided
	// this will avoid strong quota limit imposed by code hosting platform
    Domain Domain

	// The name of the branch used to check for existence of the files referenced
	// in the publiccode.yml
    Branch string

	// The path or URL used as base of relative files in publiccode.yml (eg. logo). If
	// given, it will be used for existence checks and such, instead of the `url` key
	// in the publiccode.yml file
	BaseURL string
}

// Parser is a helper class for parsing publiccode.yml files.
type Parser struct {
	disableNetwork bool
	domain         Domain
	branch         string
	baseURL        *url.URL
}

// Domain is a single code hosting service.
type Domain struct {
	// Domains.yml data
	Host        string   `yaml:"host"`
	UseTokenFor []string `yaml:"use-token-for"`
	BasicAuth   []string `yaml:"basic-auth"`
}

// NewParser creates and returns a new Parser instance based on the provided ParserConfig.
//
// It returns a pointer to the newly created Parser, if successful, and an error.
// The error is non-nil if there was an issue initializing the Parser.
//
// Example usage:
//
//  config := ParserConfig{
//      DisableNetwork: true
//  }
//  parser, err := NewParser(config)
//  if err != nil {
//      log.Fatalf("Failed to create parser: %v", err)
//  }
//  // Use parser...
func NewParser(config ParserConfig) (*Parser, error) {
	parser := Parser{
		disableNetwork: config.DisableNetwork,
		domain:         config.Domain,
		branch:         config.Branch,
	}

	if config.BaseURL != "" {
		var err error
		if parser.baseURL, err = toURL(config.BaseURL); err != nil {
			return nil, err
		}
	}

	return &parser, nil
}

// NewDefaultParser creates and returns a new Parser instance, using the default config.
//
// It returns a pointer to the newly created Parser, if successful, and an error.
// The error is non-nil if there was an issue initializing the Parser.
//
// The default config is ParserConfig's fields zero values.
//
// Example usage:
//
//  parser, err := NewDefaultParser()
//  if err != nil {
//      log.Fatalf("Failed to create parser: %v", err)
//  }
//  // Use parser...
func NewDefaultParser() (*Parser, error) {
	return NewParser(ParserConfig{})
}

// ParseStream reads from the provided io.Reader and attempts to parse the input
// stream into a PublicCode object.
//
// Returns a non-nil error if there is an issue with the parsing process and a
// a struct implementing the Publiccode interface, depending on the version
// of the publiccode.yml file parsed.
//
// The only possible type returned is V0, representing v0.* files.
//
// PublicCode can be nil when there are major parsing issues.
// It can also be non-nil even in presence of errors: in that case the returned struct
// is filled as much as possible, based on what is successful parsed.
//
// Example usage:
//
//  file, err := os.Open("publiccode.yml")
//  if err != nil {
//      log.Fatalf("Failed to open file: %v", err)
//  }
//  defer file.Close()
//
//  publicCode, err := parser.ParseStream(file)
//  if err != nil {
//      log.Printf("Parsing errors: %v", err)
//  }
//  if publicCode != nil {
//      switch pc := publicCode.(type) {
//      case *publiccode.V0:
//          fmt.Println(pc)
//      default:
//          log.Println("PublicCode parsed, but unexpected type found.")
//      }
//  }
func (p *Parser) ParseStream(in io.Reader) (PublicCode, error) {
	b, err := io.ReadAll(in)
	if err != nil {
		return nil, ValidationResults{newValidationError("", fmt.Sprintf("Can't read the stream: %v", err))}
	}

	if !utf8.Valid(b) {
		return nil, ValidationResults{newValidationError("", "Invalid UTF-8")}
	}

	// First, decode the YAML into yaml.Node so we can access line and column
	// numbers.
	var node yaml.Node

	d := yaml.NewDecoder(bytes.NewReader(b))
	d.KnownFields(true)
	err = d.Decode(&node)

	if err == nil && len(node.Content) > 0 {
		node = *node.Content[0]
	} else {
		// YAML is malformed
		return nil, ValidationResults{toValidationError(err.Error(), nil)}
	}

	_, version := getNodes("publiccodeYmlVersion", &node)
	if version == nil {
		return nil, ValidationResults{newValidationError("publiccodeYmlVersion", "required")}
	}
	if version.ShortTag() != "!!str" {
		line, column := getPositionInFile("publiccodeYmlVersion", node)

		return nil, ValidationResults{ValidationError{
			Key: "publiccodeYmlVersion",
			Description: "wrong type for this field",
			Line: line,
			Column: column,
		}}
	}

	var ve ValidationResults

	if slices.Contains(SupportedVersions, version.Value) && !strings.HasPrefix(version.Value, "0.3") {
		latestVersion := SupportedVersions[len(SupportedVersions)-1]
		line, column := getPositionInFile("publiccodeYmlVersion", node)

		ve = append(ve, ValidationWarning{
			Key: "publiccodeYmlVersion",
			Description: fmt.Sprintf(
				"v%s is not the latest version, use '%s'. Parsing this file as v%s.",
				version.Value,
				latestVersion,
				latestVersion,
			),
			Line: line,
			Column: column,
		})
	}

	var publiccode PublicCode
	var validateFields validateFn
	var decodeResults ValidationResults

	if version.Value[0] == '0' {
		v0 := V0{}
		validateFields = validateFieldsV0

		decodeResults = decode(b, &v0, node)
		publiccode = v0
	} else {
		v1 := V1{}
		validateFields = validateFieldsV1

		decodeResults = decode(b, &v1, node)
		publiccode = v1
	}

	if decodeResults != nil {
		ve = append(ve, decodeResults...)
	}

	validate := publiccodeValidator.New()

	err = validate.Struct(publiccode)
	if err != nil {
		tagMap := map[string]string{
			"gt": "must be more than",
			"oneof": "must be one of the following:",
			"email": "must be a valid email",
			"date": "must be a date with format 'YYYY-MM-DD'",
			"umax": "must be less or equal than",
			"umin": "must be more or equal than",
			"url_http_url": "must be an HTTP URL",
			"url_url": "must be a valid URL",
			"is_category_v0_2": "must be a valid category",
			"is_scope_v0_2": "must be a valid scope",
			"is_italian_ipa_code": "must be a valid Italian Public Administration Code (iPA)",
			"iso3166_1_alpha2_lowercase": "must be a valid lowercase ISO 3166-1 alpha-2 two-letter country code",
			"bcp47_language_tag": "must be a valid BCP 47 language",
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
			key := strings.Replace(
				err.Namespace(),
				fmt.Sprintf("V%d.", publiccode.Version()),
				"",
				1,
			)
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

	// baseURL was not set to a local path, let's autodetect it from the
	// publiccode.yml url key
	//
	// We need the baseURL to perform network checks.
	if p.baseURL == nil && !p.disableNetwork && publiccode.Url() != nil {
		rawRoot, err := vcsurl.GetRawRoot((*url.URL)(publiccode.Url()), p.branch)
		if err != nil {
			line, column := getPositionInFile("url", node)

			ve = append(ve, ValidationError{
				Key: "url",
				Description: fmt.Sprintf("failed to get raw URL for code repository at %s: %s", publiccode.Url(), err),
				Line: line,
				Column: column,
			})

			// Return early because proceeding with no baseURL would result in a lot
			// of duplicate errors stemming from its absence.
			return publiccode, ve
		}

		p.baseURL = rawRoot
	}

	if err = validateFields(publiccode, *p, !p.disableNetwork); err != nil {
		for _, err := range err.(ValidationResults) {
			switch err := err.(type) {
			case ValidationError:
				err.Line, err.Column = getPositionInFile(err.Key, node)
				ve = append(ve, err)
			case ValidationWarning:
				err.Line, err.Column = getPositionInFile(err.Key, node)
				ve = append(ve, err)
			}
		}
	}

	if (len(ve) == 0) {
		return publiccode, nil
	}

	return publiccode, ve
}

// Parse reads from the provided uri and attempts to parse it into
// a PublicCode object.
//
// Returns a non-nil error if there is an issue with the parsing process and a
// a struct implementing the Publiccode interface, depending on the version
// of the publiccode.yml file parsed.
//
// The only possible type returned is V0, representing v0.* files.
//
// PublicCode can be nil when there are major parsing issues.
// It can also be non-nil even in presence of errors: in that case the returned struct
// is filled as much as possible, based on what is successful parsed.
//
// Example usage:
//
//  // publicCode, err := parser.Parse("file:///app/publiccode.yml")
//  publicCode, err := parser.Parse("https://foobar.example.org/publiccode.yml")
//  if err != nil {
//      log.Printf("Parsing errors: %v", err)
//  }
//  if publicCode != nil {
//      switch pc := publicCode.(type) {
//      case *publiccode.V0:
//          fmt.Println(pc)
//      default:
//          log.Println("PublicCode parsed, but unexpected type found.")
//      }
//  }
func (p *Parser) Parse(uri string) (PublicCode, error) {
	var stream io.Reader

	url, err := toURL(uri);
	if err != nil {
		return nil, fmt.Errorf("Invalid URL '%s': %w", uri, err)
	}

	if url.Scheme == "file" {
		stream, err = os.Open(url.Path)
		if err != nil {
			return nil, fmt.Errorf("Can't open file '%s': %w", url.Path, err)
		}
	} else {
		resp, err := http.Get(uri)
		if err != nil {
			return nil, fmt.Errorf("Can't GET '%s': %w", uri, err)
		}
		defer resp.Body.Close()

		stream = resp.Body
	}

	return p.ParseStream(stream)
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

func toValidationError(errorText string, node *yaml.Node) ValidationError {
	r := regexp.MustCompile(`(line ([0-9]+): )`)
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

	var key string
	if node != nil {
		key = getKeyAtLine(*node, line, "")
	}

	return ValidationError{
		Key: key,
		Description: errorText,
		Line: line,
		Column: 1,
	}
}

// Decode the YAML into a PublicCode structure, so we get type errors
func decode[T any](data []byte, publiccode *T, node yaml.Node) ValidationResults {
	var ve ValidationResults

	d := yaml.NewDecoder(bytes.NewReader(data))
	d.KnownFields(true)
	if err := d.Decode(&publiccode); err != nil {
		switch err := err.(type) {
			case *yaml.TypeError:
				for _, errorText := range err.Errors {
					ve = append(ve, toValidationError(errorText, &node))
				}
			default:
				ve = append(ve, newValidationError("", err.Error()))
		}
	}

	return ve
}

// Turn the input into a valid URL
func toURL(file string) (*url.URL, error) {
	if _, u := urlutil.IsValidURL(file); u != nil {
		return u, nil
	}

	if path, err := filepath.Abs(file); err == nil {
		return &url.URL{Scheme: "file", Path: path }, nil
	} else {
		return nil, err
	}
}
