package publiccode

import (
	"bytes"
	"fmt"
	"net/url"

	"github.com/dyatlov/go-oembed/oembed"
)

// checkOembed tells whether the link is hosted on a valid oembed provider.
// Reference: https://oembed.com/providers.json
func (p *Parser) checkOembed(link string) (string, *url.URL, error) {
	u, err := url.Parse(link)
	if err != nil {
		return "", nil, err
	}

	if p.DisableNetwork {
		return link, u, nil
	}

	// Load oembed library and providers.js on from base64 variable
	providers, err := Asset("data/oembed_providers.json")
	if err != nil {
		panic("error reading oembed providers list")
	}
	oe := oembed.NewOembed()
	oe.ParseProviders(bytes.NewReader(providers))

	item := oe.FindItem(link)
	if item == nil {
		return "", nil, fmt.Errorf("invalid oembed link: %s", link)
	}

	info, err := item.FetchOembed(oembed.Options{URL: link})
	if err != nil {
		return "", nil, fmt.Errorf("invalid oembed link: %s", err)
	}
	if info.Status >= 300 {
		return "", nil, fmt.Errorf("invalid oembed link Status: %d", info.Status)
	}

	// save the Oembed HTML
	p.OEmbed[link] = info.HTML

	return link, u, nil
}
