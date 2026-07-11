package validators

import "github.com/go-playground/validator/v10"

// isScopeV1 validates a v1 intendedAudience scope. It is the v0 scope
// list without "government", removed in v1.
func isScopeV1(fl validator.FieldLevel) bool {
	s := fl.Field().String()
	if s == "government" {
		return false
	}

	_, ok := supportedScopesV0[s]

	return ok
}
