package publiccode

import (
	"bytes"
	"fmt"
	"net/url"
	"testing"
)

// Test that the exported YAML passes validation again, and that re-exporting it
// matches the first export (lossless roundtrip).
func TestExport(t *testing.T) {
	parser, err := NewParser(ParserConfig{DisableNetwork: true})
	if err != nil {
		t.Errorf("Can't create Parser: %v", err)
	}

	publiccode, err := parser.Parse("testdata/v0/valid/valid.yml")
	if err != nil {
		t.Errorf("Failed to parse valid file: %v", err)
	}

	yaml1, err := publiccode.ToYAML()
	if err != nil {
		t.Errorf("Failed to export YAML: %v", err)
	}

	// var p2 *Parser
	// if p2, err = NewParser("/dev/null"); err != nil {
	// 	t.Errorf("Can't create Parser: %v", err)
	// }

	publiccode2, err := parser.ParseStream(bytes.NewBuffer(yaml1))
	if err != nil {
		t.Errorf("Failed to parse exported file: %v", err)
	}

	yaml2, err := publiccode2.ToYAML()
	if err != nil {
		t.Errorf("Failed to export YAML again: %v", err)
	}

	if !bytes.Equal(yaml1, yaml2) {
		t.Errorf("Exported YAML files do not match; roundtrip is not lossless")
	}
}

// Test the toURL function
func TestToURL(t *testing.T) {
	var err error

	expected := map[string]*url.URL{
		"file.txt":                              &url.URL{Scheme: "file", Path: fmt.Sprintf("%s/file.txt", cwd)},
		"/path/file.txt":                        &url.URL{Scheme: "file", Path: "/path/file.txt"},
		"https://developers.italia.it/":         &url.URL{Scheme: "https", Host: "developers.italia.it", Path: "/"},
		"https://developers.italia.it/file.txt": &url.URL{Scheme: "https", Host: "developers.italia.it", Path: "/file.txt"},
		"http://developers.italia.it/":          &url.URL{Scheme: "http", Host: "developers.italia.it", Path: "/"},
	}

	for in, out := range expected {
		var u *url.URL
		if u, err = toURL(in); err != nil {
			t.Errorf("%s: got error %v", in, err)
		}

		if *u != *out {
			t.Errorf("%s: expected %v got %v", in, out, u)
		}
	}
}
