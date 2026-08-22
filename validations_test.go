package publiccode

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestGetBasicAuth(t *testing.T) {
	d := Domain{BasicAuth: []string{"user:pass"}}
	result := getBasicAuth(d)
	if !strings.HasPrefix(result, "Basic ") {
		t.Errorf("expected 'Basic ' prefix, got %q", result)
	}
}

func TestGetBasicAuthEmpty(t *testing.T) {
	d := Domain{}
	result := getBasicAuth(d)
	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}

func TestIsHostInDomainTrue(t *testing.T) {
	d := Domain{UseTokenFor: []string{"example.com"}}
	if !isHostInDomain(d, "https://example.com/foo") {
		t.Error("expected true")
	}
}

func TestIsHostInDomainFalse(t *testing.T) {
	d := Domain{UseTokenFor: []string{"example.com"}}
	if isHostInDomain(d, "https://other.com/foo") {
		t.Error("expected false")
	}
}

func TestIsHostInDomainEmptyUseTokenFor(t *testing.T) {
	d := Domain{}
	if isHostInDomain(d, "https://example.com") {
		t.Error("expected false for empty UseTokenFor")
	}
}

func TestIsHostInDomainInvalidURL(t *testing.T) {
	d := Domain{UseTokenFor: []string{"example.com"}}
	if isHostInDomain(d, "not-a-url") {
		t.Error("expected false for relative URL with no host")
	}
}

func TestIsHostInDomainMalformedURL(t *testing.T) {
	d := Domain{UseTokenFor: []string{"example.com"}}
	// "%" is a truly malformed URL that causes url.Parse to fail
	if isHostInDomain(d, "%") {
		t.Error("expected false for malformed URL")
	}
}

func TestGetHeaderFromDomainNoMatch(t *testing.T) {
	d := Domain{UseTokenFor: []string{"example.com"}}
	headers := getHeaderFromDomain(d, "https://other.com/foo")
	if headers != nil {
		t.Error("expected nil headers when host not in domain")
	}
}

func TestGetHeaderFromDomainMatch(t *testing.T) {
	d := Domain{UseTokenFor: []string{"example.com"}, BasicAuth: []string{"user:pass"}}
	headers := getHeaderFromDomain(d, "https://example.com/foo")
	if headers == nil {
		t.Fatal("expected non-nil headers")
	}
	if _, ok := headers["Authorization"]; !ok {
		t.Error("expected Authorization header")
	}
}

func TestIsReachableMissingScheme(t *testing.T) {
	p, _ := NewParser(ParserConfig{DisableNetwork: true})
	_, err := p.isReachable(url.URL{Scheme: "", Host: "example.com", Path: "/"})
	if err == nil {
		t.Fatal("expected error for missing scheme")
	}
	if !strings.Contains(err.Error(), "missing URL scheme") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestIsReachableNon200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer srv.Close()

	p, _ := NewParser(ParserConfig{AllowNetworkToPrivateHosts: true})
	parsed, _ := url.Parse(srv.URL + "/path")
	reachable, err := p.isReachable(*parsed)
	if reachable {
		t.Error("expected not reachable for 404 response")
	}
	if err == nil {
		t.Error("expected error for non-200 response")
	}
}

func TestIsImageFileInvalidExtension(t *testing.T) {
	p, _ := NewParser(ParserConfig{DisableNetwork: true})
	u := url.URL{Scheme: "file", Path: "/tmp/test.gif"}
	ok, err := p.isImageFile(u, CheckLocal)
	if ok {
		t.Error("expected false for .gif extension")
	}
	if err == nil {
		t.Error("expected error for invalid extension")
	}
}

func TestIsImageFileValidExtensionMissing(t *testing.T) {
	p, _ := NewParser(ParserConfig{DisableNetwork: true})
	u := url.URL{Scheme: "file", Path: "/nonexistent/test.png"}
	ok, err := p.isImageFile(u, CheckLocal)
	if ok {
		t.Error("expected false for nonexistent file")
	}
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestValidLogoInvalidExtension(t *testing.T) {
	p, _ := NewParser(ParserConfig{DisableNetwork: true})
	u := url.URL{Scheme: "file", Path: "/tmp/test.gif"}
	ok, err := p.validLogo(u, CheckLocal)
	if ok {
		t.Error("expected false for .gif extension")
	}
	if err == nil {
		t.Error("expected error for invalid extension")
	}
}

func TestValidLogoMissingFile(t *testing.T) {
	p, _ := NewParser(ParserConfig{DisableNetwork: true})
	u := url.URL{Scheme: "file", Path: "/nonexistent/logo.svg"}
	ok, err := p.validLogo(u, CheckLocal)
	if ok {
		t.Error("expected false for nonexistent file")
	}
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestValidLogoRemoteNoNetwork(t *testing.T) {
	// Remote URL but network disabled: validLogo returns true for svg without downloading.
	p, _ := NewParser(ParserConfig{DisableNetwork: true})

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	parsed, _ := url.Parse(srv.URL + "/logo.svg")
	// CheckLocal: fileExists returns true for non-file scheme, and validLogo skips download
	ok, err := p.validLogo(*parsed, CheckLocal)
	if !ok {
		t.Errorf("expected true for remote SVG with CheckLocal (no download): %v", err)
	}
}

func TestValidLogoRemoteDownloadFailure(t *testing.T) {
	// Remote PNG URL that is reachable (200) but returns non-PNG data.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "logo.png") {
			// Return 200 but with invalid PNG content
			w.Header().Set("Content-Type", "image/png")
			_, _ = w.Write([]byte("not a valid PNG file at all"))
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer srv.Close()

	p, _ := NewParser(ParserConfig{AllowNetworkToPrivateHosts: true})
	parsed, _ := url.Parse(srv.URL + "/logo.png")
	ok, err := p.validLogo(*parsed, CheckNetwork)
	// The downloaded file is not a valid PNG, so DecodeConfig should fail.
	if ok {
		t.Error("expected false for invalid PNG content")
	}
	if err == nil {
		t.Error("expected error for invalid PNG content")
	}
}

func TestCheckModeSelectsTheExternalChecks(t *testing.T) {
	const (
		missingLocalFile = "testdata/v0/invalid/no-network/logo_missing_file.yml"
		deadRemoteURL    = "testdata/v0/invalid/logo_missing_url.yml"
	)

	missingFileError := ValidationResults{
		ValidationError{"", "logo", "no such file: " + cwd + "/testdata/v0/invalid/no-network/no_such_file.png", 18, 1},
	}
	unreachableURLError := ValidationResults{
		ValidationError{"", "logo", "HTTP GET failed for https://google.com/no_such_file.png: not found", 18, 1},
	}

	tests := map[string]struct {
		file string
		mode CheckMode
		want error
	}{
		"missing local file, network": {missingLocalFile, CheckNetwork, missingFileError},
		"missing local file, local":   {missingLocalFile, CheckLocal, missingFileError},
		"missing local file, none":    {missingLocalFile, CheckNone, nil},
		"dead remote URL, network":    {deadRemoteURL, CheckNetwork, unreachableURLError},
		"dead remote URL, local":      {deadRemoteURL, CheckLocal, nil},
		"dead remote URL, none":       {deadRemoteURL, CheckNone, nil},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			p, err := NewParser(ParserConfig{ExternalChecks: test.mode})
			if err != nil {
				t.Fatalf("can't create parser: %v", err)
			}

			_, err = p.Parse(test.file)

			checkParseErrors(t, err, testType{test.file, test.want})
		})
	}
}

func TestCheckNoneKeepsTheOtherChecks(t *testing.T) {
	p, err := NewParser(ParserConfig{ExternalChecks: CheckNone})
	if err != nil {
		t.Fatalf("can't create parser: %v", err)
	}

	_, err = p.Parse("testdata/v0/invalid/categories_invalid.yml")

	checkParseErrors(t, err, testType{"testdata/v0/invalid/categories_invalid.yml", ValidationResults{
		ValidationError{
			"",
			"categories[0]",
			"categories[0] must be a valid category " +
				"(see https://github.com/publiccodeyml/publiccode.yml/blob/main/docs/standard/categories-list.rst)",
			13, 5,
		},
	}})
}

func TestIsReachableSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	p, _ := NewParser(ParserConfig{AllowNetworkToPrivateHosts: true})
	parsed, _ := url.Parse(srv.URL + "/path")
	reachable, err := p.isReachable(*parsed)
	if !reachable {
		t.Errorf("expected reachable for 200 response, got err: %v", err)
	}
}
