package netutil

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"syscall"
	"time"
)

// MaxResponseBytes caps how many bytes the parser reads from a single remote
// resource. Without a cap an attacker-controlled URL in a publiccode.yml could
// point at an endless or huge response and exhaust the memory/disk of the host
// running the parser (e.g. the publiccode-crawler).
const MaxResponseBytes int64 = 20 << 20 // 20 MiB

// ErrBlockedAddress is returned when a request would connect to a non-public
// (loopback, private, link-local, ...) address. It protects against SSRF where
// a URL in an untrusted publiccode.yml is used to reach internal services or
// cloud metadata endpoints.
var ErrBlockedAddress = errors.New("blocked attempt to connect to a non-public address")

// ErrResponseTooLarge is returned when a remote response exceeds MaxResponseBytes.
var ErrResponseTooLarge = fmt.Errorf("response body exceeds the maximum allowed size of %d bytes", MaxResponseBytes)

// isPublicIP reports whether ip is a globally routable address the parser is
// allowed to reach. Loopback, unspecified, private (RFC 1918 / RFC 4193),
// link-local (incl. the 169.254.169.254 cloud metadata address), CGNAT and
// multicast ranges are refused.
func isPublicIP(ip net.IP) bool {
	if ip == nil {
		return false
	}

	if ip.IsLoopback() || ip.IsUnspecified() || ip.IsPrivate() ||
		ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsMulticast() ||
		ip.IsInterfaceLocalMulticast() {
		return false
	}

	// 100.64.0.0/10 (CGNAT, RFC 6598) is not covered by IsPrivate().
	if ip4 := ip.To4(); ip4 != nil {
		if ip4[0] == 100 && ip4[1] >= 64 && ip4[1] <= 127 {
			return false
		}
	}

	return true
}

// newDialControl returns a net.Dialer Control function that rejects connections
// to non-public addresses. It runs after DNS resolution, with the concrete IP
// about to be dialed, so it also defeats DNS-rebinding and is re-evaluated on
// every redirect hop.
func newDialControl() func(network, address string, c syscall.RawConn) error {
	return func(_, address string, _ syscall.RawConn) error {
		host, _, err := net.SplitHostPort(address)
		if err != nil {
			return fmt.Errorf("parsing dial address %q: %w", address, err)
		}

		ip := net.ParseIP(host)
		if !isPublicIP(ip) {
			return fmt.Errorf("%w: %s", ErrBlockedAddress, address)
		}

		return nil
	}
}

// limitedBody wraps a response body and returns ErrResponseTooLarge once more
// than max bytes have been read, so a single response cannot exhaust memory.
type limitedBody struct {
	body io.ReadCloser
	max  int64
	read int64
}

func (l *limitedBody) Read(p []byte) (int, error) {
	if l.read > l.max {
		return 0, ErrResponseTooLarge
	}

	n, err := l.body.Read(p)
	l.read += int64(n)

	if l.read > l.max {
		return n, ErrResponseTooLarge
	}

	return n, err //nolint:wrapcheck // must forward io.EOF and other sentinels unwrapped
}

func (l *limitedBody) Close() error {
	return l.body.Close() //nolint:wrapcheck // transparent pass-through of the wrapped body
}

// safeTransport wraps a base RoundTripper enforcing the response size limit.
// SSRF protection is enforced one layer below, in the dialer Control function.
type safeTransport struct {
	base http.RoundTripper
	max  int64
}

func (t *safeTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := t.base.RoundTrip(req)
	if err != nil {
		return nil, err //nolint:wrapcheck // http.Client inspects the transport error; keep it intact
	}

	// Reject early when the server advertises an over-limit Content-Length.
	if resp.ContentLength > t.max {
		_ = resp.Body.Close()

		return nil, ErrResponseTooLarge
	}

	resp.Body = &limitedBody{body: resp.Body, max: t.max}

	return resp, nil
}

// SafeHTTPClient builds an *http.Client hardened against SSRF and unbounded
// downloads. When allowPrivate is true the SSRF address filtering is disabled
// (used for trusted input and tests that target loopback servers); the response
// size limit is always enforced.
func SafeHTTPClient(timeout time.Duration, allowPrivate bool) *http.Client {
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}
	if !allowPrivate {
		dialer.Control = newDialControl()
	}

	transport := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           dialer.DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	return &http.Client{
		Timeout:   timeout,
		Transport: &safeTransport{base: transport, max: MaxResponseBytes},
	}
}
