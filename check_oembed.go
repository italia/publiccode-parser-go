package publiccode

import (
	"bytes"
	"fmt"
	"net/url"

	"github.com/dyatlov/go-oembed/oembed"
	"github.com/italia/publiccode-parser-go/v3/data"
)

// checkOembed tells whether the link is hosted on a valid oembed provider.
// Reference: https://oembed.com/providers.json
func (p *Parser) isOembedURL(url *url.URL) (bool, error) {
	if p.DisableNetwork {
		return true, nil
	}

	// Load oembed library and providers.js on from base64 variable
	b := data.OembedProviders
	oe := oembed.NewOembed()
	_ = oe.ParseProviders(bytes.NewReader(b))

	link := url.String()
	item := oe.FindItem(link)
	if item == nil {
		return false, fmt.Errorf("invalid oembed link: %s", link)
	}

	info, err := item.FetchOembed(oembed.Options{URL: link})
	if err != nil {
		return false, fmt.Errorf("invalid oembed link: %s", err)
	}
	if info.Status >= 300 {
		return false, fmt.Errorf("invalid oembed link Status: %d", info.Status)
	}

	return true, nil
}
