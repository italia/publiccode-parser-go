package publiccode

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

type testType struct {
	file string
	err  error
}

// Test publiccode.yml local files for key errors.
func TestDecodeValueErrors(t *testing.T) {
	testFiles := []testType{
		// A complete and valid yml.
		{"tests/valid.yml", nil},
		// A complete and valid minimal yml.
		{"tests/valid.minimal.yml", nil},
		// Fields must be valid against different type
		{"tests/valid_maintenance_contacts_phone.yml", nil}, // Valid maintenance/contacts/phone.
		// Test if dependsOn multiple subkeys are kept
		{"tests/valid_dependsOn.yml", nil},

		// Missing mandatory fields.

		{"tests/missing_publiccodeYmlVersion.yml", ErrorInvalidValue{Key: "publiccodeYmlVersion", Reason: "missing mandatory key"}},
		{"tests/missing_name.yml", ErrorInvalidValue{Key: "name", Reason: "missing mandatory key"}},
		{"tests/missing_legal_license.yml", ErrorInvalidValue{Key: "legal/license", Reason: "missing mandatory key"}},
		{"tests/missing_localisation_availableLanguages.yml", ErrorInvalidValue{Key: "localisation/availableLanguages", Reason: "missing mandatory key"}},
		{"tests/missing_localisation_localisationReady.yml", ErrorInvalidValue{Key: "localisation/localisationReady", Reason: "missing mandatory key"}},
		{"tests/missing_maintenance_contacts.yml", ErrorInvalidValue{Key: "maintenance/contacts", Reason: "missing but mandatory for \"internal\" or \"community\" maintenance"}},
		{"tests/missing_maintenance_type.yml", ErrorInvalidValue{Key: "maintenance/type", Reason: "missing mandatory key"}},
		{"tests/missing_platforms.yml", ErrorInvalidValue{Key: "platforms", Reason: "missing mandatory key"}},
		{"tests/missing_developmentStatus.yml", ErrorInvalidValue{Key: "developmentStatus", Reason: "missing mandatory key"}},
		{"tests/missing_releaseDate.yml", ErrorInvalidValue{Key: "releaseDate", Reason: "missing mandatory key"}},
		{"tests/missing_genericName.yml", ErrorInvalidValue{Key: "description/en/genericName", Reason: "missing mandatory key"}},
		{"tests/missing_shortDescription.yml", ErrorInvalidValue{Key: "description/en/shortDescription", Reason: "missing mandatory key"}},
		{"tests/missing_features.yml", ErrorInvalidValue{Key: "description/*/features", Reason: "missing mandatory key"}},
		{"tests/empty_features.yml", ErrorInvalidValue{Key: "description/*/features", Reason: "missing mandatory key"}},
		{"tests/missing_longDescription.yml", ErrorInvalidValue{Key: "description/*/longDescription", Reason: "missing mandatory key"}},
		{"tests/missing_softwareType.yml", ErrorInvalidValue{Key: "softwareType", Reason: "missing mandatory key"}},
		{"tests/missing_categories.yml", ErrorInvalidValue{Key: "categories", Reason: "missing mandatory key"}},
		{"tests/missing_url.yml", ErrorInvalidValue{Key: "url", Reason: "missing mandatory key"}},

		// Invalid fields.

		// Invalid legal/license.
		{"tests/invalid_legal_license.yml", ErrorInvalidValue{Key: "legal/license", Reason: "invalid value AGPLicense-3.0: invalid license \"AGPLicense-3.0\" at 0 (\"AGPLi\")"}},
		// Invalid categories (nil).
		{"tests/nil_categories.yml", ErrorInvalidValue{Key: "categories", Reason: "invalid value AGPLicense-3.0: invalid license \"AGPLicense-3.0\" at 0 (\"AGPLi\")"} },
		// Invalid name (nil).
		{"tests/nil_name.yml", ErrorInvalidValue{Key: "name", Reason: "invalid value AGPLicense-3.0: invalid license \"AGPLicense-3.0\" at 0 (\"AGPLi\")"} },
	}

	for _, test := range testFiles {
		t.Run(fmt.Sprintf("%v", test.err), func(t *testing.T) {
			// Parse file into pc struct.
			p := NewParser()
			p.Strict = false
			p.RemoteBaseURL = ""
			err := p.ParseFile(test.file)

			checkParseErrors(t, err, test)

			if test.file == "tests/valid.yml" {
				if !strings.Contains(p.OEmbed["https://www.youtube.com/watch?v=RaHmGbBOP84"], "<iframe ") {
					t.Errorf("Missing Oembed info")
				}
				if _, ok := p.PublicCode.Description["en"]; !ok {
					t.Errorf("Missing description/en")
				}
			}
			if test.file == "tests/valid_dependsOn.yml" {
				if len(p.PublicCode.DependsOn.Open) != 2 {
					t.Errorf("dependsOn/open length mismatch")
				}
				if len(p.PublicCode.DependsOn.Proprietary) != 1 {
					t.Errorf("dependsOn/proprietary length mismatch")
				}
				if len(p.PublicCode.DependsOn.Hardware) != 3 {
					t.Errorf("dependsOn/hardware length mismatch")
				}
			}
			if test.file == "tests/valid_maintenance_contacts_phone.yml" {
				if len(p.PublicCode.Maintenance.Contacts) != 3 {
					t.Errorf("maintenance/contacts length mismatch")
				}
				if len(p.PublicCode.Maintenance.Contractors) != 1 {
					t.Errorf("maintenance/contractors length mismatch")
				}
			}
		})
	}

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
	err := p.ParseFile("tests/valid.yml")
	if err != nil {
		t.Errorf("Failed to parse valid file: %v", err)
	}

	yaml1, err := p.ToYAML()
	if err != nil {
		t.Errorf("Failed to export YAML: %v", err)
	}

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
