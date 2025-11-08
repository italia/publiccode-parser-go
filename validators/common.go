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

func isIso3166Alpha2LowerOrUpper(fl validator.FieldLevel) bool {
	str := fl.Field().String()
	validate := validator.New(validator.WithRequiredStructEnabled())

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
	validate := validator.New(validator.WithRequiredStructEnabled())

	if stringer, ok := fl.Field().Interface().(fmt.Stringer); ok {
		err := validate.Var(stringer.String(), "http_url")

		return err == nil
	}

	//nolint:forbidigo // If we hit this, it's a programming error caught at runtime, it's good to panic.
	panic(fmt.Sprintf("Bad field type for %T. Must be implement fmt.Stringer", fl.Field().Interface()))
}

// go-playground/validator doesn't support `url` validations on non
// strings yet.
func isURL(fl validator.FieldLevel) bool {
	validate := validator.New(validator.WithRequiredStructEnabled())

	if stringer, ok := fl.Field().Interface().(fmt.Stringer); ok {
		err := validate.Var(stringer.String(), "url")

		return err == nil
	}

	//nolint:forbidigo // If we hit this, it's a programming error caught at runtime, it's good to panic.
	panic(fmt.Sprintf("Bad field type for %T. Must be implement fmt.Stringer", fl.Field().Interface()))
}

func isOrganisationURI(fl validator.FieldLevel) bool {
	validate := New()

	field := fl.Field().String()

	u, err := url.ParseRequestURI(field)
	if err != nil {
		return false
	}

	// Validate URNs as well
	if strings.EqualFold(u.Scheme, "urn") {
		err := validate.Var(field, "urn_rfc2141")
		if err != nil {
			return false
		}

		if strings.HasPrefix(u.Opaque, "x-italian-pa:") {
			ipa := strings.ReplaceAll(u.Opaque, "x-italian-pa:", "")
			err := validate.Var(ipa, "is_italian_ipa_code")

			return err == nil
		}

		return true
	}

	if u.Scheme == "" || u.Host == "" {
		return false
	}

	return true
}

// Custom validator to work around https://github.com/go-playground/validator/issues/1260
func bcp47_keys(fl validator.FieldLevel) bool {
	validate := validator.New(validator.WithRequiredStructEnabled())

	if fl.Field().Kind() != reflect.Map {
		//nolint:forbidigo // If we hit this, it's a programming error caught at runtime, it's good to panic.
		panic(fmt.Sprintf("Bad field type for %T. Must be a map", fl.Field().Interface()))
	}

	for _, k := range fl.Field().MapKeys() {
		if err := validate.Var(k.String(), "bcp47_strict_language_tag"); err != nil {
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
	// Reference: https://github.com/jshttp/media-typer/
	re := regexp.MustCompile("^ *([A-Za-z0-9][A-Za-z0-9!#$&^_-]{0,126})/([A-Za-z0-9][A-Za-z0-9!#$&^_.+-]{0,126}) *$")

	return re.MatchString(fl.Field().String())
}
