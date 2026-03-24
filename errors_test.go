package publiccode

import (
	"encoding/json"
	"testing"
)

func TestParseErrorError(t *testing.T) {
	e := ParseError{Reason: "something went wrong"}
	if e.Error() != "something went wrong" {
		t.Errorf("unexpected: %q", e.Error())
	}
}

func TestValidationErrorErrorEmptyKey(t *testing.T) {
	e := ValidationError{Key: "", Description: "bad field", Line: 1, Column: 2}
	got := e.Error()
	want := "publiccode.yml:1:2: error: bad field"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestValidationErrorErrorWithKey(t *testing.T) {
	e := ValidationError{Key: "foo", Description: "bad field", Line: 3, Column: 4}
	got := e.Error()
	want := "publiccode.yml:3:4: error: foo: bad field"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestValidationErrorMarshalJSON(t *testing.T) {
	e := ValidationError{Key: "foo", Description: "bad", Line: 1, Column: 1}
	b, err := e.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}

	if m["type"] != "error" {
		t.Errorf("expected type=error, got %v", m["type"])
	}
	if m["key"] != "foo" {
		t.Errorf("expected key=foo, got %v", m["key"])
	}
}

func TestValidationWarningMarshalJSON(t *testing.T) {
	e := ValidationWarning{Key: "foo", Description: "minor", Line: 2, Column: 3}
	b, err := e.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}

	if m["type"] != "warning" {
		t.Errorf("expected type=warning, got %v", m["type"])
	}
}

func TestValidationWarningErrorEmptyKey(t *testing.T) {
	e := ValidationWarning{Key: "", Description: "minor issue", Line: 5, Column: 6}
	got := e.Error()
	want := "publiccode.yml:5:6: warning: minor issue"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
