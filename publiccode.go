package publiccode

import (
	"net/url"
	"time"
)

const Version = "0.1"

type PublicCode struct {
	Version     string
	Url         *url.URL
	UpstreamUrl []*url.URL

	Legal struct {
		License            string
		MainCopyrightOwner string
		AuthorsFile        string
		RepoOwner          string
	}

	Maintenance struct {
		Type       string
		Until      time.Time
		Maintainer []string
	}
}
