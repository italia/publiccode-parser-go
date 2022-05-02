package publiccode

import funk "github.com/thoas/go-funk"

func (p *Parser) isScope(scope string) (bool) {
	return funk.Contains(supportedScopes, scope)
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
