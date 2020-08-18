package publiccode

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"

	funk "github.com/thoas/go-funk"
	"gopkg.in/yaml.v2"
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

	// Strict makes the parser less tolerant by allowing some errors that do not
	// affect the rendering of the software catalog. It is enabled by default.
	Strict bool

	OEmbed  map[string]string
	missing map[string]bool

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

// Parse loads the yaml bytes and tries to parse it. Return an error if fails.
func (p *Parser) Parse(in []byte) error {
	var s map[interface{}]interface{}

	d := yaml.NewDecoder(bytes.NewReader(in))
	if err := d.Decode(&s); err != nil {
		return err
	}

	if err := p.decoderec("", s); err != nil {
		return err
	}
	if err := p.finalize(); err != nil {
		return err
	}
	return nil
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
	p.Strict = true
	p.OEmbed = make(map[string]string)
	p.missing = make(map[string]bool)
	for _, k := range mandatoryKeys {
		p.missing[k] = true
	}
	return &p
}

func (p *Parser) decoderec(prefix string, s map[interface{}]interface{}) (es ErrorParseMulti) {

	for ki, v := range s {
		k, ok := ki.(string)
		if !ok {
			es = append(es, ErrorInvalidKey{Key: fmt.Sprint(ki)})
			continue
		}

		if prefix != "" {
			k = prefix + "/" + k
		}

		// if we are not running in strict mode, support legacy keys
		if !p.Strict {
			if k2, ok := renamedKeys[k]; ok {
				k = k2
			}
			if funk.Contains(removedKeys, k) {
				continue // ignore key
			}
		}

		delete(p.missing, k)

		switch v := v.(type) {
		case string:
			if err := p.decodeString(k, v); err != nil {
				es = append(es, err)
			}
		case bool:
			if err := p.decodeBool(k, v); err != nil {
				es = append(es, err)
			}
		case []interface{}:
			sl := []string{}
			sli := make(map[interface{}]interface{})

			for idx, v1 := range v {
				// if array of strings
				if s, ok := v1.(string); ok {
					sl = append(sl, s)
					if len(sl) == len(v) { //the v1.(string) check should be extracted.
						if err := p.decodeArrString(k, sl); err != nil {
							es = append(es, err)
						}
					}
					// if array of objects
				} else if _, ok := v1.(map[interface{}]interface{}); ok {
					sli[k] = v1
					if err := p.decodeArrObj(k, sli); err != nil {
						es = append(es, err)
					}

				} else {
					es = append(es, newErrorInvalidValue(k, "array element %d not a string", idx))
				}
			}

		case map[interface{}]interface{}:
			if errs := p.decoderec(k, v); len(errs) > 0 {
				es = append(es, errs...)
			}
		default:
			es = append(es, newErrorInvalidValue(k, "invalid type %T.", v))
		}
	}
	return
}

// ToYAML converts parser.PublicCode into YAML again.
func (p *Parser) ToYAML() ([]byte, error) {
	// Make a copy and set the latest versions
	pc2 := p.PublicCode
	pc2.PubliccodeYamlVersion = Version
	pc2.It.CountryExtensionVersion = ExtensionITVersion
	return yaml.Marshal(pc2)
}
