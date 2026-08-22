package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	publiccode "github.com/italia/publiccode-parser-go/v5"
)

const (
	validFile    = "../testdata/v0/valid/valid.minimal.yml"
	invalidFile  = "../testdata/v0/invalid/categories_invalid.yml"
	warningFile  = "../testdata/v0/valid_with_warnings/valid.minimal.v0.4.yml"
	notThereFile = "../testdata/v0/valid/no-such-file.yml"
)

func runCLI(t *testing.T, args ...string) (int, string, string) {
	t.Helper()

	var stdout, stderr bytes.Buffer

	code := run(args, &stdout, &stderr)

	return code, stdout.String(), stderr.String()
}

func TestRunExitCodes(t *testing.T) {
	tests := map[string]struct {
		args []string
		want int
	}{
		"valid file":               {[]string{"-no-external-checks", validFile}, 0},
		"valid file, JSON":         {[]string{"-no-external-checks", "-json", validFile}, 0},
		"invalid file":             {[]string{"-no-external-checks", invalidFile}, 1},
		"invalid file, JSON":       {[]string{"-no-external-checks", "-json", invalidFile}, 1},
		"file with warnings":       {[]string{"-no-external-checks", warningFile}, 0},
		"file with warnings, JSON": {[]string{"-no-external-checks", "-json", warningFile}, 0},
		"unreadable file":          {[]string{"-no-external-checks", notThereFile}, 1},
		"unreadable file, JSON":    {[]string{"-no-external-checks", "-json", notThereFile}, 1},
		"no file argument":         {[]string{}, 2},
		"no file argument, JSON":   {[]string{"-json"}, 2},
		"unknown flag":             {[]string{"-bogus", validFile}, 2},
		"help":                     {[]string{"-help"}, 0},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			code, stdout, stderr := runCLI(t, test.args...)
			if code != test.want {
				t.Errorf(
					"expected exit code %d, got %d\nstdout: %s\nstderr: %s",
					test.want, code, stdout, stderr,
				)
			}
		})
	}
}

func TestRunValidFileIsSilent(t *testing.T) {
	_, stdout, stderr := runCLI(t, "-no-external-checks", validFile)

	if stdout != "" {
		t.Errorf("expected no output on stdout, got %q", stdout)
	}

	if stderr != "" {
		t.Errorf("expected no output on stderr, got %q", stderr)
	}
}

func TestRunInvalidFileReportsOnStdout(t *testing.T) {
	_, stdout, _ := runCLI(t, "-no-external-checks", invalidFile)

	if !strings.Contains(stdout, "error: categories[0]") {
		t.Errorf("expected the validation error on stdout, got %q", stdout)
	}
}

func TestRunJSONValidFileIsEmptyList(t *testing.T) {
	_, stdout, _ := runCLI(t, "-no-external-checks", "-json", validFile)

	if strings.TrimSpace(stdout) != "[]" {
		t.Errorf("expected an empty JSON list, got %q", stdout)
	}
}

func TestRunJSONInvalidFileIsList(t *testing.T) {
	_, stdout, _ := runCLI(t, "-no-external-checks", "-json", invalidFile)

	results := unmarshalList(t, stdout)
	if len(results) == 0 {
		t.Fatalf("expected at least one entry, got %q", stdout)
	}

	if results[0]["type"] != "error" {
		t.Errorf("expected an entry of type error, got %v", results[0]["type"])
	}
}

func TestRunJSONUnreadableFileIsList(t *testing.T) {
	_, stdout, stderr := runCLI(t, "-no-external-checks", "-json", notThereFile)

	results := unmarshalList(t, stdout)
	if len(results) != 1 {
		t.Fatalf("expected exactly one entry, got %q", stdout)
	}

	if description, _ := results[0]["description"].(string); !strings.Contains(description, "no-such-file.yml") {
		t.Errorf("expected the file name in the description, got %v", results[0]["description"])
	}

	if stderr != "" {
		t.Errorf("expected no output on stderr, got %q", stderr)
	}
}

func TestRunNoArgumentPrintsUsageOnStderr(t *testing.T) {
	_, stdout, stderr := runCLI(t)

	if stdout != "" {
		t.Errorf("expected no output on stdout, got %q", stdout)
	}

	if !strings.HasPrefix(stderr, "Usage: publiccode-parser") {
		t.Errorf("expected the usage on stderr, got %q", stderr)
	}
}

func TestRunHelpPrintsUsageOnStdout(t *testing.T) {
	_, stdout, stderr := runCLI(t, "-help")

	if !strings.HasPrefix(stdout, "Usage: publiccode-parser") {
		t.Errorf("expected the usage on stdout, got %q", stdout)
	}

	if stderr != "" {
		t.Errorf("expected no output on stderr, got %q", stderr)
	}
}

func unmarshalList(t *testing.T, out string) []map[string]any {
	t.Helper()

	var results []map[string]any

	if err := json.Unmarshal([]byte(out), &results); err != nil {
		t.Fatalf("expected a JSON list, got %q: %s", out, err)
	}

	return results
}

func TestHasValidationErrorsNil(t *testing.T) {
	if hasValidationErrors(nil) {
		t.Error("expected false for nil")
	}
}

func TestHasValidationErrorsOnlyWarnings(t *testing.T) {
	vr := publiccode.ValidationResults{
		publiccode.ValidationWarning{Key: "foo", Description: "minor warning"},
	}
	if hasValidationErrors(vr) {
		t.Error("expected false for ValidationResults with only warnings")
	}
}

func TestHasValidationErrorsWithError(t *testing.T) {
	vr := publiccode.ValidationResults{
		publiccode.ValidationError{Key: "foo", Description: "bad field"},
	}
	if !hasValidationErrors(vr) {
		t.Error("expected true for ValidationResults with a ValidationError")
	}
}

func TestHasValidationErrorsPlainError(t *testing.T) {
	if !hasValidationErrors(errors.New("some error")) {
		t.Error("expected true for a plain error")
	}
}

func TestHasValidationErrorsMixedResults(t *testing.T) {
	vr := publiccode.ValidationResults{
		publiccode.ValidationWarning{Key: "foo", Description: "warning"},
		publiccode.ValidationError{Key: "bar", Description: "error"},
	}
	if !hasValidationErrors(vr) {
		t.Error("expected true for mixed results containing an error")
	}
}
