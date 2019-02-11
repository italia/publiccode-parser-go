package publiccode

import (
	"bytes"
	"net/url"

	"github.com/dyatlov/go-oembed/oembed"
)

// checkOembed tells whether the link is hosted on a valid oembed provider.
// Reference: https://oembed.com/providers.json
func (p *Parser) checkOembed(key string, link string) (*url.URL, error) {
	u, err := url.Parse(link)
	if err != nil {
		return nil, err
	}

	// Load oembed library and providers.js on from base64 variable
	providers, err := Asset("data/oembed_providers.json")
	if err != nil {
		return nil, newErrorInvalidValue(key, "error reading oembed providers list")
	}
	oe := oembed.NewOembed()
	oe.ParseProviders(bytes.NewReader(providers))

	item := oe.FindItem(link)
	if item == nil {
		return nil, newErrorInvalidValue(key, "invalid oembed link: %s", link)
	}

	info, err := item.FetchOembed(oembed.Options{URL: link})
	if err != nil {
		return nil, newErrorInvalidValue(key, "invalid oembed link: %s", err)
	}
	if info.Status >= 300 {
		return nil, newErrorInvalidValue(key, "invalid oembed link Status: %d", info.Status)
	}

	// save the Oembed HTML
	p.Oembed[link] = info.HTML

	return u, nil
}
