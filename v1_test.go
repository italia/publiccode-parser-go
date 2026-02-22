package publiccode

import (
	"path/filepath"
	"testing"
)

func TestInvalidTestcasesV1(t *testing.T) {
	dirs := []string{"testdata/v1/invalid/",
		"testdata/v1/invalid_valid_in_v0/",
		"testdata/v1/invalid_valid_with_warnings_in_v0/",
	}

	expected := ValidationResults{
		ValidationError{
			"publiccodeYmlVersion",
			"unsupported version: '1'. Supported versions: 0, 0.2, 0.2.0, 0.2.1, 0.2.2, 0.3, 0.3.0, 0.4, 0.4.0, 0.5.0, 0.5",
			0,
			0,
		},
	}

	for _, dir := range dirs {
		testFiles, _ := filepath.Glob(dir + "*yml")

		for _, file := range testFiles {
			t.Run(file, func(t *testing.T) {
				err := parseNoNetwork(file)
				checkParseErrors(t, err, testType{file, expected})
			})
		}
	}
}
