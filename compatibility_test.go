package publiccode_test

import (
	"testing"

	legacypubliccode "github.com/italia/publiccode-parser-go/v5"
)

// TestLegacyModulePath guards the historical import path while the project is
// transitioning to the libpubliccode name.
func TestLegacyModulePath(t *testing.T) {
	parser, err := legacypubliccode.NewParser(legacypubliccode.ParserConfig{
		DisableExternalChecks: true,
	})
	if err != nil {
		t.Fatalf("NewParser() error = %v", err)
	}

	if parser == nil {
		t.Fatal("NewParser() returned a nil parser")
	}

	var _ legacypubliccode.PublicCode = (*legacypubliccode.PublicCodeV0)(nil)
	var _ legacypubliccode.PublicCode = (*legacypubliccode.PublicCodeV1)(nil)
}
