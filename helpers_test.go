package publiccode

import (
	"log"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

type testType struct {
	file string
	err  error
}

var cwd string

func init() {
	var err error

	cwd, err = os.Getwd()
	if err != nil {
		log.Fatalf("failed to get cwd: %v", err)
	}
}

// Parse the YAML file passed as argument, using the current directory
// as base path.
//
// Return nil if the parsing succeded or an error if it failed.
func parse(file string) error {
	var p *Parser
	var err error

	if p, err = NewDefaultParser(); err != nil {
		return err
	}

	_, err = p.Parse(file)

	return err
}

func parseNoNetwork(file string) error {
	var p *Parser
	var err error

	if p, err = NewParser(ParserConfig{DisableNetwork: true}); err != nil {
		return err
	}

	_, err = p.Parse(file)

	return err
}

// Check all the YAML files matching the glob pattern and fail for each file
// with parsing or validation errors.
func checkValidFiles(pattern string, t *testing.T) {
	testFiles, _ := filepath.Glob(pattern)
	for _, file := range testFiles {
		t.Run(file, func(t *testing.T) {
			if err := parse(file); err != nil {
				t.Errorf("[%s] validation failed for valid file: %T - %s\n", file, err, err)
			}
		})
	}
}

// Check all the YAML files matching the glob pattern and fail for each file
// with parsing or validation errors, with the network disabled.
func checkValidFilesNoNetwork(pattern string, t *testing.T) {
	testFiles, _ := filepath.Glob(pattern)
	for _, file := range testFiles {
		t.Run(file, func(t *testing.T) {
			if err := parseNoNetwork(file); err != nil {
				t.Errorf("[%s] validation failed for valid file: %T - %s\n", file, err, err)
			}
		})
	}
}

func checkParseErrors(t *testing.T, err error, test testType) {
	if test.err == nil && err != nil {
		t.Errorf("[%s] unexpected error: %v\n", test.file, err)
	} else if test.err != nil && err == nil {
		t.Errorf("[%s] no error generated\n", test.file)
	} else if test.err != nil && err != nil {
		if !reflect.DeepEqual(test.err, err) {
			t.Errorf("[%s] wrong error generated:\n%T - %s\n- instead of:\n%T - %s", test.file, err, err, test.err, test.err)
		}
	}
}
