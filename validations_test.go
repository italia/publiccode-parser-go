package publiccode

import (
	"image"
	"image/png"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
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

	p, _ := NewParser(ParserConfig{})
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
	ok, err := p.isImageFile(u, false)
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
	ok, err := p.isImageFile(u, false)
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
	ok, err := p.validLogo(u, false)
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
	ok, err := p.validLogo(u, false)
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
	// network=false: fileExists returns true for non-file scheme, and validLogo skips download
	ok, err := p.validLogo(*parsed, false)
	if !ok {
		t.Errorf("expected true for remote SVG with network=false (no download): %v", err)
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

	p, _ := NewParser(ParserConfig{})
	parsed, _ := url.Parse(srv.URL + "/logo.png")
	ok, err := p.validLogo(*parsed, true)
	// The downloaded file is not a valid PNG, so DecodeConfig should fail.
	if ok {
		t.Error("expected false for invalid PNG content")
	}
	if err == nil {
		t.Error("expected error for invalid PNG content")
	}
}

func TestValidLogoPNGTooSmall(t *testing.T) {
	// Create a tiny valid PNG (1x1 pixel) which is smaller than minLogoWidth.
	f, err := os.CreateTemp("", "tiny-*.png")
	if err != nil {
		t.Fatalf("can't create temp file: %v", err)
	}
	defer os.Remove(f.Name())

	// Write a minimal 1x1 PNG.
	png.Encode(f, image.NewRGBA(image.Rect(0, 0, 1, 1)))
	f.Close()

	p, _ := NewParser(ParserConfig{DisableNetwork: true})
	u := url.URL{Scheme: "file", Path: f.Name()}
	ok, err := p.validLogo(u, false)
	if ok {
		t.Error("expected false for tiny PNG")
	}
	if err == nil {
		t.Error("expected error for too-small image")
	}
	if !strings.Contains(err.Error(), "invalid image size") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestIsReachableSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	p, _ := NewParser(ParserConfig{})
	parsed, _ := url.Parse(srv.URL + "/path")
	reachable, err := p.isReachable(*parsed)
	if !reachable {
		t.Errorf("expected reachable for 200 response, got err: %v", err)
	}
}
