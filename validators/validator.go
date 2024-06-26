package validators

import (
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

func New() *validator.Validate {
	validate := validator.New(validator.WithRequiredStructEnabled())
	_ = validate.RegisterValidation("date", isDate)
	_ = validate.RegisterValidation("iso3166_1_alpha2_lowercase", isIso3166Alpha2Lowercase)
	_ = validate.RegisterValidation("umax", uMax)
	_ = validate.RegisterValidation("umin", uMin)
	_ = validate.RegisterValidation("url_http_url", isHTTPURL)
	_ = validate.RegisterValidation("url_url", isURL)

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
