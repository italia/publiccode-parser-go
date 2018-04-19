package publiccode

import (
	"fmt"
	"io/ioutil"
	"testing"
)

func TestDecodeValueErrors(t *testing.T) {
	testFiles := []struct {
		file   string
		errkey string
	}{
		{"tests/valid.yml", ""},
		{"tests/invalid_version.yml", "version"},
		{"tests/invalid_url_schema.yml", "url"},
		{"tests/invalid_url_404notfound.yml", "url"},
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
					t.Errorf("wrong error generated: %#v", err)
				}
			}
		})
	}
}
