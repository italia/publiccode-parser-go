package publiccode

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"gopkg.in/yaml.v2"
)

func Parse(in []byte, pc *PublicCode) error {
	var s map[interface{}]interface{}

	d := yaml.NewDecoder(bytes.NewReader(in))
	if err := d.Decode(&s); err != nil {
		return err
	}

	return newParser(pc).parse(s)
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
				panic(fmt.Errorf("key \"%s\" is empty. Remove it or fill with valid values.", k))
			}
			panic(fmt.Errorf("key \"%s\" - invalid type: %T", k, v))
		}
	}
	return
}

func (p *parser) checkEmail(key string, fn string) error {
	re := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	if !re.MatchString(fn) {
		return newErrorInvalidValue(key, "invalid email: %v", fn)
	}

	return nil
}

func (p *parser) checkUrl(key string, value string) (*url.URL, error) {
	u, err := url.Parse(value)
	if err != nil {
		return nil, newErrorInvalidValue(key, "not a valid URL: %s", value)
	}
	if u.Scheme == "" {
		return nil, newErrorInvalidValue(key, "missing URL scheme: %s", value)
	}
	if r, err := http.Get(value); err != nil || r.StatusCode != 200 {
		return nil, newErrorInvalidValue(key, "URL is unreachable: %s", value)
	}

	return u, nil
}

func (p *parser) checkFile(key string, fn string) (string, error) {
	if _, err := os.Stat(fn); err != nil {
		return "", newErrorInvalidValue(key, "file does not exist: %v", fn)
	}
	return fn, nil
}

func (p *parser) checkDate(key string, value string) (time.Time, error) {
	if t, err := time.Parse("2006-01-02", value); err != nil {
		return t, newErrorInvalidValue(key, "cannot parse date: %v", err)
	} else {
		return t, nil
	}
}

func (p *parser) checkImage(key string, value string) error {
	// Based on https://github.com/italia/publiccode.yml/blob/master/schema.md#key-descriptionlogo
	//TODO: check two version of the Logo
	//TODO: check .png size
	//TODO: raster should exists iff vector does not exists
	if _, err := p.checkFile(key, value); err != nil {
		return err
	}
	fileExt := filepath.Ext(value)
	for _, v := range []string{".png", ".svg"} {
		if v == fileExt {
			p.pc.Maintenance.Type = value
			return nil
		}
	}
	return newErrorInvalidValue(key, "image must be .svg or .png: %v", value)

}
