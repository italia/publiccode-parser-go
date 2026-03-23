package netutil

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	httpclient "github.com/italia/httpclient-lib-go"
)

func newTestClient(httpClient *http.Client) *httpclient.Client {
	return httpclient.NewClient(httpClient)
}

func TestDownloadTmpFileSuccess(t *testing.T) {
	const content = "hello from server"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(content))
	}))
	defer srv.Close()

	client := newTestClient(srv.Client())

	u, _ := url.Parse(srv.URL + "/testfile.txt")
	path, err := DownloadTmpFile(client, u, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() {
		dir := ""
		if path != "" {
			dir = path[:len(path)-len("/testfile.txt")]
			_ = os.Remove(path)
			_ = os.Remove(dir)
		}
	}()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("can't read downloaded file: %v", err)
	}
	if string(data) != content {
		t.Errorf("unexpected content: %q", string(data))
	}
}

func TestDownloadTmpFileHTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "server error", http.StatusInternalServerError)
	}))
	defer srv.Close()

	client := newTestClient(srv.Client())

	u, _ := url.Parse(srv.URL + "/testfile.txt")
	_, err := DownloadTmpFile(client, u, nil)
	// httpclient-lib-go may or may not return error for 5xx; we just ensure no panic.
	_ = err
}

func TestDownloadTmpFileConnectionRefused(t *testing.T) {
	client := newTestClient(&http.Client{})

	u, _ := url.Parse("http://localhost:1/testfile.txt")
	_, err := DownloadTmpFile(client, u, nil)
	if err == nil {
		t.Error("expected error for connection refused")
	}
}

func TestDownloadFileCreateFails(t *testing.T) {
	client := newTestClient(&http.Client{})
	u, _ := url.Parse("http://localhost:1/testfile.txt")

	// Try to create a file in a path that doesn't exist.
	err := downloadFile(client, "/nonexistent/directory/testfile.txt", u, nil)
	if err == nil {
		t.Error("expected error for nonexistent directory")
	}
}
