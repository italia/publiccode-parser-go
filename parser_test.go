package publiccode

import (
	"testing"
)

func TestDecodeValueErrors(t *testing.T) {
	tests := []struct {
		yaml   string
		errkey string
	}{
		{"version: \"0.1\"\n", ""},
		{"version: \"0.2\"\n", "version"},
		{"url: \"https://example.com/path\"\n", ""},
		{"url: \"git://example.com\"\n", ""},
		{"url: \"example.com/path\"\n", "url"},
		{"upstream-url: \"https://example.com/path\"\n", ""},
		{"upstream-url: \"example.com/path\"\n", "upstream-url"},
		{"upstream-url:\n - \"example.com/path\"\n", "upstream-url"},
		{"upstream-url:\n - \"https://example.com/path\"\n - \"https://example.com/path\"\n", ""},
		{"legal:\n authors-file: \"publiccode.go\"\n", ""},
		{"legal:\n authors-file: \"non-existing-file\"\n", "legal/authors-file"},
		{"legal:\n license: \"GPL-3.0\"\n", ""},
		{"legal:\n license: \"GPL-8.0\"\n", "legal/license"},
		{"maintenance:\n until: \"2004-11-04\"\n", ""},
		{"maintenance:\n until: \"2004-15-04\"\n", "maintenance/until"},
		{"maintenance:\n type: \"community\"\n", ""},
		{"maintenance:\n type: \"commercial\"\n until: \"2004-12-01\"\n", ""},
		{"maintenance:\n type: \"non-existing\"\n", "maintenance/type"},
		{"maintenance:\n type: \"commercial\"\n", "maintenance/until"},
		{"maintenance:\n maintainer: \"Linus Torvalds\"\n", ""},
		{"maintenance:\n maintainer:\n  - \"Linus Torvalds\"\n", ""},
		{"maintenance:\n technical-contacts:\n  - name: \"Linus Torvalds\"\n  - email: \"linus@example.com\"\n", ""},
	}

	for _, test := range tests {
		t.Run(test.errkey, func(t *testing.T) {
			var pc PublicCode
			err := Parse([]byte(test.yaml), &pc)

			if test.errkey == "" && err != nil {
				t.Error("unexpected error:\n", err)
			} else if test.errkey != "" && err == nil {
				t.Error("error not generated:\n", test.yaml)
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
