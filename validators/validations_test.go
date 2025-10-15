package validators

import (
	"testing"
)

func TestIsOrganisationURI(t *testing.T) {
	v := New()

	tests := []struct {
		value  string
		expect bool
	}{
		{"", false},
		{"1", false},
		{"http", false},
		{"http:", false},
		{"http://", false},
		{"http://foobar.example", true},
		{"urn", false},
		{"urn:", false},
		{"urn:foobar", false},
		{"urn:foobar:baz", true},
		{"urn:x-italian-pa:no_such_ipa_code", false},
		{"urn:x-italian-pa:pcm", true},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			err := v.Var(tt.value, "organisation_uri")
			got := err == nil
			if got != tt.expect {
				t.Errorf("isOrganisationURI(%q) = %v, want %v", tt.value, got, tt.expect)
			}
		})
	}
}
