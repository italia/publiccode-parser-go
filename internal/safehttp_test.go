package netutil

import (
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestIsPublicIP(t *testing.T) {
	cases := map[string]bool{
		"8.8.8.8":         true,
		"1.1.1.1":         true,
		"93.184.216.34":   true,
		"127.0.0.1":       false, // loopback
		"10.0.0.1":        false, // private
		"192.168.1.1":     false, // private
		"172.16.0.1":      false, // private
		"169.254.169.254": false, // link-local (cloud metadata)
		"100.64.0.1":      false, // CGNAT
		"0.0.0.0":         false, // unspecified
		"::1":             false, // IPv6 loopback
		"fc00::1":         false, // IPv6 ULA
		"fe80::1":         false, // IPv6 link-local
	}

	for addr, want := range cases {
		got := isPublicIP(net.ParseIP(addr))
		if got != want {
			t.Errorf("isPublicIP(%s) = %v, want %v", addr, got, want)
		}
	}
}

// TestSafeHTTPClientBlocksPrivate proves the default (allowPrivate=false) client
// refuses to connect to a loopback address, defeating SSRF, while the opt-out
// client can reach it.
func TestSafeHTTPClientBlocksPrivate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	blocked := SafeHTTPClient(5*time.Second, false, "")
	if _, err := blocked.Get(srv.URL); err == nil {
		t.Fatal("expected the SSRF guard to block the loopback request")
	} else if !errors.Is(err, ErrBlockedAddress) && !strings.Contains(err.Error(), "non-public") {
		t.Errorf("expected a blocked-address error, got: %v", err)
	}

	allowed := SafeHTTPClient(5*time.Second, true, "")
	resp, err := allowed.Get(srv.URL)
	if err != nil {
		t.Fatalf("expected the opt-out client to reach the server, got: %v", err)
	}
	_ = resp.Body.Close()
}

// roundTripFunc lets us stub a RoundTripper.
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func newResponse(body string, contentLength int64) *http.Response {
	return &http.Response{
		StatusCode:    200,
		Body:          io.NopCloser(strings.NewReader(body)),
		ContentLength: contentLength,
		Header:        make(http.Header),
	}
}

func TestResponseSizeLimitOnRead(t *testing.T) {
	// Unknown Content-Length, body larger than max: the limit must trigger while
	// reading the body.
	transport := &safeTransport{
		base: roundTripFunc(func(_ *http.Request) (*http.Response, error) {
			return newResponse(strings.Repeat("A", 100), -1), nil
		}),
		max: 10,
	}

	req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)
	resp, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected transport error: %v", err)
	}

	_, err = io.ReadAll(resp.Body)
	if !errors.Is(err, ErrResponseTooLarge) {
		t.Errorf("expected ErrResponseTooLarge while reading, got: %v", err)
	}
}

func TestResponseSizeLimitOnContentLength(t *testing.T) {
	// Advertised Content-Length over the limit must be rejected up front.
	transport := &safeTransport{
		base: roundTripFunc(func(_ *http.Request) (*http.Response, error) {
			return newResponse("small", 1000), nil
		}),
		max: 10,
	}

	req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)
	if _, err := transport.RoundTrip(req); !errors.Is(err, ErrResponseTooLarge) {
		t.Errorf("expected ErrResponseTooLarge from Content-Length check, got: %v", err)
	}
}

func TestResponseUnderLimitIsReadFully(t *testing.T) {
	const body = "within limit"
	transport := &safeTransport{
		base: roundTripFunc(func(_ *http.Request) (*http.Response, error) {
			return newResponse(body, int64(len(body))), nil
		}),
		max: MaxResponseBytes,
	}

	req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)
	resp, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("unexpected read error: %v", err)
	}
	if string(got) != body {
		t.Errorf("body mismatch: got %q, want %q", got, body)
	}
}

func TestUserAgentTransportSetsTheDefault(t *testing.T) {
	const userAgent = "publiccode-parser-go/1.2.3 (+https://example.org)"

	var got string
	transport := &userAgentTransport{
		base: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			got = r.Header.Get("User-Agent")

			return newResponse("", 0), nil
		}),
		userAgent: userAgent,
	}

	req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)
	if _, err := transport.RoundTrip(req); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != userAgent {
		t.Errorf("User-Agent sent: got %q, want %q", got, userAgent)
	}

	// A RoundTripper must not modify the request it is handed.
	if _, set := req.Header["User-Agent"]; set {
		t.Errorf("the original request was modified: %q", req.Header.Get("User-Agent"))
	}
}

func TestUserAgentTransportKeepsTheCallerHeader(t *testing.T) {
	cases := map[string]string{
		"caller sets its own": "harvester/2.0",
		"caller clears it":    "",
	}

	for name, want := range cases {
		t.Run(name, func(t *testing.T) {
			var got string
			var seen bool

			transport := &userAgentTransport{
				base: roundTripFunc(func(r *http.Request) (*http.Response, error) {
					got, seen = r.Header.Get("User-Agent"), true

					return newResponse("", 0), nil
				}),
				userAgent: "publiccode-parser-go/1.2.3",
			}

			req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)
			req.Header["User-Agent"] = []string{want}

			if _, err := transport.RoundTrip(req); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !seen {
				t.Fatal("the base transport was not called")
			}
			if got != want {
				t.Errorf("User-Agent sent: got %q, want %q", got, want)
			}
		})
	}
}

// TestSafeHTTPClientSendsUserAgent checks the User-Agent all the way down to the
// wire, and that an empty one leaves Go's default alone.
func TestSafeHTTPClientSendsUserAgent(t *testing.T) {
	const userAgent = "publiccode-parser-go/1.2.3 (+https://example.org)"

	// The server echoes back what it received, so there is nothing shared
	// between the handler and the test.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, r.UserAgent())
	}))
	defer srv.Close()

	got, err := getBody(SafeHTTPClient(5*time.Second, true, userAgent), srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != userAgent {
		t.Errorf("User-Agent received by the server: got %q, want %q", got, userAgent)
	}

	if got, err = getBody(SafeHTTPClient(5*time.Second, true, ""), srv.URL); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(got, "Go-http-client/") {
		t.Errorf("an empty User-Agent must leave Go's default in place, got %q", got)
	}
}

// getBody GETs url and returns the response body as a string.
func getBody(client *http.Client, url string) (string, error) {
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	return string(body), err
}
