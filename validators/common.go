package validators

import (
	"fmt"
	"strconv"
	"time"

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
	err = validate.Var(str, "iso3166_1_alpha2")

	return err != nil
}

func uMax(fl validator.FieldLevel) bool {
	length := uniseg.GraphemeClusterCount(fl.Field().String())
	max, _ := strconv.Atoi(fl.Param())

	return length <= max
}

func uMin(fl validator.FieldLevel) bool {
	length := uniseg.GraphemeClusterCount(fl.Field().String())
	min , _ := strconv.Atoi(fl.Param())

	return length >= min
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
