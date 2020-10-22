package publiccode

import (
	"bytes"
	_ "fmt"
	"log"
	"io/ioutil"
	"net/http"
	"unicode/utf8"

	"gopkg.in/yaml.v3"
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

type T struct {
		A string
		B struct {
			RenamedC int   `yaml:"c"`
			D        []int `yaml:",flow"`
		}
	}

// Parse loads the yaml bytes and tries to parse it. Return an error if fails.
func (p *Parser) Parse(in []byte) error {
	// var s map[interface{}]interface{}
	// var s yaml.Node
	var x PublicCode
	// s := T{}

	if !utf8.Valid(in) {
		return ParseError{"Invalid UTF-8"}
	}

	d := yaml.NewDecoder(bytes.NewReader(in))
	d.KnownFields(true)
	if err := d.Decode(&x); err != nil {
		return err
	}
	// err := yaml.Unmarshal(in, &s)
	// log.Printf("%T", s)
	// log.Printf("____%s Line: %d\n%s Line content: %d", s, s.Line, s.Content[0], s.Content[1].Line)
	// if err != nil {
	// 	log.Fatalf("error: %T %s", err, err)
	// }
	// err = s.Decode(&x)
	// if err != nil {
	// 	log.Fatalf("error: %T %s", err, err)
	// }
	log.Printf("%s", x.ReleaseDateString)

	 // log.Printf("Length: %d", len(s.Content[0].Content))
	 // log.Printf("Tag: %s", s.Content[0].Content[0].Tag)
	 // log.Printf("Value: %s", s.Content[0].Content[0].Value)
	 // log.Printf("s", s)
	 // log.Printf("Line: %d", s.Line)
	 // log.Printf("Column: %d", s.Column)
	 // var x interface{};
	 // log.Printf("%s", s.Decode(x))
	// if err := p.validate(); err != nil {
	// 	return err
	// }
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

// ToYAML converts parser.PublicCode into YAML again.
func (p *Parser) ToYAML() ([]byte, error) {
	// Make a copy and set the latest versions
	pc2 := p.PublicCode
	pc2.PubliccodeYamlVersion = Version
	pc2.It.CountryExtensionVersion = ExtensionITVersion
	return yaml.Marshal(pc2)
}
