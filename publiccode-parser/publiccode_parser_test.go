package main

import (
	"errors"
	"testing"

	publiccode "github.com/italia/publiccode-parser-go/v5"
)

func TestHasValidationErrorsNil(t *testing.T) {
	if hasValidationErrors(nil) {
		t.Error("expected false for nil")
	}
}

func TestHasValidationErrorsOnlyWarnings(t *testing.T) {
	vr := publiccode.ValidationResults{
		publiccode.ValidationWarning{Key: "foo", Description: "minor warning"},
	}
	if hasValidationErrors(vr) {
		t.Error("expected false for ValidationResults with only warnings")
	}
}

func TestHasValidationErrorsWithError(t *testing.T) {
	vr := publiccode.ValidationResults{
		publiccode.ValidationError{Key: "foo", Description: "bad field"},
	}
	if !hasValidationErrors(vr) {
		t.Error("expected true for ValidationResults with a ValidationError")
	}
}

func TestHasValidationErrorsPlainError(t *testing.T) {
	if !hasValidationErrors(errors.New("some error")) {
		t.Error("expected true for a plain error")
	}
}

func TestHasValidationErrorsMixedResults(t *testing.T) {
	vr := publiccode.ValidationResults{
		publiccode.ValidationWarning{Key: "foo", Description: "warning"},
		publiccode.ValidationError{Key: "bar", Description: "error"},
	}
	if !hasValidationErrors(vr) {
		t.Error("expected true for mixed results containing an error")
	}
}
