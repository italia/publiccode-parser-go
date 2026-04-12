package publiccode

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/alranel/go-vcsurl/v2"
	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	yaml "github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
	httpclient "github.com/italia/httpclient-lib-go"
	urlutil "github.com/italia/publiccode-parser-go/v5/internal"
	publiccodeValidator "github.com/italia/publiccode-parser-go/v5/validators"
)

// Build Validator and Translator once at package init.
var (
	sharedValidate *validator.Validate
	sharedTrans    ut.Translator
)

func init() {
	sharedValidate = publiccodeValidator.New()

	enLocale := en.New()
	uni := ut.New(enLocale, enLocale)

	sharedTrans, _ = uni.GetTranslator("en")
	_ = en_translations.RegisterDefaultTranslations(sharedValidate, sharedTrans)
	_ = publiccodeValidator.RegisterLocalErrorMessages(sharedValidate, sharedTrans)
}

var reMapKey = regexp.MustCompile(`\[([[:alpha:]]+)\]`)

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

	// Timeout is the maximum duration for each HTTP request during external checks.
	// Defaults to 30s if zero.
	Timeout time.Duration
}

const defaultHTTPTimeout = 30 * time.Second

// Parser is a helper class for parsing publiccode.yml files.
type Parser struct {
	disableNetwork        bool
	disableExternalChecks bool
	domain                Domain
	branch                string
	baseURL               *url.URL
	client                *http.Client
	httpclient            *httpclient.Client
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

	timeout := config.Timeout
	if timeout == 0 {
		timeout = defaultHTTPTimeout
	}

	httpClient := &http.Client{Timeout: timeout}
	vcsurl.Client = httpClient
	p := Parser{
		disableNetwork:        config.DisableNetwork,
		disableExternalChecks: config.DisableExternalChecks,
		domain:                config.Domain,
		branch:                config.Branch,
		client:                httpClient,
		httpclient:            httpclient.NewClient(httpClient),
	}

	if config.BaseURL != "" {
		var err error
		if p.baseURL, err = toURL(config.BaseURL); err != nil {
			return nil, err
		}
	}

	return &p, nil
}

// NewDefaultParser initializes and returns a new Parser object with default settings.
func NewDefaultParser() (*Parser, error) {
	return NewParser(ParserConfig{})
}

// ParseStream reads the data and tries to parse it. Returns an error if fails.
func (p *Parser) ParseStream(in io.Reader) (PublicCode, error) {
	return p.parseStream(in, nil)
}

func (p *Parser) Parse(uri string) (PublicCode, error) {
	var stream io.ReadCloser

	fileURL, err := toURL(uri)
	if err != nil {
		return nil, fmt.Errorf("invalid URL '%s': %w", uri, err)
	}

	if fileURL.Scheme == "file" {
		stream, err = os.Open(fileURL.Path)
		if err != nil {
			return nil, fmt.Errorf("can't open file '%s': %w", fileURL.Path, err)
		}
	} else {
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, uri, nil)
		if err != nil {
			return nil, fmt.Errorf("can't build GET request for '%s': %w", uri, err)
		}

		resp, err := p.client.Do(req) //nolint:bodyclose,lll // bodyclose: closed via defer stream.Close() below; bodyclose won't realize it :(
		if err != nil {
			return nil, fmt.Errorf("can't GET '%s': %w", uri, err)
		}

		stream = resp.Body
	}

	defer stream.Close()

	return p.parseStream(stream, fileURL)
}

func (p *Parser) parseStream(in io.Reader, fileURL *url.URL) (PublicCode, error) { //nolint:maintidx
	b, err := io.ReadAll(in)
	if err != nil {
		return nil, ValidationResults{newValidationErrorf("", "Can't read the stream: %v", err)}
	}

	if !utf8.Valid(b) {
		return nil, ValidationResults{newValidationError("", "Invalid UTF-8")}
	}

	// Parse the YAML into an AST so we can look up line/column positions and
	// detect structural issues (multi-document, empty file, syntax errors).
	file, err := parser.ParseBytes(b, 0)
	if err != nil {
		var se *yaml.SyntaxError
		if errors.As(err, &se) {
			line := 0
			if tok := se.GetToken(); tok != nil {
				line = tok.Position.Line
			}

			return nil, ValidationResults{ValidationError{
				Key:         "",
				Description: se.GetMessage(),
				Line:        line,
				Column:      1,
			}}
		}

		return nil, ValidationResults{newValidationError("", err.Error())}
	}

	if len(file.Docs) == 0 || file.Docs[0].Body == nil {
		return nil, ValidationResults{newValidationError("publiccodeYmlVersion", "publiccodeYmlVersion is a required field")}
	}

	if len(file.Docs) > 1 {
		return nil, ValidationResults{newValidationError("", "multiple YAML documents in one file are not supported")}
	}

	// Extract publiccodeYmlVersion from the AST.
	versionPath, err := yaml.PathString("$.publiccodeYmlVersion")
	if err != nil {
		return nil, ValidationResults{newValidationError("", err.Error())}
	}

	versionNode, err := versionPath.FilterFile(file)
	if err != nil || versionNode == nil {
		return nil, ValidationResults{newValidationError("publiccodeYmlVersion", "publiccodeYmlVersion is a required field")}
	}

	strNode, ok := versionNode.(*ast.StringNode)
	if !ok {
		line, column := getPositionInFile("publiccodeYmlVersion", file)

		return nil, ValidationResults{ValidationError{
			Key:         "publiccodeYmlVersion",
			Description: "wrong type for this field",
			Line:        line,
			Column:      column,
		}}
	}

	version := strNode.Value

	if !slices.Contains(SupportedVersions, version) {
		return nil, ValidationResults{
			newValidationErrorf("publiccodeYmlVersion",
				"unsupported version: '%s'. Supported versions: %s",
				version,
				strings.Join(SupportedVersions, ", ")),
		}
	}

	var ve ValidationResults

	if slices.Contains(SupportedVersions, version) && version != "0" && !strings.HasPrefix(version, "0.5") {
		latestVersion := SupportedVersions[len(SupportedVersions)-1]
		line, column := getPositionInFile("publiccodeYmlVersion", file)

		ve = append(ve, ValidationWarning{
			Key: "publiccodeYmlVersion",
			Description: fmt.Sprintf(
				"v%s is not the latest version, use '%s'. Parsing this file as v%s.",
				version,
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

	if version[0] == '0' {
		v0 := &PublicCodeV0{}
		validateFields = validateFieldsV0

		decodeResults = decode(b, v0, file)
		publiccode = v0
	} else {
		v1 := &PublicCodeV1{}
		validateFields = validateFieldsV1

		decodeResults = decode(b, v1, file)
		publiccode = v1
	}

	if decodeResults != nil {
		ve = append(ve, decodeResults...)
	}

	err = sharedValidate.Struct(publiccode)
	if err != nil {
		var validationErrs validator.ValidationErrors
		if errors.As(err, &validationErrs) {
			for _, err := range validationErrs {
				key := strings.SplitN(err.Namespace(), ".", 2)[1]
				key = reMapKey.ReplaceAllString(key, ".$1")

				line, column := getPositionInFile(key, file)

				ve = append(ve, ValidationError{
					Key:         key,
					Description: err.Translate(sharedTrans),
					Line:        line,
					Column:      column,
				})
			}
		}
	}

	// Compute the base URL for this call without mutating p (goroutine-safety).
	var currentBaseURL *url.URL

	// baseURL was not set by the user (with ParserConfig{BaseURL: "..."})),
	// We need a base URL to perform external checks on relative files (eg. logo).
	if p.baseURL == nil {
		// If we parsed from an actual local or remote file, use its dir
		if fileURL != nil {
			u := *fileURL

			currentBaseURL = &u
			currentBaseURL.Path = path.Dir(fileURL.Path)
		}
	} else {
		u := *p.baseURL

		currentBaseURL = &u
	}

	// Still no base URL: we parsed from a stream, try to use the publiccode.yml's `url` field
	if currentBaseURL == nil && !p.disableNetwork && publiccode.Url() != nil {
		rawRoot, err := vcsurl.GetRawRoot((*url.URL)(publiccode.Url()), p.branch)
		if err != nil {
			line, column := getPositionInFile("url", file)

			ve = append(ve, ValidationError{
				Key:         "url",
				Description: fmt.Sprintf("failed to get raw URL for code repository at %s: %s", publiccode.Url(), err),
				Line:        line,
				Column:      column,
			})

			// Return early because proceeding with no base URL would result in a lot
			// of duplicate errors stemming from its absence.
			return asPublicCode(publiccode), ve
		}

		currentBaseURL = rawRoot
	}

	// Still no base URL: DisableNetwork is true, use the current working directory as a fallback
	if currentBaseURL == nil {
		cwd, err := os.Getwd()
		if err != nil {
			ve = append(ve, newValidationErrorf("", "no baseURL set and failed to get working directory: %s", err))

			return asPublicCode(publiccode), ve
		}

		currentBaseURL = &url.URL{Scheme: "file", Path: cwd}
	}

	if err = validateFields(publiccode, p, !p.disableNetwork, currentBaseURL); err != nil {
		var vr ValidationResults
		if errors.As(err, &vr) {
			for _, result := range vr {
				var valErr ValidationError

				var valWarn ValidationWarning

				switch {
				case errors.As(result, &valErr):
					valErr.Line, valErr.Column = getPositionInFile(valErr.Key, file)
					ve = append(ve, valErr)
				case errors.As(result, &valWarn):
					valWarn.Line, valWarn.Column = getPositionInFile(valWarn.Key, file)
					ve = append(ve, valWarn)
				}
			}
		}
	}

	// v0: Copy data from deprecated fields to the canonical ones, where possible
	if v0, ok := publiccode.(*PublicCodeV0); ok {
		// Auto-copy the deprecated field into the new one, so we can guarantee
		// to the user that `IT` is always the field to check
		if v0.It != nil && v0.IT == nil {
			v0.IT = v0.It
		}

		it := v0.IT

		// Auto-copy the deprecated field into the new one, so we can guarantee
		// to the user that `organisation` is always the field to check
		if it != nil && it.Riuso.CodiceIPA != "" && v0.Organisation == nil {
			v0.Organisation = &OrganisationV0{}

			v0.Organisation.URI = "urn:x-italian-pa:" + it.Riuso.CodiceIPA
			v0.Organisation.Name = v0.Legal.RepoOwner
		}
	}

	if len(ve) == 0 {
		return asPublicCode(publiccode), nil
	}

	return asPublicCode(publiccode), ve
}

// Ensure the returned value implements PublicCode as a struct, not as a pointer.
func asPublicCode(pc PublicCode) PublicCode {
	switch v := pc.(type) {
	case *PublicCodeV0:
		return *v
	default:
		return v
	}
}

// getPositionInFile returns the line and column of the key in the YAML AST.
// Uses dot notation (e.g. "organisation.name") with optional array
// indices (e.g. "localisation.availableLanguages[0]").
func getPositionInFile(key string, file *ast.File) (int, int) {
	if len(file.Docs) == 0 || file.Docs[0].Body == nil {
		return 0, 0
	}

	parts := splitKeyParts(key)
	if len(parts) == 0 {
		return 0, 0
	}

	line, col := findKeyPos(file.Docs[0].Body, parts)

	return line, col
}

// splitKeyParts splits a dot-separated key path into parts, preserving array
// indices. Eg. "a.b[0].c" -> ["a", "b[0]", "c"].
func splitKeyParts(key string) []string {
	return strings.Split(key, ".")
}

// findKeyPos traverses the AST and returns the position of the key node
// identified by parts. Returns 0,0 if not found.
func findKeyPos(node ast.Node, parts []string) (int, int) {
	if node == nil || len(parts) == 0 {
		return 0, 0
	}

	part := parts[0]

	arrayIdx := -1
	stringKey := ""
	basePart := part

	// Parse array index "foo[0]" or string map key "foo[en-US]".
	if idx := strings.LastIndex(part, "["); idx >= 0 && strings.HasSuffix(part, "]") {
		idxStr := part[idx+1 : len(part)-1]
		n := 0
		valid := len(idxStr) > 0

		for _, c := range idxStr {
			if c < '0' || c > '9' {
				valid = false

				break
			}

			n = n*10 + int(c-'0')
		}

		if valid {
			arrayIdx = n
			basePart = part[:idx]
		} else if len(idxStr) > 0 {
			stringKey = idxStr
			basePart = part[:idx]
		}
	}

	switch n := node.(type) {
	case *ast.MappingNode:
		for _, mv := range n.Values {
			tok := mv.Key.GetToken()
			if tok == nil {
				continue
			}

			keyVal := tok.Value

			if keyVal != basePart {
				continue
			}

			if arrayIdx >= 0 {
				// Navigate into the sequence at this index.
				seq, ok := mv.Value.(*ast.SequenceNode)
				if !ok || arrayIdx >= len(seq.Values) {
					// Value is not a sequence or index out of bounds:
					// fall back to the key's own position.
					if tok := mv.Key.GetToken(); tok != nil {
						return tok.Position.Line, tok.Position.Column
					}

					return 0, 0
				}

				elem := seq.Values[arrayIdx]

				if len(parts) == 1 {
					tok := elem.GetToken()
					if tok != nil {
						return tok.Position.Line, tok.Position.Column
					}

					return 0, 0
				}

				return findKeyPos(elem, parts[1:])
			}

			if stringKey != "" {
				// Non-numeric bracket key (eg. "description[en-US]"): navigate
				// into the value treating the bracket content as the next key.
				return findKeyPos(mv.Value, append([]string{stringKey}, parts[1:]...))
			}

			if len(parts) == 1 {
				tok := mv.Key.GetToken()

				return tok.Position.Line, tok.Position.Column
			}

			return findKeyPos(mv.Value, parts[1:])
		}

		// Key not found.
		// For flow mappings ({...}) return the '{' token (e.g. "legal: {}" when
		// "legal.license" is missing).
		if n.IsFlowStyle {
			if tok := n.GetToken(); tok != nil {
				return tok.Position.Line, tok.Position.Column
			}
		}
		// For block mappings, only fall back to the first child key when this is
		// the final key segment. Mid path misses (e.g. "legal" not found while
		// looking for "legal.license") return 0,0 because we have no useful anchor.
		if len(parts) == 1 && len(n.Values) > 0 {
			if tok := n.Values[0].Key.GetToken(); tok != nil {
				return tok.Position.Line, tok.Position.Column
			}
		}
	case *ast.DocumentNode:
		return findKeyPos(n.Body, parts)
	default:
		// Node is a scalar or unexpected type
		// Return the node's own token so the position points to the
		// mistyped value.
		if tok := node.GetToken(); tok != nil {
			return tok.Position.Line, tok.Position.Column
		}
	}

	return 0, 0
}

// findKeyAtLine traverses the AST and returns the dot-separated YAML
// key path for the node whose key token is at the given line.
func findKeyAtLine(node ast.Node, targetLine int, prefix string) string {
	if node == nil {
		return ""
	}

	switch n := node.(type) {
	case *ast.MappingNode:
		for _, mv := range n.Values {
			if result := findKeyAtLine(mv, targetLine, prefix); result != "" {
				return result
			}
		}
	case *ast.MappingValueNode:
		keyTok := n.Key.GetToken()
		if keyTok == nil {
			return ""
		}

		var fullKey string

		if prefix == "" {
			fullKey = keyTok.Value
		} else {
			fullKey = prefix + "." + keyTok.Value
		}

		if keyTok.Position == nil || keyTok.Position.Line == targetLine {
			return fullKey
		}

		if result := findKeyAtLine(n.Value, targetLine, fullKey); result != "" {
			return result
		}
	case *ast.SequenceNode:
		for i, entry := range n.Values {
			seqKey := fmt.Sprintf("%s[%d]", prefix, i)

			if tok := entry.GetToken(); tok != nil && tok.Position != nil && tok.Position.Line == targetLine {
				return seqKey
			}

			if result := findKeyAtLine(entry, targetLine, seqKey); result != "" {
				return result
			}
		}
	case *ast.DocumentNode:
		return findKeyAtLine(n.Body, targetLine, prefix)
	}

	return ""
}

// decode decodes the YAML bytes into publiccode and returns any decode errors
// (type mismatches, unknown fields, syntax errors) as ValidationResults.
func decode[T any](data []byte, publiccode *T, file *ast.File) ValidationResults {
	var ve ValidationResults

	d := yaml.NewDecoder(bytes.NewReader(data), yaml.DisallowUnknownField())

	if err := d.Decode(publiccode); err != nil {
		var (
			unknownErr *yaml.UnknownFieldError
			yamlErr    yaml.Error
		)

		switch {
		case errors.As(err, &unknownErr):
			line := 0
			if unknownErr.Token != nil {
				line = unknownErr.Token.Position.Line
			}

			key := findKeyAtLine(file.Docs[0].Body, line, "")
			ve = append(ve, ValidationError{
				Key:         key,
				Description: unknownErr.Message,
				Line:        line,
				Column:      1,
			})
		case errors.As(err, &yamlErr):
			line := 0
			if tok := yamlErr.GetToken(); tok != nil {
				line = tok.Position.Line
			}

			key := findKeyAtLine(file.Docs[0].Body, line, "")
			ve = append(ve, ValidationError{
				Key:         key,
				Description: "wrong type for this field",
				Line:        line,
				Column:      1,
			})
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
		return nil, fmt.Errorf("getting absolute path for %q: %w", file, err)
	}
}
