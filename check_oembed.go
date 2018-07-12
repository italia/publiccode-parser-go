package publiccode

import (
	"bytes"
	"io/ioutil"
	"net/url"

	"github.com/dyatlov/go-oembed/oembed"
)

// checkOembed tells whether the link is hosted on a valid oembed provider.
// Reference: https://oembed.com/providers.json
func (p *parser) checkOembed(key string, link *url.URL) (*url.URL, error) {
	if link.String() == "" {
		return link, newErrorInvalidValue(key, "empty oembed link")
	}

	// Load oembed library and providers.js on from base64 variable
	oe := oembed.NewOembed()
	file := "./data/oembed_providers.json"
	dataFile, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, newErrorInvalidValue(key, "error reading oembed providers list")
	}
	providers := dataFile
	oe.ParseProviders(bytes.NewReader(providers))

	item := oe.FindItem(link.String())

	if item != nil {
		info, err := item.FetchOembed(oembed.Options{URL: link.String()})
		if err != nil {
			return link, newErrorInvalidValue(key, "invalid oembed link: %s", err)
		} else {
			if info.Status >= 300 {
				return link, newErrorInvalidValue(key, "invalid oembed link Status: %d", info.Status)
			} else {
				return link, nil
			}
		}
	}

	return link, newErrorInvalidValue(key, "invalid oembed link: %s", link)
}
