package validators;

import (
	"strconv"
	"time"

	"github.com/rivo/uniseg"
	"github.com/go-playground/validator/v10"
	"golang.org/x/text/language"
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

func isBcp47(fl validator.FieldLevel) bool {
	str := fl.Field().String()

	_, err := language.Parse(str)

	return err == nil
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
