package publiccode

import (
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"

	httpclient "github.com/italia/httpclient-lib-go"
)

// downloadFile download the file in the path.
func downloadFile(filepath string, url string, headers map[string]string) error {
	// Create the file.
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data from the url.
	resp, err := httpclient.GetURL(url, headers)
	if err != nil {
		return err
	}

	reader := bytes.NewReader(resp.Body)

	// Write the body to file.
	_, err = io.Copy(out, reader)

	return err
}

// Caller is responsible for removing the temporary directory.
func downloadTmpFile(url string, headers map[string]string) (string, error) {
	// Create a temp dir
	tmpdir, err := ioutil.TempDir("", "publiccode.yml-parser-go")
	if err != nil {
		log.Fatal(err)
	}

	// Download the file in the temp dir.
	tmpFile := filepath.Join(tmpdir, path.Base(url))
	err = downloadFile(tmpFile, url, headers)
	if err != nil {
		return "", err
	}

	return tmpFile, nil
}
