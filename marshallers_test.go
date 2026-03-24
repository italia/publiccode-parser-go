package publiccode

import (
	"encoding/json"
	"net/url"
	"testing"
)

func TestURLMarshalJSONNil(t *testing.T) {
	var u *URL
	b, err := u.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != "null" {
		t.Errorf("expected null, got %s", b)
	}
}

func TestURLMarshalJSONNonNil(t *testing.T) {
	raw, _ := url.Parse("https://example.com/foo")
	u := (*URL)(raw)
	b, err := u.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}

	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		t.Fatal(err)
	}
	if s != "https://example.com/foo" {
		t.Errorf("unexpected: %q", s)
	}
}

func TestURLUnmarshalYAMLError(t *testing.T) {
	var u URL
	// Pass an unmarshal function that always returns an error.
	err := u.UnmarshalYAML(func(v any) error {
		return &url.Error{Op: "unmarshal", URL: "test", Err: url.EscapeError("forced error")}
	})
	if err == nil {
		t.Error("expected error from unmarshal, got nil")
	}
}

func TestUrlOrUrlArrayUnmarshalYAMLSingle(t *testing.T) {
	var a UrlOrUrlArray
	err := a.UnmarshalYAML(func(v any) error {
		switch dst := v.(type) {
		case *[]*URL:
			// Simulate failure for multi-URL so it falls back to single.
			return &url.Error{Op: "unmarshal", URL: "test", Err: url.EscapeError("not a list")}
		case *URL:
			raw, _ := url.Parse("https://example.com")
			*dst = URL(*raw)
			return nil
		}
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(a) != 1 {
		t.Errorf("expected 1 element, got %d", len(a))
	}
}

func TestUrlOrUrlArrayUnmarshalYAMLMulti(t *testing.T) {
	raw1, _ := url.Parse("https://example.com/1")
	raw2, _ := url.Parse("https://example.com/2")
	u1 := (*URL)(raw1)
	u2 := (*URL)(raw2)

	var a UrlOrUrlArray
	err := a.UnmarshalYAML(func(v any) error {
		if dst, ok := v.(*[]*URL); ok {
			*dst = []*URL{u1, u2}
			return nil
		}
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(a) != 2 {
		t.Errorf("expected 2 elements, got %d", len(a))
	}
}

func TestURLUnmarshalYAMLBadURLString(t *testing.T) {
	var u URL
	// Unmarshal succeeds but returns a string that url.Parse cannot parse.
	err := u.UnmarshalYAML(func(v any) error {
		if s, ok := v.(*string); ok {
			*s = "://missing-scheme"
			return nil
		}
		return nil
	})
	if err == nil {
		t.Error("expected error from url.Parse for malformed URL string")
	}
}

func TestUrlOrUrlArrayUnmarshalYAMLBothFail(t *testing.T) {
	var a UrlOrUrlArray
	err := a.UnmarshalYAML(func(v any) error {
		return &url.Error{Op: "unmarshal", URL: "test", Err: url.EscapeError("always fails")}
	})
	if err == nil {
		t.Error("expected error, got nil")
	}
}
