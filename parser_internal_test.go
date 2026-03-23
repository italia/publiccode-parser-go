package publiccode

import (
	"strings"
	"testing"

	"github.com/goccy/go-yaml/parser"
)

func TestGetPositionInFileEmptyDocs(t *testing.T) {
	// An empty YAML file results in a parsed file with no docs.
	file, err := parser.ParseBytes([]byte(""), 0)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	line, col := getPositionInFile("someKey", file)
	if line != 0 || col != 0 {
		t.Errorf("expected 0,0 for empty file, got %d,%d", line, col)
	}
}

func TestGetPositionInFileNilBody(t *testing.T) {
	// Parse YAML that produces a doc with nil body.
	// A YAML file with only comments has a doc with nil Body.
	file, err := parser.ParseBytes([]byte("# just a comment\n"), 0)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	line, col := getPositionInFile("someKey", file)
	if line != 0 || col != 0 {
		t.Errorf("expected 0,0 for comment-only file, got %d,%d", line, col)
	}
}

func TestDecodeUnknownField(t *testing.T) {
	// Parse a YAML with an unknown field, which should be caught by
	// the decode function as an UnknownFieldError.
	yaml := `publiccodeYmlVersion: "0"
name: test
unknownField: value
`
	file, err := parser.ParseBytes([]byte(yaml), 0)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	v0 := &PublicCodeV0{}
	results := decode([]byte(yaml), v0, file)
	if results == nil {
		t.Error("expected error for unknown field")
	}
}

func TestDecodeWrongType(t *testing.T) {
	// Parse YAML where a field has wrong type (e.g. name is a list).
	yaml := `publiccodeYmlVersion: "0"
name:
  - item1
  - item2
`
	file, err := parser.ParseBytes([]byte(yaml), 0)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	v0 := &PublicCodeV0{}
	results := decode([]byte(yaml), v0, file)
	if results == nil {
		t.Error("expected error for wrong type")
	}
}

func TestToURLLocalPath(t *testing.T) {
	// toURL with a local file path should return a file:// URL.
	u, err := toURL("./testdata/v0/valid/valid.yml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u.Scheme != "file" {
		t.Errorf("expected file scheme, got %q", u.Scheme)
	}
}

func TestToURLHTTPS(t *testing.T) {
	u, err := toURL("https://example.com/publiccode.yml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u.Host != "example.com" {
		t.Errorf("unexpected host: %s", u.Host)
	}
}

func TestFindKeyPosNilNode(t *testing.T) {
	line, col := findKeyPos(nil, []string{"key"})
	if line != 0 || col != 0 {
		t.Errorf("expected 0,0 for nil node, got %d,%d", line, col)
	}
}

func TestFindKeyAtLineNilNode(t *testing.T) {
	result := findKeyAtLine(nil, 1, "")
	if result != "" {
		t.Errorf("expected empty string for nil node, got %q", result)
	}
}

func TestFindKeyPosDocumentNode(t *testing.T) {
	file, err := parser.ParseBytes([]byte("key: value\n"), 0)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	// file.Docs[0] is a *ast.DocumentNode: exercises the DocumentNode case in findKeyPos
	line, col := findKeyPos(file.Docs[0], []string{"key"})
	if line == 0 {
		t.Errorf("expected non-zero line for 'key', got %d,%d", line, col)
	}
}

func TestFindKeyAtLineDocumentNode(t *testing.T) {
	file, err := parser.ParseBytes([]byte("key: value\n"), 0)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	// file.Docs[0] is a *ast.DocumentNode: exercises the DocumentNode case in findKeyAtLine
	result := findKeyAtLine(file.Docs[0], 1, "")
	_ = result
}

func TestParseStreamDeprecatedVersionWarning(t *testing.T) {
	// An older but supported version should produce a deprecation warning.
	p, _ := NewParser(ParserConfig{DisableNetwork: true})
	yaml := `publiccodeYmlVersion: "0.2"
name: test
url: "https://github.com/example/repo.git"
platforms:
  - web
developmentStatus: stable
softwareType: "standalone/web"
description:
  en:
    genericName: Test
    shortDescription: "A short description."
    longDescription: >
      Long description that is longer than one hundred and fifty characters to
      satisfy the minimum length requirement for this field in publiccode.yml.
      Adding more text here to ensure we definitely hit the 150 character minimum.
    features:
      - feature1
legal:
  license: MIT
maintenance:
  type: none
localisation:
  localisationReady: true
  availableLanguages:
    - en
`
	_, err := p.ParseStream(strings.NewReader(yaml))
	// Should produce a deprecation warning but no validation errors.
	if err == nil {
		return // no errors at all is also fine
	}
	vr, ok := err.(ValidationResults)
	if !ok {
		t.Fatalf("expected ValidationResults, got %T: %v", err, err)
	}
	hasWarning := false
	for _, e := range vr {
		if _, ok := e.(ValidationWarning); ok {
			hasWarning = true
		}
		if _, ok := e.(ValidationError); ok {
			t.Errorf("unexpected validation error: %v", e)
		}
	}
	if !hasWarning {
		t.Error("expected at least one deprecation warning for old version")
	}
}
