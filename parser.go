package publiccode

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/alranel/go-vcsurl/v2"
	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	urlutil "github.com/italia/publiccode-parser-go/v4/internal"
	publiccodeValidator "github.com/italia/publiccode-parser-go/v4/validators"
	"gopkg.in/yaml.v3"
)

type ParserConfig struct {
	// DisableNetwork disables all network tests (eg. URL existence). This
	// results in much faster parsing.
	DisableNetwork bool

	// DisableExternalChecks disables ALL the additional checks on external files
	// and resources, local or remote (eg. existence, images actually being images, etc.).
	//
	// It implies DisableNetwork = true.
	DisableExternalChecks bool

	// Domain will have domain specific settings, including basic auth if provided
	// this will avoid strong quota limit imposed by code hosting platform
	Domain Domain

	// The name of the branch used to check for existence of the files referenced
	// in the publiccode.yml
	Branch string

	// The URL used as base of relative files in publiccode.yml (eg. logo)
	// It can be a local file with the 'file' scheme.
	BaseURL string
}

// Parser is a helper class for parsing publiccode.yml files.
type Parser struct {
	disableNetwork        bool
	disableExternalChecks bool
	domain                Domain
	branch                string
	baseURL               *url.URL
	fileURL               *url.URL

	// This is the baseURL we'll try to compute and use between
	// Parse{,Stream)() calls.
	//
	// XXX: It's an hack and it requires to fix the design mistake in the public API.
	// This makes Parse{,Stream}() not thread safe.
	currentBaseURL *url.URL
}

// Domain is a single code hosting service.
type Domain struct {
	// Domains.yml data
	Host        string   `yaml:"host"`
	UseTokenFor []string `yaml:"use-token-for"`
	BasicAuth   []string `yaml:"basic-auth"`
}

// NewParser initializes and returns a new Parser object following the settings in
// ParserConfig.
func NewParser(config ParserConfig) (*Parser, error) {
	if config.DisableExternalChecks {
		config.DisableNetwork = true
	}

	parser := Parser{
		disableNetwork:        config.DisableNetwork,
		disableExternalChecks: config.DisableExternalChecks,
		domain:                config.Domain,
		branch:                config.Branch,
	}

	if config.BaseURL != "" {
		var err error
		if parser.baseURL, err = toURL(config.BaseURL); err != nil {
			return nil, err
		}
	}

	return &parser, nil
}

func NewDefaultParser() (*Parser, error) {
	return NewParser(ParserConfig{})
}

// ParseStream reads the data and tries to parse it. Returns an error if fails.
func (p *Parser) ParseStream(in io.Reader) (PublicCode, error) { //nolint:maintidx
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
		return nil, ValidationResults{newValidationError("publiccodeYmlVersion", "publiccodeYmlVersion is a required field")}
	}

	if version.ShortTag() != "!!str" {
		line, column := getPositionInFile("publiccodeYmlVersion", node)

		return nil, ValidationResults{ValidationError{
			Key:         "publiccodeYmlVersion",
			Description: "wrong type for this field",
			Line:        line,
			Column:      column,
		}}
	}

	if !slices.Contains(SupportedVersions, version.Value) {
		return nil, ValidationResults{
			newValidationError("publiccodeYmlVersion", fmt.Sprintf(
				"unsupported version: '%s'. Supported versions: %s",
				version.Value,
				strings.Join(SupportedVersions, ", "),
			)),
		}
	}

	var ve ValidationResults

	if slices.Contains(SupportedVersions, version.Value) && version.Value != "0" && !strings.HasPrefix(version.Value, "0.4") {
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
			Line:   line,
			Column: column,
		})
	}

	var publiccode PublicCode

	var validateFields validateFn

	var decodeResults ValidationResults

	if version.Value[0] == '0' {
		v0 := PublicCodeV0{}
		validateFields = validateFieldsV0

		decodeResults = decode(b, &v0, node)
		publiccode = v0
	}

	// When publiccode.yml v1.x is released, the code will look
	// like this:
	// } else {
	// 	v1 := PublicCodeV1{}
	// 	validateFields = validateFieldsV1
	//
	// 	decodeResults = decode(b, &v1, node)
	// 	publiccode = v1
	// }

	if decodeResults != nil {
		ve = append(ve, decodeResults...)
	}

	validate := publiccodeValidator.New()

	en := en.New()
	uni := ut.New(en, en)

	trans, _ := uni.GetTranslator("en")
	_ = en_translations.RegisterDefaultTranslations(validate, trans)
	_ = publiccodeValidator.RegisterLocalErrorMessages(validate, trans)

	err = validate.Struct(publiccode)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			var sb strings.Builder

			sb.WriteString(err.Translate(trans))

			// TODO: find a cleaner way
			key := strings.Replace(
				err.Namespace(),
				fmt.Sprintf("PublicCodeV%d.", publiccode.Version()),
				"",
				1,
			)
			m := regexp.MustCompile(`\[([[:alpha:]]+)\]`)
			key = m.ReplaceAllString(key, ".$1")

			line, column := getPositionInFile(key, node)

			ve = append(ve, ValidationError{
				Key:         key,
				Description: sb.String(),
				Line:        line,
				Column:      column,
			})
		}
	}

	p.currentBaseURL = nil

	// baseURL was not set by the user (with ParserConfig{BaseURL: "..."})),
	// We need a base URL to perform external checks on relative files (eg. logo).
	if p.baseURL == nil {
		// If we parsed from an actual local or remote file, use its dir
		if p.fileURL != nil {
			u := *p.fileURL

			p.currentBaseURL = &u
			p.currentBaseURL.Path = path.Dir(p.fileURL.Path)
		}
	} else {
		u := *p.baseURL

		p.currentBaseURL = &u
	}

	// Still no base URL: we parsed from a stream, try to use the publiccode.yml's `url` field
	if p.currentBaseURL == nil && !p.disableNetwork && publiccode.Url() != nil {
		rawRoot, err := vcsurl.GetRawRoot((*url.URL)(publiccode.Url()), p.branch)
		if err != nil {
			line, column := getPositionInFile("url", node)

			ve = append(ve, ValidationError{
				Key:         "url",
				Description: fmt.Sprintf("failed to get raw URL for code repository at %s: %s", publiccode.Url(), err),
				Line:        line,
				Column:      column,
			})

			// Return early because proceeding with no base URL would result in a lot
			// of duplicate errors stemming from its absence.
			return publiccode, ve
		}

		p.currentBaseURL = rawRoot
	}

	// Still no base URL: DisableNetwork is true, use the current working directory as a fallback
	if p.currentBaseURL == nil {
		cwd, err := os.Getwd()
		if err != nil {
			ve = append(ve, newValidationError("", fmt.Sprintf("no baseURL set and failed to get working directory: %s", err)))

			return publiccode, ve
		}

		p.currentBaseURL = &url.URL{Scheme: "file", Path: cwd}
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

	// If the deprecated 'it' was used instead of 'IT', copy the data
	// in the canonical 'IT' field.
	if v0, ok := publiccode.(PublicCodeV0); ok {
		v0.IT = v0.It
	}

	if len(ve) == 0 {
		return publiccode, nil
	}

	return publiccode, ve
}

func (p *Parser) Parse(uri string) (PublicCode, error) {
	var stream io.Reader

	url, err := toURL(uri)
	if err != nil {
		return nil, fmt.Errorf("invalid URL '%s': %w", uri, err)
	}

	p.fileURL = url

	if url.Scheme == "file" {
		stream, err = os.Open(url.Path)
		if err != nil {
			return nil, fmt.Errorf("can't open file '%s': %w", url.Path, err)
		}
	} else {
		resp, err := http.Get(uri)
		if err != nil {
			return nil, fmt.Errorf("can't GET '%s': %w", uri, err)
		}

		defer func() {
			_ = resp.Body.Close()
			p.fileURL = nil
		}()

		stream = resp.Body
	}

	return p.ParseStream(stream)
}

func getNodes(key string, node *yaml.Node) (*yaml.Node, *yaml.Node) {
	for i := 0; i < len(node.Content); i += 2 {
		childNode := *node.Content[i]

		if childNode.Value == key {
			return &childNode, node.Content[i+1]
		}
	}

	return nil, nil
}

func getPositionInFile(key string, node yaml.Node) (int, int) {
	n := &node

	keys := strings.Split(key, ".")
	for _, path := range keys[:len(keys)-1] {
		_, n = getNodes(path, n)

		// This should not happen, but let's be defensive
		if n == nil {
			return 0, 0
		}
	}

	parentNode := n

	n, _ = getNodes(keys[len(keys)-1], n)

	if n != nil {
		return n.Line, n.Column
	} else {
		return parentNode.Line, parentNode.Column
	}
}

// getKeyAtLine returns the key name at line "line" for the YAML document
// represented at parentNode.
func getKeyAtLine(parentNode yaml.Node, line int, path string) string {
	key := path

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
	if len(matches) > 1 {
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
		Key:         key,
		Description: errorText,
		Line:        line,
		Column:      1,
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

func toURL(file string) (*url.URL, error) {
	if _, u := urlutil.IsValidURL(file); u != nil {
		return u, nil
	}

	if path, err := filepath.Abs(file); err == nil {
		return &url.URL{Scheme: "file", Path: path}, nil
	} else {
		return nil, err
	}
}
