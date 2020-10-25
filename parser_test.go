package publiccode

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path"
	"path/filepath"
	"strings"
	"testing"
)

type testType struct {
	file string
	err  error
}

// Parse the YAML file passed as argument.
// Return nil if the parsing succeded or an error if it failed.
func parse(path string) error {
	p := NewParser()
	p.Strict = false
	p.RemoteBaseURL = ""
	return p.ParseFile(path)
}

// Check all the YAML files matching the glob pattern and Fail for each file
// with parsing or validation errors.
func checkValidFiles(pattern string, t *testing.T) {
	testFiles, _ := filepath.Glob(pattern)
	for _, file := range testFiles {
		t.Run(file, func(t *testing.T) {
			err := parse(file)
			if (err != nil) {
				t.Errorf("[%s] validation failed for valid file: %T - %s\n", file, err, err)
			}
		})
	}
}

func checkParseErrors(t *testing.T, err error, test testType) {
	if test.err == nil && err != nil {
		t.Errorf("[%s] unexpected error: %v\n", test.file, err)
	} else if test.err != nil && err == nil {
		t.Errorf("[%s] no error generated\n", test.file)
	} else if test.err != nil && err != nil {
		if multi, ok := err.(ErrorParseMulti); ok {
			if len(multi) != 1 {
				t.Errorf("[%s] too many errors generated; 1 was expected but got:\n", test.file)
				for _, e := range multi {
					t.Errorf("  * %s\n", e)
				}
				return
			}
			err = multi[0]
		}

		if err != test.err {
			t.Errorf("[%s] wrong error generated:\n%T - %s\n- instead of:\n%T - %s", test.file, test.err, test.err, err, err)
		}
	}
}

// Test v0.1 invalid YAML testcases (testdata/v0.1/invalid/).
func TestInvalidTestcasesV0_1(t *testing.T) {
	expected := map[string]error{
		"publiccodeYmlVersion_missing.yml":	ErrorInvalidValue{
			"publiccodeYmlVersion", "missing mandatory key",
		},
		"name_missing.yml": ErrorInvalidValue{"name", "missing mandatory key"},
		"legal_license_missing.yml": ErrorInvalidValue{"legal/license", "missing mandatory key"},
		"localisation_availableLanguages_missing.yml": ErrorInvalidValue{
			"localisation/availableLanguages", "missing mandatory key",
		},
		"localisation_localisationReady_missing.yml": ErrorInvalidValue{
			"localisation/localisationReady", "missing mandatory key",
		},
		"maintenance_contacts_missing.yml": ErrorInvalidValue{
			"maintenance/contacts",
			"missing but mandatory for \"internal\" or \"community\" maintenance",
		},
		"maintenance_type_missing.yml": ErrorInvalidValue{"maintenance/type", "missing mandatory key"},
		"platforms_missing.yml": ErrorInvalidValue{"platforms", "missing mandatory key"},
		"developmentStatus_missing.yml": ErrorInvalidValue{"developmentStatus", "missing mandatory key"},
		"releaseDate_missing.yml": ErrorInvalidValue{"releaseDate", "missing mandatory key"},
		// "genericName_missing.yml": ErrorInvalidValue{"description/en/genericName", "missing mandatory key"},
		"shortDescription_missing.yml": ErrorInvalidValue{"description/en/shortDescription", "missing mandatory key"},
		"features_missing.yml": ErrorInvalidValue{"description/*/features", "missing mandatory key"},
		"features_empty.yml": ErrorInvalidValue{"description/*/features", "missing mandatory key"},
		"longDescription_missing.yml": ErrorInvalidValue{"description/*/longDescription", "missing mandatory key"},
		"softwareType_missing.yml": ErrorInvalidValue{"softwareType", "missing mandatory key"},
		"categories_missing.yml": ErrorInvalidValue{"categories", "missing mandatory key"},
		"url_missing.yml": ErrorInvalidValue{"url", "missing mandatory key"},
		"legal_license_invalid.yml": ErrorInvalidValue{
			"legal/license",
			"invalid value AGPLicense-3.0: invalid license \"AGPLicense-3.0\" at 0 (\"AGPLi\")",
		},
		"categories_nil.yml": ErrorInvalidValue{"categories", "invalid type <nil>."},
		"categories_empty.yml": ErrorInvalidValue{"categories", "invalid type <nil>."},
		"name_nil.yml": ErrorInvalidValue{"name", "invalid type <nil>."},
		"unicode_grapheme_clusters.yml": ErrorInvalidValue{"description/eng/longDescription", "too short (135), min 500 chars"},
		"file_encoding.yml": ParseError{"Invalid UTF-8"},
	}

    testFiles, _ := filepath.Glob("testdata/v0.1/invalid/*.yml")
	for _, file := range testFiles {
		baseName := path.Base(file)
		if expected[baseName] == nil {
			t.Errorf("No expected data for file %s", baseName)
		}
		t.Run(file, func(t *testing.T) {
			err := parse(file)
			checkParseErrors(t, err, testType{file, expected[baseName]})
		})
	}
}

// Test v0.1 valid YAML testcases (testdata/v0.1/valid/).
func TestValidTestcasesV0_1(t *testing.T) {
	checkValidFiles("testdata/v0.1/valid/*.yml", t)
}

// Test v0.2 valid YAML testcases (testdata/v0.2/valid/).
func TestValidTestcasesV0_2(t *testing.T) {
    checkValidFiles("testdata/v0.2/valid/*.yml", t)
}

// Test publiccode.yml remote files for key errors.
func TestDecodeValueErrorsRemote(t *testing.T) {
	testRemoteFiles := []testType{
		{"https://raw.githubusercontent.com/pagopa/io-app/master/publiccode.yml", nil},
	}

	for _, test := range testRemoteFiles {
		t.Run(fmt.Sprintf("%v", test.err), func(t *testing.T) {
			// Parse data into pc struct.
			p := NewParser()
			p.Strict = false
			p.RemoteBaseURL = "https://raw.githubusercontent.com/pagopa/io-app/master"
			err := p.ParseRemoteFile(test.file)

			checkParseErrors(t, err, test)
		})
	}
}

// Test that relative paths are turned into absolute paths.
func TestRelativePaths(t *testing.T) {
	// Parse file into pc struct.
	const url = "https://raw.githubusercontent.com/italia/18app/master/publiccode.yml"
	p := NewParser()
	p.Strict = false
	p.RemoteBaseURL = "https://raw.githubusercontent.com/italia/18app/master"
	err := p.ParseRemoteFile(url)
	if err != nil {
		t.Errorf("Failed to parse remote file from %v: %v", url, err)
	}

	if strings.Index(p.PublicCode.Description["it"].Screenshots[0], p.RemoteBaseURL) != 0 {
		t.Errorf("Relative path was not turned into absolute URL: %v", p.PublicCode.Description["it"].Screenshots[0])
	}
}

// Test that the exported YAML passes validation again, and that re-exporting it
// matches the first export (lossless roundtrip).
func TestExport(t *testing.T) {
	p := NewParser()
	p.Strict = false
	p.DisableNetwork = true
	err := p.ParseFile("testdata/v0.2/valid/valid.yml")
	if err != nil {
		t.Errorf("Failed to parse valid file: %v", err)
	}

	yaml1, err := p.ToYAML()
	if err != nil {
		t.Errorf("Failed to export YAML: %v", err)
	}
	_ = ioutil.WriteFile("testdata/v0.2/valid/valid.yml.golden", yaml1, 0644)

	p2 := NewParser()
	p2.Strict = false
	p2.DisableNetwork = true
	err = p2.Parse(yaml1)
	if err != nil {
		t.Errorf("Failed to parse exported file: %v", err)
	}

	yaml2, err := p2.ToYAML()
	if err != nil {
		t.Errorf("Failed to export YAML again: %v", err)
	}

	if !bytes.Equal(yaml1, yaml2) {
		t.Errorf("Exported YAML files do not match; roundtrip is not lossless")
	}
}
