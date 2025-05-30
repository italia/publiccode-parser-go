package validators

import (
	"bufio"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/italia/publiccode-parser-go/v4/data"
)

// isCodiceIPA returns true if the field is a valid Italian Public Administration Code
// (iPA) from https://indicepa.gov.it.
func isItalianIpaCode(fl validator.FieldLevel) bool {
	input := data.ItIpaCodes

	// Scan the file, line by line.
	scanner := bufio.NewScanner(strings.NewReader(input))
	for scanner.Scan() {
		// The iPA codes should be validated as case-insensitive, according
		// to the IndicePA guidelines.
		if strings.EqualFold(scanner.Text(), fl.Field().String()) {
			return true
		}
	}

	return false
}
