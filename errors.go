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

// ErrorInvalidKey represents an error caused by an invalid key.
type ErrorInvalidKey struct {
	Key string
}

func (e ErrorInvalidKey) Error() string {
	return fmt.Sprintf("invalid key: %s", e.Key)
}

// ErrorInvalidValue represents an error caused by an invalid value.
type ErrorInvalidValue struct {
	Key    string
	Reason string
}

type ValidationError struct {
	Reason string
}

func (e ErrorInvalidValue) Error() string {
	return fmt.Sprintf("%s: %s", e.Key, e.Reason)
}

func newErrorInvalidValue(key string, reason string, args ...interface{}) ErrorInvalidValue {
	return ErrorInvalidValue{Key: key, Reason: fmt.Sprintf(reason, args...)}
}

// ErrorParseMulti represents an error caused by a multivalue key.
type ErrorParseMulti []error

func (es ErrorParseMulti) Error() string {
	var ss []string
	for _, e := range es {
		ss = append(ss, e.Error())
	}
	return strings.Join(ss, "\n")
}
