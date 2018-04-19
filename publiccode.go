package publiccode

import (
	"net/url"
	"time"
)

const Version = "0.0.1"

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
		Type              string
		Until             time.Time
		Maintainer        []string
		TechnicalContacts []Contact
	}

	Description struct {
		Name        string
		Logo        []string
		Shortdesc   []Desc
		LongDesc    []Desc
		Screenshots []string
		Videos      []*url.URL
		Version     string
		Released    time.Time
		Platforms   string
	}
	Meta struct {
		Scope    []string
		PaType   []string
		Category string
		Tags     []string
		UsedBy   []string
	}
	Dependencies struct {
		Hardware    []string
		Open        []string
		Proprietary []string
	}
}

type Contact struct {
	Name        string
	Email       string
	Affiliation string
}

type Desc struct {
	En string
	It string
}
