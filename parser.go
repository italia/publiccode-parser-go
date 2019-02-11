package publiccode

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"

	"gopkg.in/yaml.v2"
)

// LocalBasePath is a filesystem path pointing to the directory where the
// publiccode.yml is located. It's used as a base for relative paths. If
// left empty, RemoteBaseURL will be used.
var LocalBasePath = ""

// RemoteBaseURL is the URL pointing to the directory where the publiccode.yml
// file is located. It's used for validating abolute URLs and as a base for
// relative paths. If left empty, absolute URLs will not be validated and
// no remote validation of files with relative paths will be performed. If
// not left empty, publiccode.yml keys with relative paths will be turned
// into absolute URLs.
// (eg: https://raw.githubusercontent.com/gith002/Medusa/master)
var RemoteBaseURL = ""

// Lock uses sync.Mutex lock/unlock for goroutines.
var Lock sync.Mutex

// Parse loads the yaml bytes and tries to parse it. Return an error if fails.
func Parse(in []byte) (*PublicCode, error) {
	var s map[interface{}]interface{}
	// Lock for goroutines.
	Lock.Lock()

	d := yaml.NewDecoder(bytes.NewReader(in))
	if err := d.Decode(&s); err != nil {
		return nil, err
	}
	// Unlock for goroutines.
	Lock.Unlock()

	var pc PublicCode
	err := newParser(&pc).parse(s)
	return &pc, err
}

// ParseFile loads a publiccode.yml file from a given file path.
func ParseFile(file string) (*PublicCode, error) {
	// Read data.
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	return Parse(data)
}

// ParseRemoteFile loads a publiccode.yml file from its raw URL.
func ParseRemoteFile(url string) (*PublicCode, error) {
	// Read data.
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return Parse(data)
}

type parser struct {
	pc      *PublicCode
	missing map[string]bool
}

func newParser(pc *PublicCode) *parser {
	var p parser
	p.pc = pc
	p.missing = make(map[string]bool)
	for _, k := range mandatoryKeys {
		p.missing[k] = true
	}
	return &p
}

func (p *parser) parse(s map[interface{}]interface{}) error {
	if err := p.decoderec("", s); err != nil {
		return err
	}
	if err := p.finalize(); err != nil {
		return err
	}
	return nil
}

func (p *parser) decoderec(prefix string, s map[interface{}]interface{}) (es ErrorParseMulti) {

	for ki, v := range s {
		k, ok := ki.(string)
		if !ok {
			es = append(es, ErrorInvalidKey{Key: fmt.Sprint(ki)})
			continue
		}
		if prefix != "" {
			k = prefix + "/" + k
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
			if v == nil {
				panic(fmt.Errorf("key \"%s\" is empty. Remove it or fill with valid values", k))
			}
			panic(fmt.Errorf("key \"%s\" - invalid type: %T", k, v))
		}
	}
	return
}
