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
	validFile       = "../testdata/v0/valid/valid.minimal.yml"
	invalidFile     = "../testdata/v0/invalid/categories_invalid.yml"
	warningFile     = "../testdata/v0/valid_with_warnings/valid.minimal.v0.4.yml"
	notThereFile    = "../testdata/v0/valid/no-such-file.yml"
	missingLogoFile = "../testdata/v0/invalid/no-network/logo_missing_file.yml"
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
		"valid file":               {[]string{"-external-checks=none", validFile}, 0},
		"valid file, JSON":         {[]string{"-external-checks=none", "-json", validFile}, 0},
		"invalid file":             {[]string{"-external-checks=none", invalidFile}, 1},
		"invalid file, JSON":       {[]string{"-external-checks=none", "-json", invalidFile}, 1},
		"file with warnings":       {[]string{"-external-checks=none", warningFile}, 0},
		"file with warnings, JSON": {[]string{"-external-checks=none", "-json", warningFile}, 0},
		"unreadable file":          {[]string{"-external-checks=none", notThereFile}, 1},
		"unreadable file, JSON":    {[]string{"-external-checks=none", "-json", notThereFile}, 1},
		"no file argument":         {[]string{}, 2},
		"no file argument, JSON":   {[]string{"-json"}, 2},
		"unknown flag":             {[]string{"-bogus", validFile}, 2},
		"help":                     {[]string{"-help"}, 0},

		"local mode":            {[]string{"-external-checks=local", validFile}, 0},
		"unknown mode":          {[]string{"-external-checks=bogus", validFile}, 2},
		"clone mode":            {[]string{"-external-checks=clone", validFile}, 2},
		"deprecated no-network": {[]string{"-no-network", validFile}, 0},
		"deprecated no-external-checks": {
			[]string{"-no-external-checks", validFile}, 0,
		},
		"deprecated flag agreeing with the mode": {
			[]string{"-no-network", "-external-checks=local", validFile}, 0,
		},
		"deprecated flag conflicting with the mode": {
			[]string{"-no-network", "-external-checks=none", validFile}, 2,
		},
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

func TestRunLocalModeChecksLocalFiles(t *testing.T) {
	_, stdout, _ := runCLI(t, "-external-checks=local", missingLogoFile)

	if !strings.Contains(stdout, "no such file") {
		t.Errorf("expected the missing logo on stdout, got %q", stdout)
	}
}

func TestRunNoneModeSkipsLocalFiles(t *testing.T) {
	code, stdout, _ := runCLI(t, "-external-checks=none", missingLogoFile)

	if code != 0 || stdout != "" {
		t.Errorf("expected no check on the missing logo, got exit code %d and %q", code, stdout)
	}
}

func TestRunNoExternalChecksWinsOverNoNetwork(t *testing.T) {
	code, stdout, _ := runCLI(t, "-no-network", "-no-external-checks", missingLogoFile)

	if code != 0 || stdout != "" {
		t.Errorf("expected the none mode to win, got exit code %d and %q", code, stdout)
	}
}

func TestRunDeprecatedFlagsWarnOnStderr(t *testing.T) {
	tests := map[string]struct {
		flag string
		want string
	}{
		"no-network": {"-no-network", "-no-network is deprecated, use -external-checks=local\n"},
		"no-external-checks": {
			"-no-external-checks", "-no-external-checks is deprecated, use -external-checks=none\n",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			_, stdout, stderr := runCLI(t, test.flag, validFile)

			if stderr != test.want {
				t.Errorf("expected %q on stderr, got %q", test.want, stderr)
			}

			if stdout != "" {
				t.Errorf("expected no output on stdout, got %q", stdout)
			}
		})
	}
}

func TestRunUnknownModeListsTheValidOnes(t *testing.T) {
	_, _, stderr := runCLI(t, "-external-checks=bogus", validFile)

	if !strings.Contains(stderr, "none, local and network") {
		t.Errorf("expected the valid modes on stderr, got %q", stderr)
	}
}

func TestRunCloneModeIsNotImplementedYet(t *testing.T) {
	_, _, stderr := runCLI(t, "-external-checks=clone", validFile)

	if !strings.Contains(stderr, "not implemented") {
		t.Errorf("expected the clone mode to be refused, got %q", stderr)
	}
}

func TestRunConflictingModeIsRejected(t *testing.T) {
	_, _, stderr := runCLI(t, "-no-network", "-external-checks=none", validFile)

	if !strings.Contains(stderr, "-no-network and -external-checks=none") {
		t.Errorf("expected the conflict on stderr, got %q", stderr)
	}
}

func TestRunValidFileIsSilent(t *testing.T) {
	_, stdout, stderr := runCLI(t, "-external-checks=none", validFile)

	if stdout != "" {
		t.Errorf("expected no output on stdout, got %q", stdout)
	}

	if stderr != "" {
		t.Errorf("expected no output on stderr, got %q", stderr)
	}
}

func TestRunInvalidFileReportsOnStdout(t *testing.T) {
	_, stdout, _ := runCLI(t, "-external-checks=none", invalidFile)

	if !strings.Contains(stdout, "error: categories[0]") {
		t.Errorf("expected the validation error on stdout, got %q", stdout)
	}
}

func TestRunJSONValidFileIsEmptyList(t *testing.T) {
	_, stdout, _ := runCLI(t, "-external-checks=none", "-json", validFile)

	if strings.TrimSpace(stdout) != "[]" {
		t.Errorf("expected an empty JSON list, got %q", stdout)
	}
}

func TestRunJSONInvalidFileIsList(t *testing.T) {
	_, stdout, _ := runCLI(t, "-external-checks=none", "-json", invalidFile)

	results := unmarshalList(t, stdout)
	if len(results) == 0 {
		t.Fatalf("expected at least one entry, got %q", stdout)
	}

	if results[0]["type"] != "error" {
		t.Errorf("expected an entry of type error, got %v", results[0]["type"])
	}
}

func TestRunJSONReportsInputFileName(t *testing.T) {
	_, stdout, _ := runCLI(t, "-no-external-checks", "-json", invalidFile)

	results := unmarshalList(t, stdout)
	if len(results) == 0 {
		t.Fatalf("expected at least one entry, got %q", stdout)
	}

	if results[0]["file"] != "categories_invalid.yml" {
		t.Errorf("expected the name of the validated file, got %v", results[0]["file"])
	}
}

func TestRunJSONUnreadableFileIsList(t *testing.T) {
	_, stdout, stderr := runCLI(t, "-external-checks=none", "-json", notThereFile)

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
