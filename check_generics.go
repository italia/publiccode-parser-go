package publiccode

import (
	"bytes"
	"compress/gzip"
	"image"
	"image/png"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// Despite the spec requires at least 1000px, we temporarily release this constraint to 120px.
const minLogoWidth = 120

// checkEmail tells whether email is well formatted.
// In general an email is valid if compile the regex: ^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$
func (p *parser) checkEmail(key string, fn string) error {
	re := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	if !re.MatchString(fn) {
		return newErrorInvalidValue(key, "invalid email: %v", fn)
	}

	return nil
}

// checkURL tells whether the URL resource is well formatted and reachable and return it as *url.URL.
// An URL resource is well formatted if it's' a valid URL and the scheme is not empty.
// An URL resource is reachable if returns an http Status = 200 OK.
func (p *parser) checkURL(key string, value string) (*url.URL, error) {
	u, err := url.Parse(value)
	if err != nil {
		return nil, newErrorInvalidValue(key, "not a valid URL: %s", value)
	}
	if u.Scheme == "" {
		return nil, newErrorInvalidValue(key, "missing URL scheme: %s", value)
	}
	r, err := http.Get(value)
	if err != nil {
		return nil, newErrorInvalidValue(key, "Http.get failed for: %s", value)
	}
	if r.StatusCode != 200 {
		return nil, newErrorInvalidValue(key, "URL is unreachable: %s", value)
	}

	return u, nil
}

// checkFile tells whether the file resource exists and return it.
func (p *parser) checkFile(key string, fn string) (string, error) {
	if BaseDir == "" {
		// Local.
		if _, err := os.Stat(fn); err != nil {
			return "", newErrorInvalidValue(key, "file does not exist: %v", fn)
		}
	} else {
		// Remote.
		_, err := p.checkURL(key, BaseDir+fn)
		if err != nil {
			return "", newErrorInvalidValue(key, "file does not exist on remote: %v", BaseDir+fn)
		}
	}
	return fn, nil
}

// checkDate tells whether the string in input is a date in the
// format "YYYY-MM-DD", which is one of the ISO8601 allowed encoding, and return it as time.Time.
func (p *parser) checkDate(key string, value string) (time.Time, error) {
	t, err := time.Parse("2006-01-02", value)
	if err != nil {
		return t, newErrorInvalidValue(key, "cannot parse date: %v", err)
	}
	return t, nil
}

// checkImage tells whether the string in a valid image. It also checks if the file exists.
// Reference: https://github.com/publiccodenet/publiccode.yml/blob/develop/schema.md
func (p *parser) checkImage(key string, value string) (string, error) {
	validExt := []string{".jpg", ".png"}
	ext := strings.ToLower(filepath.Ext(value))

	// Check for valid extension.
	if !contains(validExt, ext) {
		return value, newErrorInvalidValue(key, "invalid file extension for: %s", value)
	}

	// Check existence of file.
	file, err := p.checkFile(key, value)

	return file, err
}

// checkLogo tells whether the string in a valid logo. It also checks if the file exists.
// Reference: https://github.com/publiccodenet/publiccode.yml/blob/develop/schema.md
func (p *parser) checkLogo(key string, value string) (string, error) {
	validExt := []string{".svg", ".svgz", ".png"}
	ext := strings.ToLower(filepath.Ext(value))

	// Check for valid extension.
	if !contains(validExt, ext) {
		return value, newErrorInvalidValue(key, "invalid file extension for: %s", value)
	}

	// Check existence of file.
	file, err := p.checkFile(key, value)
	if err != nil {
		return value, err
	}

	// Remote. Create a temp dir, download and check the file. Remove the temp dir.
	if BaseDir != "" {
		// Create a temp dir and delete after use.
		dir, err := ioutil.TempDir("", "publiccode.yml-parser-go")
		if err != nil {
			log.Fatal(err)
		}
		defer os.RemoveAll(dir)
		// Download the file in the temp dir.
		fileName := filepath.Base(value)
		tmpFile := filepath.Join(dir, fileName)
		err = downloadFile(tmpFile, BaseDir+value)
		if err != nil {
			return file, err
		}
		// Update file.
		value = tmpFile
	}

	// Check for image size if .png.
	if ext == ".png" {
		f, err := os.Open(value)
		if err != nil {
			return file, err
		}
		image, _, err := image.DecodeConfig(f)
		if err != nil {
			return file, err
		}

		if image.Width < minLogoWidth {
			return value, newErrorInvalidValue(key, "invalid image size of %d (min %dpx of width): %s", image.Width, minLogoWidth, value)
		}
	}
	return file, nil
}

// checkLogo tells whether the string in a valid logo. It also checks if the file exists.
// Reference: https://github.com/publiccodenet/publiccode.yml/blob/develop/schema.md
func (p *parser) checkMonochromeLogo(key string, value string) (string, error) {
	validExt := []string{".svg", ".svgz", ".png"}
	ext := strings.ToLower(filepath.Ext(value))

	// Check for valid extension.
	if !contains(validExt, ext) {
		return value, newErrorInvalidValue(key, "invalid file extension for: %s", value)
	}

	// Check existence of file.
	file, err := p.checkFile(key, value)
	if err != nil {
		return value, err
	}

	// Remote. Create a temp dir, download and check the file. Remove the temp dir.
	if BaseDir != "" {
		// Create a temp dir and delete after use.
		dir, err := ioutil.TempDir("", "publiccode.yml-parser-go")
		if err != nil {
			log.Fatal(err)
		}
		defer os.RemoveAll(dir)
		// Download the file in the temp dir.
		fileName := filepath.Base(value)
		tmpFile := filepath.Join(dir, fileName)
		err = downloadFile(tmpFile, BaseDir+value)
		if err != nil {
			return file, err
		}
		// Update file.
		value = tmpFile
	}

	// Check for image size if .png.
	if ext == ".png" {
		image.RegisterFormat("png", "png", png.Decode, png.DecodeConfig)

		f, err := os.Open(value)
		if err != nil {
			return file, err
		}
		defer f.Close()

		imgCfg, _, err := image.DecodeConfig(f)
		if err != nil {
			return file, err
		}
		width := imgCfg.Width
		height := imgCfg.Height

		if width < minLogoWidth {
			return value, newErrorInvalidValue(key, "invalid image size of %d (min %dpx of width): %s", width, minLogoWidth, value)
		}

		// Check if monochrome (black). Pixel by pixel.
		f.Seek(0, 0)
		img, _, err := image.Decode(f)
		if err != nil {
			return file, err
		}
		for y := 0; y < width; y++ {
			for x := 0; x < height; x++ {
				r, g, b, _ := img.At(x, y).RGBA()
				if r != 0 || g != 0 || b != 0 {
					return file, newErrorInvalidValue(key, "the monochromeLogo is not monochrome (black): %s", value)
				}
			}
		}
	} else if ext == ".svg" {
		// Regex for every hex color.
		re := regexp.MustCompile("#(?:[0-9a-fA-F]{3}){1,2}")

		// Read file data.
		data, err := ioutil.ReadFile(value)
		if err != nil {
			return file, err
		}

		for _, color := range re.FindAllString(string(data), -1) {
			if color != "#000" && color != "#000000" {
				return file, newErrorInvalidValue(key, "the monochromeLogo is not monochrome (black): %s", value)
			}
		}
	} else if ext == ".svgz" {
		// Regex for every hex color.
		re := regexp.MustCompile("#(?:[0-9a-fA-F]{3}){1,2}")

		// Read file data.
		data, err := ioutil.ReadFile(value)
		if err != nil {
			return file, err
		}
		data, err = gUnzipData(data)
		if err != nil {
			return file, err
		}

		for _, color := range re.FindAllString(string(data), -1) {
			if color != "#000" && color != "#000000" {
				return file, newErrorInvalidValue(key, "the monochromeLogo is not monochrome (black): %s", value)
			}
		}
	}

	return file, nil
}

// checkMIME tells whether the string in input is a well formatted MIME or not.
func (p *parser) checkMIME(key string, value string) error {
	// Regex for MIME.
	// Reference: https://github.com/jshttp/media-typer/
	re := regexp.MustCompile("^ *([A-Za-z0-9][A-Za-z0-9!#$&^_-]{0,126})/([A-Za-z0-9][A-Za-z0-9!#$&^_.+-]{0,126}) *$")

	if !re.MatchString(value) {
		return newErrorInvalidValue(key, " %s is not a valid MIME.", value)
	}

	return nil
}

// contains returns true if the slice of strings contains the searched string.
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// gUnzipData g-unzip a list of bytes. (used for svgz unzip)
func gUnzipData(data []byte) (resData []byte, err error) {
	b := bytes.NewBuffer(data)

	var r io.Reader
	r, err = gzip.NewReader(b)
	if err != nil {
		return
	}

	var resB bytes.Buffer
	_, err = resB.ReadFrom(r)
	if err != nil {
		return
	}

	resData = resB.Bytes()

	return
}
