package publiccode

import funk "github.com/thoas/go-funk"

// checkScope tells whether the supplied value is a valid scope or not and returns it.
func (p *Parser) checkScope(key, scope string) (string, error) {
	if funk.Contains(supportedScopes, scope) {
		return scope, nil
	}

	return "", newErrorInvalidValue(key, "unknown scope: %s", scope)
}

var supportedScopes = []string{
	"agriculture",
	"culture",
	"defence",
	"education",
	"emergency-services",
	"employment",
	"energy",
	"environment",
	"finance-and-economic-development",
	"foreign-affairs",
	"government",
	"healthcare",
	"infrastructures",
	"justice",
	"local-authorities",
	"manufacturing",
	"research",
	"science-and-technology",
	"security",
	"society",
	"sport",
	"tourism",
	"transportation",
	"welfare",
}
