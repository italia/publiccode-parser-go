package validators

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/github/go-spdx/v2/spdxexp"
	"github.com/go-playground/validator/v10"
	"github.com/rivo/uniseg"
)

func isDate(fl validator.FieldLevel) bool {
	_, err := time.Parse("2006-01-02", fl.Field().String())

	return err == nil
}

func isIso3166Alpha2Lowercase(fl validator.FieldLevel) bool {
	str := fl.Field().String()
	validate := validator.New(validator.WithRequiredStructEnabled())

	err := validate.Var(str, "lowercase")
	if err != nil {
		return false
	}

	err = validate.Var(strings.ToUpper(str), "iso3166_1_alpha2")

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

	panic(fmt.Sprintf("Bad field type for %T. Must be implement fmt.Stringer", fl.Field().Interface()))
}

// Custom validator to work around https://github.com/go-playground/validator/issues/1260
func bcp47_keys(fl validator.FieldLevel) bool {
	validate := validator.New(validator.WithRequiredStructEnabled())

	if fl.Field().Kind() != reflect.Map {
		panic(fmt.Sprintf("Bad field type for %T. Must be a map", fl.Field().Interface()))
	}

	for _, k := range fl.Field().MapKeys() {
		if err := validate.Var(k.String(), "bcp47_language_tag"); err != nil {
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
