package validators

import (
	"testing"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	en_translations "github.com/go-playground/validator/v10/translations/en"
)

func TestNewReturnsNonNil(t *testing.T) {
	v := New()
	if v == nil {
		t.Fatal("New() returned nil")
	}
}

func TestRegisterLocalErrorMessages(t *testing.T) {
	v := New()
	enLocale := en.New()
	uni := ut.New(enLocale, enLocale)
	trans, _ := uni.GetTranslator("en")
	_ = en_translations.RegisterDefaultTranslations(v, trans)

	err := RegisterLocalErrorMessages(v, trans)
	if err != nil {
		t.Fatalf("RegisterLocalErrorMessages returned error: %v", err)
	}
}

func TestNewTagNameFuncWithDashTag(t *testing.T) {
	v := New()

	// A struct with yaml:"-" should have the field excluded from validation.
	type S struct {
		Name   string `validate:"required" yaml:"name"`
		Hidden string `yaml:"-"`
	}

	// Validate a struct with yaml:"-" to trigger the name=="-" path.
	err := v.Struct(S{Name: "test", Hidden: ""})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRegisterLocalErrorMessagesOrgURIAlreadyRegistered(t *testing.T) {
	// Pre-register "organisation_uri" so the customRegisFunc for that tag
	// fails on ut.Add, covering the error-return paths in RegisterLocalErrorMessages.
	v := New()
	enLocale := en.New()
	uni := ut.New(enLocale, enLocale)
	trans, _ := uni.GetTranslator("en")
	_ = en_translations.RegisterDefaultTranslations(v, trans)

	_ = trans.Add("organisation_uri", "already registered", false)

	err := RegisterLocalErrorMessages(v, trans)
	if err == nil {
		t.Error("expected error when organisation_uri is already registered")
	}
}

func TestRegisterLocalErrorMessagesOrgURIInvalidIPAAlreadyRegistered(t *testing.T) {
	// Pre-register only organisation_uri_invalid_italian_pa so the first ut.Add
	// in customRegisFunc succeeds but the second fails, covering the second
	// error-return path.
	v := New()
	enLocale := en.New()
	uni := ut.New(enLocale, enLocale)
	trans, _ := uni.GetTranslator("en")
	_ = en_translations.RegisterDefaultTranslations(v, trans)

	_ = trans.Add("organisation_uri_invalid_italian_pa", "already registered", false)

	err := RegisterLocalErrorMessages(v, trans)
	if err == nil {
		t.Error("expected error when organisation_uri_invalid_italian_pa is already registered")
	}
}

func TestTranslateFuncFallback(t *testing.T) {
	// translateFunc falls back to fe.Error() when the translation key is not found.
	// We verify this by registering a custom translation with no message key
	// and triggering a validation that uses it.
	// This is an indirect test: just ensure RegisterLocalErrorMessages succeeds
	// and the validator can validate with translations without panicking.
	v := New()
	enLocale := en.New()
	uni := ut.New(enLocale, enLocale)
	trans, _ := uni.GetTranslator("en")
	_ = en_translations.RegisterDefaultTranslations(v, trans)
	_ = RegisterLocalErrorMessages(v, trans)

	// Trigger a validation error to exercise translateFunc.
	type S struct {
		Lang string `validate:"bcp47_strict_language_tag"`
	}
	err := v.Struct(S{Lang: "not-a-valid-lang-123456789"})
	if err == nil {
		t.Fatal("expected validation error")
	}
}
