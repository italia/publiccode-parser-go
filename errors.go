package publiccode

import (
	"encoding/json"
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
	Key         string `json:"key"`
	Description string `json:"description"`
	Line        int    `json:"line"`
	Column      int    `json:"column"`
}

func (e ValidationError) Error() string {
	key := ""
	if e.Key != "" {
		key = fmt.Sprintf("%s: ", e.Key)
	}

	return fmt.Sprintf("publiccode.yml:%d:%d: error: %s%s", e.Line, e.Column, key, e.Description)
}

func (e ValidationError) MarshalJSON() ([]byte, error) {
	type Ve ValidationError

	return json.Marshal(&struct {
		*Ve
		Type string `json:"type"`
	}{
		Ve:   (*Ve)(&e),
		Type: "error",
	})
}

func newValidationError(key string, description string, args ...interface{}) ValidationError {
	return ValidationError{Key: key, Description: fmt.Sprintf(description, args...)}
}

type ValidationWarning ValidationError

func (e ValidationWarning) Error() string {
	key := ""
	if e.Key != "" {
		key = fmt.Sprintf("%s: ", e.Key)
	}

	return fmt.Sprintf("publiccode.yml:%d:%d: warning: %s%s", e.Line, e.Column, key, e.Description)
}

func (e ValidationWarning) MarshalJSON() ([]byte, error) {
	type Ve ValidationError

	return json.Marshal(&struct {
		*Ve
		Type string `json:"type"`
	}{
		Ve:   (*Ve)(&e),
		Type: "warning",
	})
}

type ValidationResults []error

func (vr ValidationResults) Error() string {
	var s []string
	for _, e := range vr {
		s = append(s, e.Error())
	}
	return strings.Join(s, "\n")
}
