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
	"strings"
	"unicode/utf8"

	"github.com/alranel/go-vcsurl/v2"
	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
	urlutil "github.com/italia/publiccode-parser-go/v4/internal"
	publiccodeValidator "github.com/italia/publiccode-parser-go/v4/validators"
)

type ParserConfig struct {
	// DisableNetwork disables all network tests (URL existence and Oembed). This
	// results in much faster parsing.
	DisableNetwork bool

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

// NewParser initializes and returns a new Parser object following the settings in
// ParserConfig.
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

func NewDefaultParser() (*Parser, error) {
	return NewParser(ParserConfig{})
}

// ParseStream reads the data and tries to parse it. Returns an error if fails.
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
	// var node yaml.Node
	//
	// d := yaml.NewDecoder(bytes.NewReader(b))
	// d.KnownFields(true)
	// err = d.Decode(&node)
	//
	// if err == nil && len(node.Content) > 0 {
	// 	node = *node.Content[0]
	// } else {
	// 	// YAML is malformed
	// 	return nil, ValidationResults{toValidationError(err.Error(), nil)}
	// }
	//
	path, _ := yaml.PathString("$.publiccodeYmlVersion")

	var node ast.Node
	if node, err = path.ReadNode(bytes.NewReader(b)); err != nil {
		return nil, ValidationResults{newValidationError("publiccodeYmlVersion", "publiccodeYmlVersion is a required field")}
	}
	version := node.GetToken().Value

	var ve ValidationResults

	if !slices.Contains(SupportedVersions, version) {
		return nil, ValidationResults{
			newValidationError("publiccodeYmlVersion", fmt.Sprintf(
				"unsupported version: '%s'. Supported versions: %s",
				version,
				strings.Join(SupportedVersions, ", "),
			)),
		}
	}

	if slices.Contains(SupportedVersions, version) && !strings.HasPrefix(version, "0.4") {
		position := node.GetToken().Position
		latestVersion := SupportedVersions[len(SupportedVersions)-1]

		ve = append(ve, ValidationWarning{
			Key: "publiccodeYmlVersion",
			Description: fmt.Sprintf(
				"v%s is not the latest version, use '%s'. Parsing this file as v%s.",
				version,
				latestVersion,
				latestVersion,
			),
			Line:   position.Line,
			Column: position.Column,
		})
	}

	var publiccode PublicCode

	var validateFields validateFn

	var decodeResults ValidationResults

	if version[0] == '0' {
		v0 := PublicCodeV0{}
		validateFields = validateFieldsV0

		decodeResults = decode(b, &v0)
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
		var ast ast.Node
		_ = yaml.UnmarshalWithOptions(b, &ast, yaml.DisallowUnknownField())
		// TODO: err

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

			path, e := yaml.PathString("$." + key)
			// TODO: err

			if node, e = path.FilterNode(ast); e != nil {
				return nil, ValidationResults{newValidationError("XXX", "Xx")}
			}

			position := node.GetToken().Position

			ve = append(ve, ValidationError{
				Key:         key,
				Description: sb.String(),
				Line:        position.Line,
				Column:      position.Column,
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
				Key:         "url",
				Description: fmt.Sprintf("failed to get raw URL for code repository at %s: %s", publiccode.Url(), err),
				Line:        line,
				Column:      column,
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
		}()

		stream = resp.Body
	}

	return p.ParseStream(stream)
}

// Decode the YAML into a PublicCode structure, so we get type errors
func decode[T any](data []byte, publiccode *T) ValidationResults {
	var ve ValidationResults

	if err := yaml.UnmarshalWithOptions(data, &publiccode, yaml.DisallowUnknownField()); err != nil {
		switch err := err.(type) {
		case *yaml.TypeError:
			token := err.Token

			ve = append(ve, ValidationError{
				// token is the wrong type token,
				// token.Prev is ":"
				// token.Prev.Prev is the actual key
				Key:         token.Prev.Prev.Value,
				Description: "wrong type for this field",
				Line:        token.Position.Line,
				Column:      token.Position.Column,
			})
		default:
			ve = append(ve, newValidationError("", err.Error()))
		}
	}

	return ve
}

type keyFinder struct {
	key    string
	result *ast.Node
}

func (v keyFinder) Visit(node ast.Node) ast.Visitor {
	if node.GetToken().Value == v.key {
		v.result = &node
		return nil
	}

	return v
}

func getPositionInFile(key string, node ast.Node) (int, int) {
	finder := keyFinder{key: key}

	ast.Walk(finder, node)

	// This should not happen, but let's be defensive
	if finder.result == nil {
		return 0, 0
	}

	n := *finder.result
	position := n.GetToken().Position

	return position.Line, position.Column

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
