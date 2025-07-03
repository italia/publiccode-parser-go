package publiccode

import (
	urlutil "github.com/italia/publiccode-parser-go/v4/internal"
	"gopkg.in/yaml.v3"
)

// There's no v1 yet, this is just an unexported placeholder type

//nolint:unused
type publicCodeV1 struct {
	URL *URL `validate:"required,url_url" yaml:"url"`
}

//nolint:unused
func (p publicCodeV1) Version() uint {
	return 1
}

//nolint:unused
func (p publicCodeV1) ToYAML() ([]byte, error) {
	return yaml.Marshal(p)
}

//nolint:unused
func (p publicCodeV1) Url() *URL {
	if p.URL == nil {
		return nil
	}

	if ok, _ := urlutil.IsValidURL(p.URL.String()); ok {
		return p.URL
	}

	return nil
}
