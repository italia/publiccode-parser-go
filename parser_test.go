package publiccode

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
)

// Test publiccode.yml local files for key errors.
func TestDecodeValueErrors(t *testing.T) {
	BaseDir = ""

	testFiles := []struct {
		file   string
		errkey string
	}{
		// A complete and valid yml
		{"tests/valid.yml", ""}, // Valid yml.
		//
		// // Version
		// {"tests/invalid_version.yml", "publiccode-yaml-version"}, // Invalid version.

		// // Name, ApplicationSuite (no test), URL, LandingURL
		// // Name
		// {"tests/invalid_name_missing.yml", "name"}, // Missing name.
		// // Url
		// {"tests/invalid_url_missing.yml", "url"},     // Missing url.
		// {"tests/invalid_url_schema.yml", "url"},      // Missing schema.
		// {"tests/invalid_url_404notfound.yml", "url"}, // 404 not found.
		// // LandingUrl
		// {"tests/invalid_landingUrl_schema.yml", "landingURL"},      // Missing schema.
		// {"tests/invalid_landingUrl_404notfound.yml", "landingURL"}, // 404 not found.

	}

	for _, test := range testFiles {
		t.Run(test.errkey, func(t *testing.T) {

			// Read data.
			data, err := ioutil.ReadFile(test.file)
			if err != nil {
				fmt.Println(err)
				return
			}

			// Parse data into pc struct.
			var pc PublicCode
			err = Parse(data, &pc)

			//spew.Dump(pc)

			if test.errkey == "" && err != nil {
				t.Error("unexpected error:\n", err)
			} else if test.errkey != "" && err == nil {
				t.Error("error not generated:\n", test.file)
			} else if test.errkey != "" && err != nil {
				if multi, ok := err.(ErrorParseMulti); !ok {
					panic(err)
				} else if len(multi) != 1 {
					t.Errorf("too many errors generated: %#v", multi)
				} else if e, ok := multi[0].(ErrorInvalidValue); !ok || e.Key != test.errkey {
					t.Errorf("wrong error generated: %#v - instead of %s", e.Key, test.errkey)
				}
			}
		})
	}
}

// Test publiccode.yml remote files for key errors.
func TestDecodeValueErrorsRemote(t *testing.T) {
	BaseDir = "https://bitbucket.org/marco-capobussi/publiccode-example/raw/master/"

	testRemoteFiles := []struct {
		file   string
		errkey string
	}{
		// // A complete and valid REMOTE yml
		// {"https://bitbucket.org/marco-capobussi/publiccode-example/raw/master/publiccode.yml", ""}, // Valid remote publiccode.yml.
		//
		// // A complete but invalid REMOTE yml
		// {"https://bitbucket.org/marco-capobussi/publiccode-example/raw/master/publiccode.yml-invalid", "description/logo"}, // Invalid remote publiccode.yml.
	}

	for _, test := range testRemoteFiles {
		t.Run(test.errkey, func(t *testing.T) {

			// Read data.
			resp, err := http.Get(test.file)
			if err != nil {
				fmt.Println(err)
				return
			}
			defer resp.Body.Close()
			data, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				fmt.Println(err)
				return
			}

			// Parse data into pc struct.
			var pc PublicCode
			err = Parse(data, &pc)

			if test.errkey == "" && err != nil {
				t.Error("unexpected error:\n", err)
			} else if test.errkey != "" && err == nil {
				t.Error("error not generated:\n", test.file)
			} else if test.errkey != "" && err != nil {
				if multi, ok := err.(ErrorParseMulti); !ok {
					panic(err)
				} else if len(multi) != 1 {
					t.Errorf("too many errors generated: %#v", multi)
				} else if e, ok := multi[0].(ErrorInvalidValue); !ok || e.Key != test.errkey {
					t.Errorf("wrong error generated: %#v - instead of %s", e, test.errkey)
				}
			}
		})
	}
}
