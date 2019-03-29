package publiccode

import (
	"golang.org/x/text/language"
)

// checkLanguageCode tells whether the supplied language code is valid according to BCP47 or not and returns it in canolical form.
func (p *Parser) checkLanguageCode(key, code string) (string, error) {
	if !p.Strict {
		if code == "italian" {
			code = "it"
		}
	}

	// language.Parse() also accepts three-letter languages so it's highly tolerant
	t, err := language.Parse(code)
	if err != nil {
		return "", newErrorInvalidValue(key, "unknown ISO 3166-1 alpha-3 language code: %s", code)
	}

	return t.String(), nil
}
