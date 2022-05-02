package validators;

import (
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

func New() *validator.Validate {
	validate := validator.New()
	validate.RegisterValidation("date", isDate)
	validate.RegisterValidation("iso3166_1_alpha2_lowercase", isIso3166Alpha2Lowercase)
	validate.RegisterValidation("umax", uMax)
	validate.RegisterValidation("umin", uMin)

	validate.RegisterValidation("is_category_v0_2", isCategory_v0_2)
	validate.RegisterValidation("is_scope_v0_2", isScope_v0_2)

	validate.RegisterValidation("is_italian_ipa_code", isItalianIpaCode)

	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("yaml"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
	return validate
}
