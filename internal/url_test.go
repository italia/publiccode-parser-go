package netutil

import (
	"net/url"
	"testing"
)

func TestIsValidURLHTTPS(t *testing.T) {
	ok, u := IsValidURL("https://example.com")
	if !ok {
		t.Error("expected valid URL")
	}
	if u == nil {
		t.Error("expected non-nil url.URL")
	}
}

func TestIsValidURLHTTP(t *testing.T) {
	ok, u := IsValidURL("http://example.com/path")
	if !ok {
		t.Error("expected valid URL")
	}
	if u == nil {
		t.Error("expected non-nil url.URL")
	}
}

func TestIsValidURLInvalid(t *testing.T) {
	ok, u := IsValidURL("not a url")
	if ok {
		t.Error("expected invalid URL")
	}
	if u != nil {
		t.Error("expected nil url.URL")
	}
}

func TestIsValidURLFTPScheme(t *testing.T) {
	ok, u := IsValidURL("ftp://example.com")
	if ok {
		t.Error("expected ftp to be unsupported")
	}
	if u != nil {
		t.Error("expected nil url.URL")
	}
}

func TestIsValidURLNoHost(t *testing.T) {
	ok, u := IsValidURL("file:///foo")
	if ok {
		t.Error("expected false for file:// (no host)")
	}
	if u != nil {
		t.Error("expected nil url.URL")
	}
}

func TestDisplayURLFile(t *testing.T) {
	u := &url.URL{Scheme: "file", Path: "/foo/bar"}
	got := DisplayURL(u)
	if got != "/foo/bar" {
		t.Errorf("expected /foo/bar, got %q", got)
	}
}

func TestDisplayURLHTTPS(t *testing.T) {
	u := &url.URL{Scheme: "https", Host: "example.com", Path: "/foo"}
	got := DisplayURL(u)
	if got != "https://example.com/foo" {
		t.Errorf("unexpected: %q", got)
	}
}
