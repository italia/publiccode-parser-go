package validators

import (
	"testing"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
)

func TestBCP47KeysValidMap(t *testing.T) {
	v := New()

	type S struct {
		M map[string]string `validate:"bcp47_keys"`
	}

	err := v.Struct(S{M: map[string]string{"en": "English", "it": "Italian"}})
	if err != nil {
		t.Errorf("unexpected error for valid BCP47 keys: %v", err)
	}
}

func TestBCP47KeysInvalidMap(t *testing.T) {
	v := New()

	type S struct {
		M map[string]string `validate:"bcp47_keys"`
	}

	err := v.Struct(S{M: map[string]string{"not-valid-lang-12345": "bad"}})
	if err == nil {
		t.Error("expected error for invalid BCP47 key")
	}
}

func TestIsHTTPURLValid(t *testing.T) {
	v := New()

	type S struct {
		U *testURL `validate:"omitnil,url_http_url"`
	}

	u := testURL("https://example.com")
	err := v.Struct(S{U: &u})
	if err != nil {
		t.Errorf("unexpected error for valid HTTP URL: %v", err)
	}
}

func TestIsHTTPURLInvalid(t *testing.T) {
	v := New()

	type S struct {
		U *testURL `validate:"omitnil,url_http_url"`
	}

	u := testURL("not-an-http-url")
	err := v.Struct(S{U: &u})
	if err == nil {
		t.Error("expected error for invalid HTTP URL")
	}
}

func TestIsURLValid(t *testing.T) {
	v := New()

	type S struct {
		U *testURL `validate:"omitnil,url_url"`
	}

	u := testURL("https://example.com/path")
	err := v.Struct(S{U: &u})
	if err != nil {
		t.Errorf("unexpected error for valid URL: %v", err)
	}
}

func TestIsURLInvalid(t *testing.T) {
	v := New()

	type S struct {
		U *testURL `validate:"omitnil,url_url"`
	}

	u := testURL("not-a-valid-url")
	err := v.Struct(S{U: &u})
	if err == nil {
		t.Error("expected error for invalid URL")
	}
}

// testURL is a type that implements fmt.Stringer for testing url validators.
type testURL string

func (u testURL) String() string {
	return string(u)
}

// Ensure testURL satisfies the Stringer interface (used by validators).
var _ interface{ String() string } = testURL("")

func newTranslator(t *testing.T) ut.Translator {
	t.Helper()
	enLocale := en.New()
	uni := ut.New(enLocale, enLocale)
	trans, _ := uni.GetTranslator("en")
	return trans
}

func TestRegistrationFuncPanic(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Error("expected panic from registrationFunc when tag conflicts")
		}
	}()

	trans := newTranslator(t)

	// Register the tag once.
	regFunc1 := registrationFunc("test_panic_tag", "first translation", false)
	if err := regFunc1(trans); err != nil {
		t.Fatalf("first registration failed: %v", err)
	}

	// Register same tag again with override=false — should panic.
	regFunc2 := registrationFunc("test_panic_tag", "second translation", false)
	if err := regFunc2(trans); err != nil {
		// If it returns an error instead of panicking, that's also acceptable.
		_ = err
	}
}

func TestIsHTTPURLPanicNonStringer(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for non-Stringer field with url_http_url")
		}
	}()

	v := New()

	type S struct {
		U int `validate:"url_http_url"`
	}

	_ = v.Struct(S{U: 42})
}

func TestIsURLPanicNonStringer(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for non-Stringer field with url_url")
		}
	}()

	v := New()

	type S struct {
		U int `validate:"url_url"`
	}

	_ = v.Struct(S{U: 42})
}

func TestBCP47KeysPanicNonMap(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for non-map field with bcp47_keys")
		}
	}()

	v := New()

	type S struct {
		M int `validate:"bcp47_keys"`
	}

	_ = v.Struct(S{M: 42})
}

func TestTranslateFuncErrorFallback(t *testing.T) {
	// translateFunc falls back to fe.Error() when ut.T returns an error.
	// This happens when the tag is not registered in the translator.
	v := validator.New(validator.WithRequiredStructEnabled())
	_ = v.RegisterValidation("unregistered_trans_tag", func(fl validator.FieldLevel) bool {
		return false
	})

	trans := newTranslator(t)

	// Register the translation with translateFunc but DO NOT add the tag to the translator.
	_ = v.RegisterTranslation("unregistered_trans_tag", trans,
		func(u ut.Translator) error {
			// Don't add the tag — translateFunc will get an error from ut.T.
			return nil
		},
		translateFunc,
	)

	type S struct {
		V string `validate:"unregistered_trans_tag"`
	}
	err := v.Struct(S{V: "anything"})
	if err == nil {
		t.Fatal("expected validation error")
	}

	// Translate the error — this calls translateFunc.
	if ve, ok := err.(validator.ValidationErrors); ok {
		for _, fe := range ve {
			// This calls translateFunc, which should fallback to fe.Error()
			// when the translation key is not found.
			result := fe.Translate(trans)
			if result == "" {
				t.Error("expected non-empty translated message")
			}
		}
	}
}
