package publiccode

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestParseBlocksSSRFByDefault proves that, with the default configuration used
// by the publiccode-crawler, an attacker-controlled URL pointing at an internal
// (here: loopback) address is refused, preventing SSRF. The opt-in flag lets
// trusted callers reach it.
func TestParseBlocksSSRFByDefault(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/yaml")
		_, _ = w.Write([]byte("publiccodeYmlVersion: \"0.4\"\n"))
	}))
	defer srv.Close()

	p, err := NewParser(ParserConfig{})
	if err != nil {
		t.Fatal(err)
	}

	_, err = p.Parse(srv.URL + "/publiccode.yml")
	if err == nil {
		t.Fatal("expected the default parser to block the loopback (SSRF) request")
	}
	if !strings.Contains(err.Error(), "non-public") {
		t.Errorf("expected a blocked-address error, got: %v", err)
	}

	// With the opt-out enabled the same request is allowed to connect (it then
	// fails validation for unrelated reasons, but it is no longer blocked at the
	// network layer).
	pAllowed, err := NewParser(ParserConfig{AllowNetworkToPrivateHosts: true})
	if err != nil {
		t.Fatal(err)
	}

	if _, err = pAllowed.Parse(srv.URL + "/publiccode.yml"); err != nil {
		if strings.Contains(err.Error(), "non-public") {
			t.Errorf("opt-out client must not be blocked at the network layer, got: %v", err)
		}
	}
}
