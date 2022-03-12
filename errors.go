package publiccode

import (
	"fmt"
	"strings"
)

// A generic parse error.
type ParseError struct {
	Reason string
}

func (e ParseError) Error() string {
	return e.Reason
}

type ValidationError struct {
	Key string         `json:"key"`
	Description string `json:"description"`
	Line int           `json:"line"`
	Column int         `json:"column"`
}
func (e ValidationError) Error() string {
	key := ""
	if e.Key != "" {
		key = fmt.Sprintf("%s: ", e.Key)
	}

	return fmt.Sprintf("publiccode.yml:%d:%d: %s%s", e.Line, e.Column, key, e.Description)
}

func newValidationError(key string, description string, args ...interface{}) ValidationError {
	return ValidationError{Key: key, Description: fmt.Sprintf(description, args...)}
}

type ValidationErrors []ValidationError

func (ve ValidationErrors) Error() string {
	var ss []string
	for _, e := range ve {
		ss = append(ss, e.Error())
	}
	return strings.Join(ss, "\n")
}
