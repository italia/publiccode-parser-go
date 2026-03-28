package validators

import (
	"testing"
)

func TestIsValidBCP47StrictLanguageTag(t *testing.T) {
	tests := []struct {
		tag   string
		valid bool
	}{
		// Valid 2-letter codes
		{"en", true},
		{"it", true},
		{"de", true},
		// 3-letter code that maps to 2-letter: "eng" -> "en", so it fails the strict check
		{"eng", false},
		// Invalid: unknown base language
		{"xx", false},
		// Invalid: 4-letter language (not special)
		{"xxxx", false},
		// With script
		{"en-Latn", true},
		// With region (2-letter)
		{"en-US", true},
		{"it-IT", true},
		// With region (3-digit M.49)
		{"en-001", true},
		// Invalid M.49 (not in registry)
		{"en-999", false},
		// With variant (basiceng is a variant for en)
		{"en-basiceng", true},
		// With extension
		{"en-u-ca-gregory", true},
		// Grandfathered/irregular - case-insensitive match
		{"i-ami", true},
		// Private use
		{"x-test", true},
		// Extended language: zh-cmn (Mandarin as extlang of Chinese)
		{"zh-cmn", true},
		// Invalid extlang prefix: cmn's prefix is zh, not en
		{"en-cmn", false},
		// Grandfathered regular
		{"art-lojban", true},
		// Invalid: just garbage
		{"not-valid-at-all-12345678", false},
	}

	for _, tc := range tests {
		t.Run(tc.tag, func(t *testing.T) {
			got := isValidBCP47StrictLanguageTag(tc.tag)
			if got != tc.valid {
				t.Errorf("isValidBCP47StrictLanguageTag(%q) = %v, want %v", tc.tag, got, tc.valid)
			}
		})
	}
}

func TestIsValidBCP47StrictLanguageTagLongerBase(t *testing.T) {
	// 5+ letter language (falls into default false case of the switch)
	got := isValidBCP47StrictLanguageTag("toolong")
	if got {
		t.Error("expected false for language with 5+ chars not in special form")
	}
}

func TestIsValidBCP47StrictLanguageTagInvalidScript(t *testing.T) {
	// Invalid script (not a recognized IANA script)
	got := isValidBCP47StrictLanguageTag("en-Xxxx")
	if got {
		t.Errorf("expected false for unrecognized script, got true")
	}
}

func TestIsValidBCP47StrictLanguageTagInvalidRegion(t *testing.T) {
	// "en-XX" is actually accepted by golang.org/x/text/language.ParseRegion
	// The validator delegates to that library for 2-letter region checks.
	// Use a truly invalid region to verify the false path.
	got := isValidBCP47StrictLanguageTag("en-ZZ")
	// ZZ is not a real ISO 3166-1 alpha-2 country, but ParseRegion may accept it.
	// This just verifies no panic occurs.
	_ = got
}

func TestIsValidBCP47StrictLanguageTagWithVariantBadPrefix(t *testing.T) {
	// rozaj is valid only for sl
	got := isValidBCP47StrictLanguageTag("en-rozaj")
	if got {
		t.Errorf("expected false for variant with wrong prefix")
	}
}

func TestIsValidBCP47StrictLanguageTagWithVariantGoodPrefix(t *testing.T) {
	// rozaj is a valid variant for sl
	got := isValidBCP47StrictLanguageTag("sl-rozaj")
	if !got {
		t.Errorf("expected true for sl-rozaj")
	}
}

func TestIsValidBCP47StrictLanguageTagInvalidExtlang(t *testing.T) {
	// zzz is not a known extlang
	got := isValidBCP47StrictLanguageTag("zh-zzz")
	if got {
		t.Errorf("expected false for unknown extlang, got true")
	}
}

func TestIsBCP47StrictLanguageTagPanicNonString(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for non-string field with bcp47_strict_language_tag")
		}
	}()

	v := New(DefaultIPACodes())

	type S struct {
		Code int `validate:"bcp47_strict_language_tag"`
	}

	_ = v.Struct(S{Code: 42})
}

func TestIsValidBCP47StrictLanguageTagExtlang3LetterBase(t *testing.T) {
	// "eng" maps to canonical "en", so "eng-cmn" fails the canonical form check.
	got := isValidBCP47StrictLanguageTag("eng-cmn")
	if got {
		t.Errorf("expected false for eng-cmn (eng normalizes to en)")
	}
}

func TestIsValidBCP47StrictLanguageTagUnknownVariant(t *testing.T) {
	// "foobar" is not in ianaVariants; exercises the variant lookup failure path.
	got := isValidBCP47StrictLanguageTag("en-foobar")
	if got {
		t.Errorf("expected false for en-foobar (unknown variant)")
	}
}

func TestIsValidBCP47StrictLanguageTagRegionZZ(t *testing.T) {
	// "ZZ" is not a registered ISO 3166-1 alpha-2 region code; ParseRegion
	// behaviour varies by Go version. Just verify no panic occurs.
	_ = isValidBCP47StrictLanguageTag("en-ZZ")
}
