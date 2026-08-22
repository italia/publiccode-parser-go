package publiccode

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ParseError is generic parse error.
type ParseError struct {
	Reason string
}

func (e ParseError) Error() string {
	return e.Reason
}

// defaultFileName is reported by diagnostics coming from an input with no
// name of its own, as happens when parsing a stream.
const defaultFileName = "publiccode.yml"

type ValidationError struct {
	// Name of the validated file, with no directory part. It is empty when
	// the input is a stream, which has no name.
	File        string `json:"file"`
	Key         string `json:"key"`
	Description string `json:"description"`
	Line        int    `json:"line"`
	Column      int    `json:"column"`
}

func (e ValidationError) Error() string {
	key := ""
	if e.Key != "" {
		key = e.Key + ": "
	}

	return fmt.Sprintf("%s:%d:%d: error: %s%s", e.fileName(), e.Line, e.Column, key, e.Description)
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

func (e ValidationError) fileName() string {
	if e.File == "" {
		return defaultFileName
	}

	return e.File
}

func newValidationError(key string, description string) ValidationError {
	return ValidationError{Key: key, Description: description}
}

func newValidationErrorf(key string, description string, args ...any) ValidationError {
	return newValidationError(key, fmt.Sprintf(description, args...))
}

//nolint:errname,lll // ValidationWarning is intentionally named as a warning, not an error, even though it implements error.
type ValidationWarning ValidationError

func newValidationWarning(key string, description string) ValidationWarning {
	return ValidationWarning{Key: key, Description: description}
}

func newValidationWarningf(key string, description string, args ...any) ValidationWarning {
	return newValidationWarning(key, fmt.Sprintf(description, args...))
}

func (e ValidationWarning) Error() string {
	key := ""
	if e.Key != "" {
		key = e.Key + ": "
	}

	return fmt.Sprintf("%s:%d:%d: warning: %s%s", ValidationError(e).fileName(), e.Line, e.Column, key, e.Description)
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

type ValidationResults []error //nolint:errname // intentionally named as a collection, not an error type

func (vr ValidationResults) Error() string {
	s := make([]string, 0, len(vr))
	for _, e := range vr {
		s = append(s, e.Error())
	}

	return strings.Join(s, "\n")
}
