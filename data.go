package publiccode

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
)

// downloadFile download the file in the path.
func downloadFile(filepath string, url string) error {
	// Create the file.
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data from the url.
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Write the body to file.
	_, err = io.Copy(out, resp.Body)

	return err
}

// Caller is responsible for removing the temporary directory.
func downloadTmpFile(url string) (string, error) {
	// Create a temp dir
	tmpdir, err := ioutil.TempDir("", "publiccode.yml-parser-go")
	if err != nil {
		log.Fatal(err)
	}

	// Download the file in the temp dir.
	tmpFile := filepath.Join(tmpdir, path.Base(url))
	err = downloadFile(tmpFile, url)
	if err != nil {
		return "", err
	}

	return tmpFile, nil
}
