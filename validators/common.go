package validators

import (
	"fmt"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/github/go-spdx/v2/spdxexp"
	"github.com/go-playground/validator/v10"
	"github.com/rivo/uniseg"
)

// Reference: https://github.com/jshttp/media-typer/
var reMIMEType = regexp.MustCompile(`^ *([A-Za-z0-9][A-Za-z0-9!#$&^_-]{0,126})/([A-Za-z0-9][A-Za-z0-9!#$&^_.+-]{0,126}) *$`) //nolint:lll // RFC 2045 regex

var sharedValidator = validator.New(validator.WithRequiredStructEnabled())

func isIso3166Alpha2LowerOrUpper(fl validator.FieldLevel) bool {
	str := fl.Field().String()
	validate := sharedValidator

	// Either all lower or all upper, don't allow mixed case
	if str != strings.ToLower(str) && str != strings.ToUpper(str) {
		return false
	}

	err := validate.Var(strings.ToUpper(str), "iso3166_1_alpha2")

	return err == nil
}

func uMax(fl validator.FieldLevel) bool {
	length := uniseg.GraphemeClusterCount(fl.Field().String())
	param, _ := strconv.Atoi(fl.Param())

	return length <= param
}

func uMin(fl validator.FieldLevel) bool {
	length := uniseg.GraphemeClusterCount(fl.Field().String())
	param, _ := strconv.Atoi(fl.Param())

	return length >= param
}

// go-playground/validator doesn't support `http_url` validations on non
// strings yet.
func isHTTPURL(fl validator.FieldLevel) bool {
	validate := sharedValidator

	if stringer, ok := fl.Field().Interface().(fmt.Stringer); ok {
		err := validate.Var(stringer.String(), "http_url")

		return err == nil
	}

	//nolint:forbidigo // If we hit this, it's a programming error caught at runtime, it's good to panic.
	panic(fmt.Sprintf("Bad field type for %T. Must implement fmt.Stringer", fl.Field().Interface()))
}

// go-playground/validator doesn't support `url` validations on non
// strings yet.
func isURL(fl validator.FieldLevel) bool {
	validate := sharedValidator

	if stringer, ok := fl.Field().Interface().(fmt.Stringer); ok {
		err := validate.Var(stringer.String(), "url")

		return err == nil
	}

	//nolint:forbidigo // If we hit this, it's a programming error caught at runtime, it's good to panic.
	panic(fmt.Sprintf("Bad field type for %T. Must implement fmt.Stringer", fl.Field().Interface()))
}

// MakeIsOrganisationURI returns a validator.Func that validates an organisation URI,
// including Italian PA URNs (urn:x-italian-pa:<codiceIPA>), using the provided
// IPA codes set. The inner validator is built once and captured in the closure.
func MakeIsOrganisationURI(codes map[string]struct{}) validator.Func {
	inner := validator.New(validator.WithRequiredStructEnabled())
	_ = inner.RegisterValidation("is_italian_ipa_code", MakeIsItalianIpaCode(codes))

	return func(fl validator.FieldLevel) bool {
		field := fl.Field().String()

		u, err := url.ParseRequestURI(field)
		if err != nil {
			return false
		}

		// Validate URNs as well
		if strings.EqualFold(u.Scheme, "urn") {
			if err := inner.Var(field, "urn_rfc2141"); err != nil {
				return false
			}

			if strings.HasPrefix(strings.ToLower(u.Opaque), "x-italian-pa:") {
				ipa := u.Opaque[len("x-italian-pa:"):]

				return inner.Var(ipa, "is_italian_ipa_code") == nil
			}

			return true
		}

		if u.Scheme == "" || u.Host == "" {
			return false
		}

		return true
	}
}

// Custom validator to work around https://github.com/go-playground/validator/issues/1260
func bcp47_keys(fl validator.FieldLevel) bool {
	if fl.Field().Kind() != reflect.Map {
		//nolint:forbidigo // If we hit this, it's a programming error caught at runtime, it's good to panic.
		panic(fmt.Sprintf("Bad field type for %T. Must be a map", fl.Field().Interface()))
	}

	for _, k := range fl.Field().MapKeys() {
		if !isValidBCP47StrictLanguageTag(k.String()) {
			return false
		}
	}

	return true
}

func isSPDXExpression(fl validator.FieldLevel) bool {
	valid, _ := spdxexp.ValidateLicenses([]string{fl.Field().String()})

	return valid
}

// isMIMEType checks whether the string in input is a well formed MIME type or not.
func isMIMEType(fl validator.FieldLevel) bool {
	return reMIMEType.MatchString(fl.Field().String())
}
