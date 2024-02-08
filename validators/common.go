package validators;

import (
	"regexp"
	"strconv"
	"time"

	"github.com/rivo/uniseg"
	"github.com/go-playground/validator/v10"
)

func isDate(fl validator.FieldLevel) bool {
	_, err := time.Parse("2006-01-02", fl.Field().String())

	return err == nil
}

func isIso3166Alpha2Lowercase(fl validator.FieldLevel) bool {
	str := fl.Field().String()
	validate := validator.New()

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

// isMIMEType checks whether the string in input is a well formed MIME type or not.
func isMIMEType(fl validator.FieldLevel) bool {
	// Reference: https://github.com/jshttp/media-typer/
	re := regexp.MustCompile("^ *([A-Za-z0-9][A-Za-z0-9!#$&^_-]{0,126})/([A-Za-z0-9][A-Za-z0-9!#$&^_.+-]{0,126}) *$")

	return re.MatchString(fl.Field().String())
}
