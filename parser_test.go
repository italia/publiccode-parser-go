package publiccode

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"
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

// TestParseConcurrent verifies that multiple goroutines can call Parse on the same
// Parser instance simultaneously without data races, and that each call uses the
// correct baseURL for its file.
//
// Run with "go test -race"
func TestParseConcurrent(t *testing.T) {
	const goroutines = 40

	parser, err := NewParser(ParserConfig{DisableNetwork: true})
	if err != nil {
		t.Fatalf("can't create parser: %v", err)
	}
	// Has a logo and screenshots that only exist under no-network/assets/img/
	fileA := "testdata/v0/valid/no-network/valid.yml"

	// lives in the parent directory and has no screenshots
	fileB := "testdata/v0/valid/categories_empty.yml"

	_, errA := parser.Parse(fileA)
	_, errB := parser.Parse(fileB)

	var wg sync.WaitGroup

	type result struct {
		file string
		err  error
	}

	results := make([]result, goroutines)

	for i := range goroutines {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()

			file := fileA
			if i%2 == 1 {
				file = fileB
			}

			_, results[i].err = parser.Parse(file)
			results[i].file = file
		}(i)
	}

	wg.Wait()

	for i, r := range results {
		var expected error

		// If base URLs leaked between goroutines, a fileA goroutine
		// receiving fileB's baseURL would fail to locate the screenshots
		// and return spurious validation errors.
		if r.file == fileA {
			expected = errA
		} else {
			expected = errB
		}

		if !reflect.DeepEqual(r.err, expected) {
			t.Errorf("goroutine %d (%s): got %v, want %v", i, r.file, r.err, expected)
		}
	}
}

// TestDefaultTimeout checks that the default HTTP timeout is set when
// ParserConfig.Timeout is zero.
func TestDefaultTimeout(t *testing.T) {
	p, err := NewParser(ParserConfig{})
	if err != nil {
		t.Fatal(err)
	}

	if p.client.Timeout != defaultHTTPTimeout {
		t.Errorf("expected default timeout %v, got %v", defaultHTTPTimeout, p.client.Timeout)
	}
}

// TestParseTimeout checks that Parse() fails when the server exceeds the
// configured timeout while fetching the publiccode.yml itself.
func TestParseTimeout(t *testing.T) {
	fixture, err := os.ReadFile("testdata/v0/valid/valid.yml")
	if err != nil {
		t.Fatal(err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond)
		w.Header().Set("Content-Type", "text/yaml")
		w.Write(fixture) //nolint:errcheck
	}))
	defer srv.Close()

	p, err := NewParser(ParserConfig{Timeout: 1 * time.Millisecond})
	if err != nil {
		t.Fatal(err)
	}

	_, err = p.Parse(srv.URL + "/publiccode.yml")
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}

	if !strings.Contains(err.Error(), "deadline exceeded") && !strings.Contains(err.Error(), "timeout") {
		t.Errorf("expected timeout error, got: %v", err)
	}
}

// TestValidationTimeout checks that external validation checks (logo) fail when
// the server exceeds the configured timeout.
func TestValidationTimeout(t *testing.T) {
	slow := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer slow.Close()

	// Read logo_with_url.yml and redirect its logo to the slow server.
	fixture, err := os.ReadFile("testdata/v0/valid/logo_with_url.yml")
	if err != nil {
		t.Fatal(err)
	}

	yml := bytes.ReplaceAll(fixture,
		[]byte("https://raw.githubusercontent.com/italia/publiccode-parser-go/refs/heads/main/testdata/v0/valid/assets/img/logo.png"),
		[]byte(slow.URL+"/logo.png"),
	)

	fast := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/yaml")
		w.Write(yml) //nolint:errcheck
	}))
	defer fast.Close()

	p, err := NewParser(ParserConfig{Timeout: 1 * time.Millisecond})
	if err != nil {
		t.Fatal(err)
	}

	_, err = p.Parse(fast.URL + "/publiccode.yml")
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}

	if !strings.Contains(err.Error(), "deadline exceeded") && !strings.Contains(err.Error(), "timeout") {
		t.Errorf("expected timeout error, got: %v", err)
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
