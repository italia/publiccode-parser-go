package publiccode

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

// errReader always returns an error when Read is called.
type errReader struct{}

func (e errReader) Read(_ []byte) (int, error) {
	return 0, errors.New("read error")
}

func TestNewParserBaseURL(t *testing.T) {
	p, err := NewParser(ParserConfig{BaseURL: "https://example.com"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.baseURL == nil {
		t.Error("expected baseURL to be set")
	}
	if p.baseURL.Host != "example.com" {
		t.Errorf("unexpected host: %s", p.baseURL.Host)
	}
}

func TestNewParserDisableExternalChecksImpliesDisableNetwork(t *testing.T) {
	p, err := NewParser(ParserConfig{DisableExternalChecks: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !p.disableNetwork {
		t.Error("expected disableNetwork to be true when DisableExternalChecks is true")
	}
}

func TestParseNonexistentFile(t *testing.T) {
	p, _ := NewParser(ParserConfig{DisableNetwork: true})
	_, err := p.Parse("/nonexistent/path/file.yml")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
	if !strings.Contains(err.Error(), "can't open file") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestParseHTTPError(t *testing.T) {
	p, _ := NewParser(ParserConfig{DisableNetwork: true})
	// Port 1 is not listening, so this should fail with a connection error.
	_, err := p.Parse("http://localhost:1/publiccode.yml")
	if err == nil {
		t.Error("expected error for unreachable HTTP server")
	}
	if !strings.Contains(err.Error(), "can't GET") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestParseStreamInvalidUTF8(t *testing.T) {
	p, _ := NewParser(ParserConfig{DisableNetwork: true})
	// Invalid UTF-8 bytes.
	invalid := []byte{0xff, 0xfe, 0x00}
	_, err := p.ParseStream(bytes.NewReader(invalid))
	if err == nil {
		t.Fatal("expected error for invalid UTF-8")
	}
	if !strings.Contains(err.Error(), "Invalid UTF-8") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestParseStreamMultiDoc(t *testing.T) {
	p, _ := NewParser(ParserConfig{DisableNetwork: true})
	yaml := "---\na: 1\n---\nb: 2\n"
	_, err := p.ParseStream(strings.NewReader(yaml))
	if err == nil {
		t.Fatal("expected error for multi-doc YAML")
	}
	if !strings.Contains(err.Error(), "multiple YAML documents") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestParseStreamEmptyYAML(t *testing.T) {
	p, _ := NewParser(ParserConfig{DisableNetwork: true})
	_, err := p.ParseStream(strings.NewReader(""))
	if err == nil {
		t.Fatal("expected error for empty YAML")
	}
	if !strings.Contains(err.Error(), "publiccodeYmlVersion") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestParseStreamVersionNotString(t *testing.T) {
	p, _ := NewParser(ParserConfig{DisableNetwork: true})
	yaml := "publiccodeYmlVersion: 123\n"
	_, err := p.ParseStream(strings.NewReader(yaml))
	if err == nil {
		t.Fatal("expected error for non-string version")
	}
	if !strings.Contains(err.Error(), "publiccodeYmlVersion") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestParseStreamUnsupportedVersion(t *testing.T) {
	p, _ := NewParser(ParserConfig{DisableNetwork: true})
	yaml := "publiccodeYmlVersion: \"99.99\"\n"
	_, err := p.ParseStream(strings.NewReader(yaml))
	if err == nil {
		t.Fatal("expected error for unsupported version")
	}
	if !strings.Contains(err.Error(), "unsupported version") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestParseStreamSyntaxError(t *testing.T) {
	p, _ := NewParser(ParserConfig{DisableNetwork: true})
	// Malformed YAML.
	yaml := "key: [\n"
	_, err := p.ParseStream(strings.NewReader(yaml))
	if err == nil {
		t.Fatal("expected error for malformed YAML")
	}
}

// TestIPACodesURLFetch verifies that WithIPACodesURL fetches and uses the
// provided list, making a code from the served file valid.
func TestIPACodesURLFetch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("TESTCODE\n"))
	}))
	defer srv.Close()

	p, err := NewParser(ParserConfig{IPACodesURL: srv.URL})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := p.validate.Var("TESTCODE", "is_italian_ipa_code"); err != nil {
		t.Errorf("expected TESTCODE to be valid with custom list: %v", err)
	}

	if err := p.validate.Var("pcm", "is_italian_ipa_code"); err == nil {
		t.Error("expected 'pcm' to be invalid when not in custom list")
	}
}

// TestIPACodesURLFetchError verifies that an unreachable IPACodesURL returns an
// error from NewParser.
func TestIPACodesURLFetchError(t *testing.T) {
	_, err := NewParser(ParserConfig{IPACodesURL: "http://127.0.0.1:1/ipa_codes.txt"})
	if err == nil {
		t.Fatal("expected error for unreachable IPACodesURL")
	}
}

// TestIPACodesDefaultEmbedded verifies that the embedded list is used when
// IPACodesURL is empty.
func TestIPACodesDefaultEmbedded(t *testing.T) {
	p, err := NewParser(ParserConfig{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := p.validate.Var("pcm", "is_italian_ipa_code"); err != nil {
		t.Errorf("expected 'pcm' to be valid with embedded list: %v", err)
	}
}

func TestParseStreamReaderError(t *testing.T) {
	p, _ := NewParser(ParserConfig{DisableNetwork: true})
	_, err := p.ParseStream(errReader{})
	if err == nil {
		t.Fatal("expected error from reader failure")
	}
	if !strings.Contains(err.Error(), "Can't read the stream") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestParseStreamMissingVersion(t *testing.T) {
	p, _ := NewParser(ParserConfig{DisableNetwork: true})
	// Valid YAML but no publiccodeYmlVersion key.
	yaml := "name: test\n"
	_, err := p.ParseStream(strings.NewReader(yaml))
	if err == nil {
		t.Fatal("expected error for missing publiccodeYmlVersion")
	}
	if !strings.Contains(err.Error(), "publiccodeYmlVersion") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestParseStreamWithBaseURL(t *testing.T) {
	// A parser with BaseURL set exercises the p.baseURL != nil branch in parseStream.
	p, _ := NewParser(ParserConfig{
		BaseURL:        "https://example.com/org/repo/",
		DisableNetwork: true,
	})
	// Minimal YAML: just a version so we get past early returns into base-URL computation.
	_, _ = p.ParseStream(strings.NewReader("publiccodeYmlVersion: \"0\"\n"))
}

func TestParseStreamNetworkURLUnknownVCS(t *testing.T) {
	// With DisableNetwork=false, no BaseURL, and a URL from an unknown VCS host,
	// vcsurl.GetRawRoot fails and parseStream returns early before any network checks.
	p, _ := NewParser(ParserConfig{DisableNetwork: false})
	yaml := "publiccodeYmlVersion: \"0\"\nurl: \"https://example.com/org/repo.git\"\n"
	_, _ = p.ParseStream(strings.NewReader(yaml))
}

func TestParseHTTPServer(t *testing.T) {
	content, err := os.ReadFile("testdata/v0/valid/valid.yml")
	if err != nil {
		t.Fatal(err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/yaml")
		_, _ = w.Write(content)
	}))
	defer srv.Close()

	p, _ := NewParser(ParserConfig{DisableNetwork: true})
	// Should successfully parse the YAML served over HTTP.
	_, _ = p.Parse(srv.URL + "/publiccode.yml")
}
