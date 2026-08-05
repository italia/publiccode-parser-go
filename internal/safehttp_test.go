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

	blocked := SafeHTTPClient(5*time.Second, false)
	if _, err := blocked.Get(srv.URL); err == nil {
		t.Fatal("expected the SSRF guard to block the loopback request")
	} else if !errors.Is(err, ErrBlockedAddress) && !strings.Contains(err.Error(), "non-public") {
		t.Errorf("expected a blocked-address error, got: %v", err)
	}

	allowed := SafeHTTPClient(5*time.Second, true)
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
