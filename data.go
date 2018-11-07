package publiccode

import (
	"io"
	"net/http"
	"os"
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
