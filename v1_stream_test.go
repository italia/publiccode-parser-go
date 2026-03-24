package publiccode

import (
	"strings"
	"testing"
)

// TestParseStreamV1Branch tests the v1 code path in parseStream by temporarily
// adding "1" to SupportedVersions.
func TestParseStreamV1Branch(t *testing.T) {
	// Temporarily add "1" to SupportedVersions.
	original := SupportedVersions
	SupportedVersions = append(SupportedVersions, "1")
	defer func() { SupportedVersions = original }()

	p, _ := NewParser(ParserConfig{DisableNetwork: true, DisableExternalChecks: true})

	yaml := `publiccodeYmlVersion: "1"
name: TestApp
url: "https://github.com/italia/developers.italia.it.git"
platforms:
  - web
developmentStatus: stable
softwareType: "standalone/web"
description:
  en:
    shortDescription: "Short description."
    longDescription: >
      Long description that is longer than one hundred and fifty characters
      to satisfy the minimum length requirement for this field in publiccode.yml.
      This is definitely more than enough text to meet the requirement here.
    features:
      - Feature one
legal:
  license: MIT
maintenance:
  type: none
localisation:
  localisationReady: true
  availableLanguages:
    - en
`
	pc, err := p.ParseStream(strings.NewReader(yaml))
	if err != nil {
		// A deprecation warning about version is expected, but no hard errors.
		vr, ok := err.(ValidationResults)
		if !ok {
			t.Fatalf("unexpected error type: %T: %v", err, err)
		}
		for _, e := range vr {
			if _, ok := e.(ValidationError); ok {
				t.Errorf("unexpected validation error: %v", e)
			}
		}
	}

	if pc == nil {
		t.Fatal("expected non-nil PublicCode result")
	}

	if pc.Version() != 1 {
		t.Errorf("expected version 1, got %d", pc.Version())
	}
}

// TestParseStreamV1BranchDecodeError tests that decode errors in v1 are handled.
func TestParseStreamV1BranchDecodeError(t *testing.T) {
	original := SupportedVersions
	SupportedVersions = append(SupportedVersions, "1")
	defer func() { SupportedVersions = original }()

	p, _ := NewParser(ParserConfig{DisableNetwork: true, DisableExternalChecks: true})

	// applicationSuite should be a string, not a list.
	yaml := `publiccodeYmlVersion: "1"
applicationSuite:
  - not-a-string
name: TestApp
`
	_, err := p.ParseStream(strings.NewReader(yaml))
	if err == nil {
		t.Fatal("expected error for wrong type in applicationSuite")
	}
}
