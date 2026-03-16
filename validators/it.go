package validators

import (
	"bufio"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/italia/publiccode-parser-go/v5/data"
)

// ipaCodes is a set of valid IPA codes, lowercased for case-insensitive lookup.
var ipaCodes map[string]struct{}

func init() {
	scanner := bufio.NewScanner(strings.NewReader(data.ItIpaCodes))
	ipaCodes = make(map[string]struct{}, 24000)

	for scanner.Scan() {
		ipaCodes[strings.ToLower(scanner.Text())] = struct{}{}
	}
}

// isCodiceIPA returns true if the field is a valid Italian Public Administration Code
// (iPA) from https://indicepa.gov.it.
func isItalianIpaCode(fl validator.FieldLevel) bool {
	_, ok := ipaCodes[strings.ToLower(fl.Field().String())]

	return ok
}
