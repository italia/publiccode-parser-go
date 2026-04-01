package validators

import (
	"bufio"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/italia/publiccode-parser-go/v5/data"
)

// DefaultIPACodes parses the embedded IPA codes list and returns it as a set.
func DefaultIPACodes() map[string]struct{} {
	scanner := bufio.NewScanner(strings.NewReader(data.ItIpaCodes))
	codes := make(map[string]struct{}, 24000)

	for scanner.Scan() {
		codes[strings.ToLower(scanner.Text())] = struct{}{}
	}

	return codes
}

// MakeIsItalianIpaCode returns a validator.Func that checks whether a field value
// is a valid Italian Public Administration Code (iPA) from https://indicepa.gov.it,
// using the provided codes set.
func MakeIsItalianIpaCode(codes map[string]struct{}) validator.Func {
	return func(fl validator.FieldLevel) bool {
		_, ok := codes[strings.ToLower(fl.Field().String())]

		return ok
	}
}
