package netutil

import (
	"bytes"
	"io"
	"log"
	"net/url"
	"os"
	"path"
	"path/filepath"

	httpclient "github.com/italia/httpclient-lib-go"
)

// downloadFile download the file in the path.
func downloadFile(filepath string, url *url.URL, headers map[string]string) error {
	// Create the file.
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data from the url.
	resp, err := httpclient.GetURL(url.String(), headers)
	if err != nil {
		return err
	}

	reader := bytes.NewReader(resp.Body)

	// Write the body to file.
	_, err = io.Copy(out, reader)

	return err
}

// Caller is responsible for removing the temporary directory.
func DownloadTmpFile(url *url.URL, headers map[string]string) (string, error) {
	// Create a temp dir
	tmpdir, err := os.MkdirTemp("", "publiccode.yml-parser-go")
	if err != nil {
		log.Fatal(err)
	}

	// Download the file in the temp dir.
	tmpFile := filepath.Join(tmpdir, path.Base(url.Path))
	err = downloadFile(tmpFile, url, headers)
	if err != nil {
		return "", err
	}

	return tmpFile, nil
}
