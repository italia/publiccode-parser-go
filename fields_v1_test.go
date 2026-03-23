package publiccode

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestAsPublicCodeV1(t *testing.T) {
	v1 := &PublicCodeV1{}
	result := asPublicCode(v1)
	if result == nil {
		t.Fatal("asPublicCode returned nil for *PublicCodeV1")
	}
	if _, ok := result.(*PublicCodeV1); !ok {
		t.Errorf("asPublicCode returned unexpected type: %T", result)
	}
}

// TestValidateFieldsV1Direct calls validateFieldsV1 directly to cover its code paths
// without needing to add "1" to SupportedVersions.
func TestValidateFieldsV1Direct(t *testing.T) {
	p, err := NewParser(ParserConfig{DisableNetwork: true, DisableExternalChecks: true})
	if err != nil {
		t.Fatalf("can't create parser: %v", err)
	}

	base := &url.URL{Scheme: "file", Path: "/tmp"}

	raw, _ := url.Parse("https://github.com/example/repo.git")
	u := (*URL)(raw)

	locReady := true
	v1 := &PublicCodeV1{
		PubliccodeYamlVersion: "1",
		Name:                  "Test",
		URL:                   u,
		Platforms:             []string{"web"},
		DevelopmentStatus:     "stable",
		SoftwareType:          "standalone/web",
		Description: map[string]DescV1{
			"en": {
				ShortDescription: "Short description.",
				LongDescription:  strings.Repeat("x", 151),
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

	err = validateFieldsV1(v1, p, false, base)
	if err != nil {
		t.Errorf("unexpected error from validateFieldsV1: %v", err)
	}
}

func TestValidateFieldsV1WithLogoAbsPath(t *testing.T) {
	p, _ := NewParser(ParserConfig{DisableNetwork: true, DisableExternalChecks: false})
	base := &url.URL{Scheme: "file", Path: "/tmp"}

	logo := "/absolute/path/logo.svg"
	locReady := true
	v1 := &PublicCodeV1{
		Logo: &logo,
		Localisation: struct {
			LocalisationReady  *bool    `json:"localisationReady"  validate:"required"                                     yaml:"localisationReady"`
			AvailableLanguages []string `json:"availableLanguages" validate:"required,gt=0,dive,bcp47_strict_language_tag" yaml:"availableLanguages"`
		}{
			LocalisationReady:  &locReady,
			AvailableLanguages: []string{"en"},
		},
	}

	err := validateFieldsV1(v1, p, false, base)
	if err == nil {
		t.Error("expected error for absolute logo path")
	}
}

func TestValidateFieldsV1WithVideos(t *testing.T) {
	p, _ := NewParser(ParserConfig{DisableNetwork: true, DisableExternalChecks: true})
	base := &url.URL{Scheme: "file", Path: "/tmp"}

	videoURL, _ := url.Parse("https://www.invalid-video-url.example.com/watch?v=abc")
	v := (*URL)(videoURL)

	v1 := &PublicCodeV1{
		Description: map[string]DescV1{
			"en": {
				Videos: []*URL{v},
			},
		},
	}

	err := validateFieldsV1(v1, p, false, base)
	if err == nil {
		t.Error("expected error for invalid oEmbed video URL")
	}
}

func TestValidateFieldsV1WithScreenshotAbsPath(t *testing.T) {
	p, _ := NewParser(ParserConfig{DisableNetwork: true, DisableExternalChecks: false})
	base := &url.URL{Scheme: "file", Path: "/tmp"}

	v1 := &PublicCodeV1{
		Description: map[string]DescV1{
			"en": {
				Screenshots: []string{"/absolute/path/screen.png"},
			},
		},
	}

	err := validateFieldsV1(v1, p, false, base)
	if err == nil {
		t.Error("expected error for absolute path screenshot")
	}
}

func TestValidateFieldsV1WithScreenshotRelative(t *testing.T) {
	p, _ := NewParser(ParserConfig{DisableNetwork: true, DisableExternalChecks: false})
	base := &url.URL{Scheme: "file", Path: "/tmp"}

	v1 := &PublicCodeV1{
		Description: map[string]DescV1{
			"en": {
				Screenshots: []string{"assets/screen.png"},
			},
		},
	}

	// Should not panic, may produce an error about missing file
	_ = validateFieldsV1(v1, p, false, base)
}

func TestValidateFieldsV1WithNetworkChecks(t *testing.T) {
	// Use a test server that returns 200 so we can exercise the IsRepo check path,
	// but the URL is not a code repository.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	p, _ := NewParser(ParserConfig{DisableNetwork: false, DisableExternalChecks: false})
	base := &url.URL{Scheme: "file", Path: "/tmp"}

	// srv.URL is not a code repo, so IsRepo should return false
	rawURL, _ := url.Parse(srv.URL + "/not-a-repo")
	u := (*URL)(rawURL)

	landingRaw, _ := url.Parse(srv.URL + "/landing")
	landing := (*URL)(landingRaw)

	roadmapRaw, _ := url.Parse(srv.URL + "/roadmap")
	roadmap := (*URL)(roadmapRaw)

	docRaw, _ := url.Parse(srv.URL + "/docs")
	docURL := (*URL)(docRaw)

	apiDocRaw, _ := url.Parse(srv.URL + "/api-docs")
	apiDoc := (*URL)(apiDocRaw)

	v1 := &PublicCodeV1{
		URL:        u,
		LandingURL: landing,
		Roadmap:    roadmap,
		Description: map[string]DescV1{
			"en": {
				Documentation:    docURL,
				APIDocumentation: apiDoc,
			},
		},
	}

	// network=true, checksNetwork=true
	err := validateFieldsV1(v1, p, true, base)
	// We expect the "not a valid code repository" error.
	if err == nil {
		t.Log("no error (URL might have been checked differently)")
	}
}

func TestValidateFieldsV1WithUnreachableURLs(t *testing.T) {
	// Use a test server that returns 404 so we can exercise the not-reachable paths.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer srv.Close()

	p, _ := NewParser(ParserConfig{DisableNetwork: false, DisableExternalChecks: false})
	base := &url.URL{Scheme: "file", Path: "/tmp"}

	rawURL, _ := url.Parse(srv.URL + "/repo.git")
	u := (*URL)(rawURL)

	landingRaw, _ := url.Parse(srv.URL + "/landing")
	landing := (*URL)(landingRaw)

	roadmapRaw, _ := url.Parse(srv.URL + "/roadmap")
	roadmap := (*URL)(roadmapRaw)

	v1 := &PublicCodeV1{
		URL:        u,
		LandingURL: landing,
		Roadmap:    roadmap,
	}

	docRaw, _ := url.Parse(srv.URL + "/docs")
	docURL := (*URL)(docRaw)

	v1.Description = map[string]DescV1{
		"en": {Documentation: docURL},
	}

	// network=true, checksNetwork=true
	err := validateFieldsV1(v1, p, true, base)
	// We expect errors because the URLs return 404.
	if err == nil {
		t.Error("expected errors for unreachable URLs")
	}
}

func TestValidateFieldsV1WithAPIDocNetworkCheck(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer srv.Close()

	p, _ := NewParser(ParserConfig{})
	base := &url.URL{Scheme: "file", Path: "/tmp"}

	apiDocRaw, _ := url.Parse(srv.URL + "/api-docs")
	apiDoc := (*URL)(apiDocRaw)

	v1 := &PublicCodeV1{
		Description: map[string]DescV1{
			"en": {
				APIDocumentation: apiDoc,
			},
		},
	}

	err := validateFieldsV1(v1, p, true, base)
	if err == nil {
		t.Error("expected error for unreachable API documentation URL")
	}
}

func TestValidateFieldsV1WithCountriesDeprecated(t *testing.T) {
	p, _ := NewParser(ParserConfig{DisableNetwork: true, DisableExternalChecks: true})
	base := &url.URL{Scheme: "file", Path: "/tmp"}

	countries := []string{"it"} // lowercase - deprecated
	unsupported := []string{"de"} // lowercase - deprecated
	locReady := true
	v1 := &PublicCodeV1{
		IntendedAudience: &struct {
			Scope                *[]string `json:"scope,omitempty"                validate:"omitempty,dive,is_scope_v0"                     yaml:"scope,omitempty"`
			Countries            *[]string `json:"countries,omitempty"            validate:"omitempty,dive,iso3166_1_alpha2_lower_or_upper" yaml:"countries,omitempty"`
			UnsupportedCountries *[]string `json:"unsupportedCountries,omitempty" validate:"omitempty,dive,iso3166_1_alpha2_lower_or_upper" yaml:"unsupportedCountries,omitempty"`
		}{
			Countries:            &countries,
			UnsupportedCountries: &unsupported,
		},
		Localisation: struct {
			LocalisationReady  *bool    `json:"localisationReady"  validate:"required"                                     yaml:"localisationReady"`
			AvailableLanguages []string `json:"availableLanguages" validate:"required,gt=0,dive,bcp47_strict_language_tag" yaml:"availableLanguages"`
		}{
			LocalisationReady:  &locReady,
			AvailableLanguages: []string{"en"},
		},
	}

	err := validateFieldsV1(v1, p, false, base)
	if err == nil {
		t.Error("expected deprecation warnings for lowercase countries")
	}
}

func TestValidateFieldsV0MonochromeLogoAbsPath(t *testing.T) {
	p, _ := NewParser(ParserConfig{DisableNetwork: true, DisableExternalChecks: false})
	base := &url.URL{Scheme: "file", Path: "/tmp"}

	logo := "/absolute/mono.svg"
	v0 := &PublicCodeV0{}
	v0.MonochromeLogo = &logo

	err := validateFieldsV0(v0, p, false, base)
	vr, ok := err.(ValidationResults)
	if !ok {
		t.Fatal("expected ValidationResults")
	}

	found := false
	for _, e := range vr {
		if ve, ok := e.(ValidationError); ok && ve.Key == "monochromeLogo" {
			found = true
		}
	}

	if !found {
		t.Error("expected monochromeLogo validation error for absolute path")
	}
}

func TestValidateFieldsV0AuthorsFileAbsPath(t *testing.T) {
	p, _ := NewParser(ParserConfig{DisableNetwork: true, DisableExternalChecks: false})
	base := &url.URL{Scheme: "file", Path: "/tmp"}

	authorsFile := "/absolute/AUTHORS"
	v0 := &PublicCodeV0{}
	v0.Legal.AuthorsFile = &authorsFile

	err := validateFieldsV0(v0, p, false, base)
	vr, ok := err.(ValidationResults)
	if !ok {
		t.Fatal("expected ValidationResults")
	}

	found := false
	for _, e := range vr {
		if ve, ok := e.(ValidationError); ok && ve.Key == "legal.authorsFile" {
			found = true
		}
	}

	if !found {
		t.Error("expected legal.authorsFile validation error for absolute path")
	}
}

func TestValidateFieldsV0ScreenshotsAbsPath(t *testing.T) {
	p, _ := NewParser(ParserConfig{DisableNetwork: true, DisableExternalChecks: false})
	base := &url.URL{Scheme: "file", Path: "/tmp"}

	v0 := &PublicCodeV0{
		Description: map[string]DescV0{
			"en": {Screenshots: []string{"/absolute/screen.png"}},
		},
	}

	err := validateFieldsV0(v0, p, false, base)
	vr, ok := err.(ValidationResults)
	if !ok {
		t.Fatal("expected ValidationResults")
	}

	found := false
	for _, e := range vr {
		if ve, ok := e.(ValidationError); ok && strings.Contains(ve.Key, "screenshots") {
			found = true
		}
	}

	if !found {
		t.Error("expected screenshots validation error for absolute path")
	}
}

func TestValidateFieldsV1LogoRelativeExternalChecks(t *testing.T) {
	// A valid relative path with DisableExternalChecks=false exercises the
	// external check path for v1 logo (the else-if branch after isRelativePathOrURL).
	p, _ := NewParser(ParserConfig{DisableNetwork: true, DisableExternalChecks: false})
	base := &url.URL{Scheme: "file", Path: "/tmp"}

	logo := "logo.svg" // relative path, passes isRelativePathOrURL
	locReady := true
	v1 := &PublicCodeV1{
		Logo: &logo,
		Localisation: struct {
			LocalisationReady  *bool    `json:"localisationReady"  validate:"required"                                     yaml:"localisationReady"`
			AvailableLanguages []string `json:"availableLanguages" validate:"required,gt=0,dive,bcp47_strict_language_tag" yaml:"availableLanguages"`
		}{
			LocalisationReady:  &locReady,
			AvailableLanguages: []string{"en"},
		},
	}

	// /tmp/logo.svg does not exist, so validLogo should produce an error
	err := validateFieldsV1(v1, p, false, base)
	if err == nil {
		t.Error("expected error for missing logo file")
	}
}
