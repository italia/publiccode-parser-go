package publiccode

import (
	"strings"
	"testing"
)

type testType struct {
	file   string
	errkey string
}

// Test publiccode.yml local files for key errors.
func TestDecodeValueErrors(t *testing.T) {
	RemoteBaseURL = ""

	testFiles := []testType{
		// A complete and valid yml.
		{"tests/valid.yml", ""},
		//A complete and valid minimal yml.
		{"tests/valid.minimal.yml", ""},

		//Missing mandatory fields.
		{"tests/missing_publiccodeYmlVersion.yml", "publiccodeYmlVersion"},                       // Missing version.
		{"tests/missing_name.yml", "name"},                                                       // Missing name.
		{"tests/missing_legal_license.yml", "legal/license"},                                     // Missing legal/license.
		{"tests/missing_localisation_availableLanguages.yml", "localisation/availableLanguages"}, // Missing localisation/availableLanguages.
		{"tests/missing_localisation_localisationReady.yml", "localisation/localisationReady"},   // Missing localisation/localisationReady.
		{"tests/missing_maintenance_contacts.yml", "maintenance/contacts"},                       // Missing maintenance/contacts.
		{"tests/missing_maintenance_type.yml", "maintenance/type"},                               // Missing maintenance/type.
		{"tests/missing_platforms.yml", "platforms"},                                             // Missing platforms.
		{"tests/missing_releaseDate.yml", "releaseDate"},                                         // Missing releaseDate.
		{"tests/missing_softwareType.yml", "softwareType"},                                       // Missing softwareType/type.
		{"tests/missing_tags.yml", "tags"},                                                       // Missing tags.
		{"tests/missing_url.yml", "url"},                                                         // Missing url.

		// Invalid fields.
		{"tests/invalid_legal_license.yml", "legal/license"}, // Invalid legal/license.
	}

	for _, test := range testFiles {
		t.Run(test.errkey, func(t *testing.T) {
			// Parse file into pc struct.
			p := NewParser()
			err := p.ParseFile(test.file)

			checkParseErrors(t, err, test)

			if test.file == "tests/valid.yml" {
				if !strings.Contains(p.OEmbed["https://www.youtube.com/watch?v=RaHmGbBOP84"], "<iframe ") {
					t.Errorf("Missing Oembed info")
				}
			}
		})
	}

}

// Test publiccode.yml remote files for key errors.
func TestDecodeValueErrorsRemote(t *testing.T) {
	RemoteBaseURL = "https://raw.githubusercontent.com/gith002/Medusa/master/"

	testRemoteFiles := []testType{
		// A complete and valid REMOTE yml, except for publiccode-yaml-version instead of
		{"https://raw.githubusercontent.com/gith002/Medusa/master/publiccode.yml", "publiccode-yaml-version : String"},
	}

	for _, test := range testRemoteFiles {
		t.Run(test.errkey, func(t *testing.T) {
			// Parse data into pc struct.
			p := NewParser()
			err := p.ParseRemoteFile(test.file)

			checkParseErrors(t, err, test)
		})
	}
}

func checkParseErrors(t *testing.T, err error, test testType) {
	if test.errkey == "" && err != nil {
		t.Errorf("[%s] unexpected error: %v\n", test.file, err)
	} else if test.errkey != "" && err == nil {
		t.Errorf("[%s] no error generated\n", test.file)
	} else if test.errkey != "" && err != nil {
		if multi, ok := err.(ErrorParseMulti); !ok {
			panic(err)
		} else if len(multi) != 1 {
			t.Errorf("[%s] too many errors generated; 1 was expected but got:\n", test.file)
			for _, e := range multi {
				t.Errorf("  * %s\n", e)
			}
		} else if e, ok := multi[0].(ErrorInvalidValue); ok && (e.Key != test.errkey) {
			t.Errorf("[%s] wrong error generated: %s - key: %#v - instead of %s", test.file, e, e.Key, test.errkey)
		} else if e, ok := multi[0].(ErrorInvalidKey); ok && (e.Key != test.errkey) {
			t.Errorf("[%s] wrong error generated: %s - key: %#v - instead of %s", test.file, e, e.Key, test.errkey)
		}
	}
}

// Test that relative paths are turned into absolute paths.
func TestRelativePaths(t *testing.T) {
	// Parse file into pc struct.
	RemoteBaseURL = "https://raw.githubusercontent.com/italia/18app/master"
	const url = "https://raw.githubusercontent.com/italia/18app/master/publiccode.yml"
	p := NewParser()
	err := p.ParseRemoteFile(url)
	if err != nil {
		t.Errorf("Failed to parse remote file from %v: %v", url, err)
	}

	if strings.Index(p.PublicCode.Description["ita"].Screenshots[0], RemoteBaseURL) != 0 {
		t.Errorf("Relative path was not turned into absolute URL: %v", p.PublicCode.Description["ita"].Screenshots[0])
	}
}
