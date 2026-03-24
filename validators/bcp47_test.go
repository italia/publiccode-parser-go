package validators

import (
	"testing"
)

func TestIsValidBCP47StrictLanguageTag(t *testing.T) {
	tests := []struct {
		tag   string
		valid bool
	}{
		// 2-char primary subtags (ISO 639-1)
		{"en", true},
		{"it", true},
		{"fr", true},
		{"de", true},

		// 3-char primary subtags with no ISO 639-1 2-char equivalent —
		// these normalize to themselves in golang.org/x/text/language.
		// 3-char codes that do have a 2-char form (e.g. "eng"→"en") are
		// rejected because the validator requires the canonical form.
		{"sgn", true},
		{"tlh", true},
		{"jbo", true},

		// 5-8 char primary subtags (RFC 5646 §2.1 registered language subtag).
		// Were incorrectly rejected before this fix.
		{"abcde", true},    // 5 chars
		{"abcdefg", true},  // 7 chars
		{"abcdefgh", true}, // 8 chars

		// 4-char primary subtag (reserved for future use per RFC 5646 §2.1)
		{"abcd", true},

		// With region subtag
		{"en-US", true},
		{"it-IT", true},
		{"en-GB", true},
		{"zh-CN", true},

		// With script subtag
		{"zh-Hant", true},
		{"zh-Hans", true},
		{"sr-Latn", true},

		// With script and region
		{"zh-Hant-TW", true},
		{"sr-Latn-RS", true},

		// With extlang subtag
		{"zh-cmn", true},

		// Grandfathered irregular tags
		{"i-ami", true},
		{"i-bnn", true},
		{"art-lojban", true},
		{"zh-min", true},

		// Private use
		{"x-private", true},
		{"x-12345678", true},

		// Empty string
		{"", false},

		// POSIX-style (underscore separator)
		{"en_US", false},
		{"en_GB", false},

		// Primary subtag too long (> 8 chars)
		{"abcdefghi", false},

		// Digits in primary subtag position
		{"1234", false},

		// 3-char code with a 2-char canonical form: requires canonical "en"
		{"eng", false},

		// Unknown extlang subtag
		{"en-xyz", false},

		// Invalid region
		{"en-ZZZ", false},
	}

	for _, tt := range tests {
		t.Run(tt.tag, func(t *testing.T) {
			got := isValidBCP47StrictLanguageTag(tt.tag)
			if got != tt.valid {
				t.Errorf("isValidBCP47StrictLanguageTag(%q) = %v, want %v", tt.tag, got, tt.valid)
			}
		})
	}
}
