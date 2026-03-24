package publiccode

import (
	"net/url"
	"testing"
)

func TestPublicCodeV0Version(t *testing.T) {
	p := PublicCodeV0{}
	if p.Version() != 0 {
		t.Errorf("expected 0, got %d", p.Version())
	}
}

func TestPublicCodeV0UrlNil(t *testing.T) {
	p := PublicCodeV0{URL: nil}
	if p.Url() != nil {
		t.Error("expected nil")
	}
}

func TestPublicCodeV0UrlValid(t *testing.T) {
	raw, _ := url.Parse("https://github.com/example/repo")
	u := (*URL)(raw)
	p := PublicCodeV0{URL: u}
	if p.Url() == nil {
		t.Error("expected non-nil URL")
	}
}

func TestPublicCodeV0UrlInvalid(t *testing.T) {
	// A URL with no host/scheme won't pass IsValidURL.
	raw := &url.URL{Path: "/just/a/path"}
	u := (*URL)(raw)
	p := PublicCodeV0{URL: u}
	if p.Url() != nil {
		t.Error("expected nil for invalid URL")
	}
}

func TestPublicCodeV0ToYAMLClearsIt(t *testing.T) {
	raw, _ := url.Parse("https://github.com/example/repo.git")
	u := (*URL)(raw)
	releaseDate := "2024-01-01"
	locReady := true
	p := PublicCodeV0{
		PubliccodeYamlVersion: "0",
		Name:                  "test",
		URL:                   u,
		ReleaseDate:           &releaseDate,
		Platforms:             []string{"web"},
		DevelopmentStatus:     "stable",
		SoftwareType:          "standalone/web",
		Description: map[string]DescV0{
			"en": {
				ShortDescription: "short",
				LongDescription:  "long description that is at least 150 chars: xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
				GenericName:      "Generic",
			},
		},
		Legal: struct {
			License            string  `json:"license"                      validate:"required,is_spdx_expression" yaml:"license"`
			MainCopyrightOwner *string `json:"mainCopyrightOwner,omitempty" yaml:"mainCopyrightOwner,omitempty"`
			RepoOwner          *string `json:"repoOwner,omitempty"          yaml:"repoOwner,omitempty"`
			AuthorsFile        *string `json:"authorsFile,omitempty"        yaml:"authorsFile,omitempty"`
		}{License: "MIT"},
		Maintenance: struct {
			Type        string          `json:"type"                  validate:"required,oneof=internal contract community none"                        yaml:"type"`
			Contractors *[]ContractorV0 `json:"contractors,omitempty" validate:"required_if=Type contract,excluded_unless=Type contract,omitempty,dive" yaml:"contractors,omitempty"`
			Contacts    *[]ContactV0    `json:"contacts,omitempty"    validate:"required_if=Type community,required_if=Type internal,omitempty,dive"    yaml:"contacts,omitempty"`
		}{Type: "none"},
		Localisation: struct {
			LocalisationReady  *bool    `json:"localisationReady"  validate:"required"                                     yaml:"localisationReady"`
			AvailableLanguages []string `json:"availableLanguages" validate:"required,gt=0,dive,bcp47_strict_language_tag" yaml:"availableLanguages"`
		}{
			LocalisationReady:  &locReady,
			AvailableLanguages: []string{"en"},
		},
	}

	it := &ITSectionV0{}
	p.It = it

	b, err := p.ToYAML()
	if err != nil {
		t.Fatalf("ToYAML failed: %v", err)
	}
	if len(b) == 0 {
		t.Error("expected non-empty YAML output")
	}
}

func TestPublicCodeV1Version(t *testing.T) {
	p := PublicCodeV1{}
	if p.Version() != 1 {
		t.Errorf("expected 1, got %d", p.Version())
	}
}

func TestPublicCodeV1UrlNil(t *testing.T) {
	p := PublicCodeV1{URL: nil}
	if p.Url() != nil {
		t.Error("expected nil")
	}
}

func TestPublicCodeV1UrlValid(t *testing.T) {
	raw, _ := url.Parse("https://github.com/example/repo")
	u := (*URL)(raw)
	p := PublicCodeV1{URL: u}
	if p.Url() == nil {
		t.Error("expected non-nil URL")
	}
}

func TestPublicCodeV1UrlInvalid(t *testing.T) {
	raw := &url.URL{Path: "/just/a/path"}
	u := (*URL)(raw)
	p := PublicCodeV1{URL: u}
	if p.Url() != nil {
		t.Error("expected nil for invalid URL")
	}
}

func TestPublicCodeV1ToYAML(t *testing.T) {
	raw, _ := url.Parse("https://github.com/example/repo.git")
	u := (*URL)(raw)
	locReady := true
	p := PublicCodeV1{
		PubliccodeYamlVersion: "1",
		Name:                  "test",
		URL:                   u,
		Platforms:             []string{"web"},
		DevelopmentStatus:     "stable",
		SoftwareType:          "standalone/web",
		Description: map[string]DescV1{
			"en": {
				ShortDescription: "short",
				LongDescription:  "long description that is at least 150 chars: xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
			},
		},
		Legal: struct {
			License            string  `json:"license"                      validate:"required,is_spdx_expression" yaml:"license"`
			MainCopyrightOwner *string `json:"mainCopyrightOwner,omitempty" yaml:"mainCopyrightOwner,omitempty"`
		}{License: "MIT"},
		Maintenance: struct {
			Type        string          `json:"type"                  validate:"required,oneof=internal contract community none"                        yaml:"type"`
			Contractors *[]ContractorV1 `json:"contractors,omitempty" validate:"required_if=Type contract,excluded_unless=Type contract,omitempty,dive" yaml:"contractors,omitempty"`
			Contacts    *[]ContactV1    `json:"contacts,omitempty"    validate:"required_if=Type community,required_if=Type internal,omitempty,dive"    yaml:"contacts,omitempty"`
		}{Type: "none"},
		Localisation: struct {
			LocalisationReady  *bool    `json:"localisationReady"  validate:"required"                                     yaml:"localisationReady"`
			AvailableLanguages []string `json:"availableLanguages" validate:"required,gt=0,dive,bcp47_strict_language_tag" yaml:"availableLanguages"`
		}{
			LocalisationReady:  &locReady,
			AvailableLanguages: []string{"en"},
		},
	}

	b, err := p.ToYAML()
	if err != nil {
		t.Fatalf("ToYAML failed: %v", err)
	}
	if len(b) == 0 {
		t.Error("expected non-empty YAML output")
	}
}
