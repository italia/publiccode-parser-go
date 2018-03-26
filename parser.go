package publiccode

import (
	"bytes"
	"fmt"
	"net/url"
	"os"
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
		case map[interface{}]interface{}:
			if errs := p.decoderec(k, v); len(errs) > 0 {
				es = append(es, errs...)
			}
		case string:
			if err := p.decodeString(k, v); err != nil {
				es = append(es, err)
			}
		case []interface{}:
			var sl []string
			for idx, v1 := range v {
				if s, ok := v1.(string); ok {
					sl = append(sl, s)
				} else {
					es = append(es, newErrorInvalidValue(k, "array element %d not a string", idx))
				}
			}
			if err := p.decodeArrString(k, sl); err != nil {
				es = append(es, err)
			}
		default:
			panic(fmt.Errorf("invalid type: %T", v))
		}
	}
	return
}

func (p *parser) checkFile(key string, fn string) error {
	if _, err := os.Stat(fn); err != nil {
		return newErrorInvalidValue(key, "file does not exist: %v", fn)
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
	return u, nil
}

func (p *parser) checkDate(key string, value string) (time.Time, error) {
	if t, err := time.Parse("2006-01-02", value); err != nil {
		return t, newErrorInvalidValue(key, "cannot parse date: %v", err)
	} else {
		return t, nil
	}
}
