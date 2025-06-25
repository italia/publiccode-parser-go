package validators

import (
	"fmt"
	"log"
	"reflect"
	"strings"

	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
)

func New() *validator.Validate {
	validate := validator.New(validator.WithRequiredStructEnabled())
	_ = validate.RegisterValidation("date", isDate)
	_ = validate.RegisterValidation("is_mime_type", isMIMEType)
	_ = validate.RegisterValidation("iso3166_1_alpha2_lowercase", isIso3166Alpha2Lowercase)
	_ = validate.RegisterValidation("umax", uMax)
	_ = validate.RegisterValidation("umin", uMin)
	_ = validate.RegisterValidation("url_http_url", isHTTPURL)
	_ = validate.RegisterValidation("url_url", isURL)
	_ = validate.RegisterValidation("is_spdx_expression", isSPDXExpression)

	_ = validate.RegisterValidation("is_category_v0", isCategoryV0)
	_ = validate.RegisterValidation("is_scope_v0", isScopeV0)

	_ = validate.RegisterValidation("is_italian_ipa_code", isItalianIpaCode)

	_ = validate.RegisterValidation("bcp47_keys", bcp47_keys)

	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("yaml"), ",", 2)[0]
		if name == "-" {
			return ""
		}

		return name
	})

	return validate
}

func RegisterLocalErrorMessages(v *validator.Validate, trans ut.Translator) error {
	var err error

	translations := []struct {
		tag             string
		translation     string
		override        bool
		customRegisFunc validator.RegisterTranslationsFunc
		customTransFunc validator.TranslationFunc
	}{
		{
			// Override the default error with a more user friendly one
			//
			// original:
			//   foo must be one of [foo bar contract community none]
			// overridden:
			//   foo must be one of the following: "foo", "bar" or "baz"
			tag:         "oneof",
			translation: "{0} must be one of the following: {1}",
			customTransFunc: func(ut ut.Translator, fe validator.FieldError) string {
				parts := strings.Fields(fe.Param())
				s := `"` + strings.Join(parts, `", "`) + `"`

				if len(parts) > 1 {
					i := strings.LastIndex(s, `", "`)
					s = s[:i] + `" or "` + s[i+4:]
				}

				s, _ = ut.T(fe.Tag(), fe.Field(), s)

				return s
			},
			override: true,
		},
		{
			tag:         "date",
			translation: "{0} must be a date with format 'YYYY-MM-DD'",
		},
		{
			// Override the default error with a more user friendly one
			//
			// original:
			//   foo is a required field
			// overridden:
			//   foo is a required field when "bar" is "foobar"
			tag: "required_if",
			customRegisFunc: func(ut ut.Translator) error {
				return ut.Add("required_if", "{0} is a required field when {1}", true)
			},
			customTransFunc: func(ut ut.Translator, fe validator.FieldError) string {
				parts := strings.Fields(fe.Param())

				t, _ := ut.T("required_if", fe.Field(), fmt.Sprintf("\"%s\" is \"%s\"", strings.ToLower(parts[0]), parts[1]))

				return t
			},
			override: true,
		},
		{
			tag:         "is_mime_type",
			translation: "{0} is not a valid MIME type",
		},
		{
			// Override the default error with a more user friendly one
			//
			// original:
			//   foo is an excluded field
			// overridden:
			//   foo is not permitted when "bar" is "foobar"
			tag: "excluded_unless",
			customRegisFunc: func(ut ut.Translator) error {
				return ut.Add("excluded_unless", "{0} must not be present unless {1}", true)
			},
			customTransFunc: func(ut ut.Translator, fe validator.FieldError) string {
				parts := strings.Fields(fe.Param())

				t, _ := ut.T("excluded_unless", fe.Field(), fmt.Sprintf("\"%s\" is \"%s\"", strings.ToLower(parts[0]), parts[1]))

				return t
			},
			override: true,
		},
		{
			tag: "umax",
			customRegisFunc: func(ut ut.Translator) error {
				return ut.Add("umax", "{0} must be a maximum of {1} characters in length", false)
			},
			customTransFunc: func(ut ut.Translator, fe validator.FieldError) string {
				t, _ := ut.T("umax", fe.Field(), fe.Param())

				return t
			},
		},
		{
			tag: "umin",
			customRegisFunc: func(ut ut.Translator) error {
				return ut.Add("umin", "{0} must be at least {1} characters in length", false)
			},
			customTransFunc: func(ut ut.Translator, fe validator.FieldError) string {
				t, _ := ut.T("umin", fe.Field(), fe.Param())

				return t
			},
		},
		{
			tag:         "url_http_url",
			translation: "{0} must be an HTTP URL",
		},
		{
			tag:         "url_url",
			translation: "{0} must be a valid URL",
		},
		{
			tag:         "is_spdx_expression",
			translation: "{0} must be a valid license (see https://spdx.org/licenses)",
		},
		{
			tag:         "is_category_v0",
			translation: "{0} must be a valid category (see https://github.com/publiccodeyml/publiccode.yml/blob/main/docs/standard/categories-list.rst)",
		},
		{
			tag:         "is_scope_v0",
			translation: "{0} must be a valid scope (see https://github.com/publiccodeyml/publiccode.yml/blob/main/docs/standard/scope-list.rst)",
		},
		{
			tag:         "is_italian_ipa_code",
			translation: "{0} must be a valid Italian Public Administration Code (iPA) (see https://www.indicepa.gov.it/public-services/opendata-read-service.php?dstype=FS&filename=amministrazioni.txt)",
		},
		{
			tag:         "iso3166_1_alpha2_lowercase",
			translation: "{0} must be a valid lowercase ISO 3166-1 alpha-2 two-letter country code",
		},
		{
			tag:         "bcp47_language_tag",
			translation: "{0} must be a valid BCP 47 language",
		},
		{
			tag:         "bcp47_keys",
			translation: "{0} must be a valid BCP 47 language",
		},
	}

	for _, t := range translations {
		if t.customTransFunc != nil && t.customRegisFunc != nil {
			err = v.RegisterTranslation(t.tag, trans, t.customRegisFunc, t.customTransFunc)
		} else if t.customTransFunc != nil && t.customRegisFunc == nil {
			err = v.RegisterTranslation(t.tag, trans, registrationFunc(t.tag, t.translation, t.override), t.customTransFunc)
		} else if t.customTransFunc == nil && t.customRegisFunc != nil {
			err = v.RegisterTranslation(t.tag, trans, t.customRegisFunc, translateFunc)
		} else {
			err = v.RegisterTranslation(t.tag, trans, registrationFunc(t.tag, t.translation, t.override), translateFunc)
		}

		if err != nil {
			return err
		}
	}

	return err
}

func registrationFunc(tag string, translation string, override bool) validator.RegisterTranslationsFunc {
	return func(ut ut.Translator) error {
		if err := ut.Add(tag, translation, override); err != nil {
			log.Fatalf("Error %s", err.Error())

			return err
		}

		return nil
	}
}

func translateFunc(ut ut.Translator, fe validator.FieldError) string {
	t, err := ut.T(fe.Tag(), fe.Field())
	if err != nil {
		return fe.(error).Error()
	}

	return t
}
