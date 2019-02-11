package publiccode

import (
	"bytes"
	"compress/gzip"
	"image"
	"image/png"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/thoas/go-funk"
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
// An URL resource is well formatted if it's a valid URL and the scheme is not empty.
// An URL resource is reachable if returns an http Status = 200 OK.
func (p *parser) checkURL(key string, value string) (*url.URL, error) {
	//fmt.Printf("checking URL: %s\n", value)

	// Check if URL is well formatted
	u, err := url.Parse(value)
	if err != nil {
		return nil, newErrorInvalidValue(key, "not a valid URL: %s: %v", value, err)
	}
	if u.Scheme == "" {
		return nil, newErrorInvalidValue(key, "missing URL scheme: %s", value)
	}

	// Check whether URL is reachable
	r, err := http.Get(value)
	if err != nil {
		return nil, newErrorInvalidValue(key, "HTTP GET failed for %s: %v", value, err)
	}
	if r.StatusCode != 200 {
		return nil, newErrorInvalidValue(key, "HTTP GET returned %d for %s; 200 was expected", r.StatusCode, value)
	}

	return u, nil
}

// getAbsolutePaths tries to compute both a local absolute path and a remote
// URL pointing to the given file, if we have enough information.
func (p *parser) getAbsolutePaths(key, file string) (string, string, error) {
	var LocalPath, RemoteURL string

	// Check if file is an absolute URL
	if _, err := url.ParseRequestURI(file); err == nil {
		// If the base URL is set, we can perform validation and try to compute the local path
		if RemoteBaseURL != "" {
			// Check if the URL matches the base URL.
			// We don't allow absolute URLs not pointing to the same repository as the
			// publiccode.yml file
			if strings.Index(file, RemoteBaseURL) != 0 {
				return "", "", newErrorInvalidValue(key, "Absolute URL (%s) is outside the repository (%s)", file, RemoteBaseURL)
			}

			// We can compute the local path by stripping the base URL.
			if LocalBasePath != "" {
				LocalPath = path.Join(LocalBasePath, strings.Replace(file, RemoteBaseURL, "", 1))
			}
		}
		RemoteURL = file
	} else {
		// If file is a relative path, let's try to compute its absolute filesystem path
		// and remote URL by prepending the base paths, if provided.
		if LocalBasePath != "" {
			LocalPath = path.Join(LocalBasePath, file)
		}
		if RemoteBaseURL != "" {
			u, err := url.Parse(RemoteBaseURL)
			if err != nil {
				return "", "", err
			}
			u.Path = path.Join(u.Path, file)
			RemoteURL = u.String()
		}
	}

	//fmt.Printf("file = %s\n", file)
	//fmt.Printf("  LocalPath = %s\n", LocalPath)
	//fmt.Printf("  RemoteURL = %s\n", RemoteURL)

	return LocalPath, RemoteURL, nil
}

// checkFile tells whether the file resource exists and return it.
func (p *parser) checkFile(key, file string) (string, error) {
	// Try to compute both a local absolute path and a remote URL pointing
	// to this file, if we have enough information.
	LocalPath, RemoteURL, err := p.getAbsolutePaths(key, file)
	if err != nil {
		return "", err
	}

	// If we have an absolute local path, perform validation on it, otherwise do it
	// on the remote URL if any. If none are available, validation is skipped.
	if LocalPath != "" {
		if _, err := os.Stat(LocalPath); err != nil {
			return "", newErrorInvalidValue(key, "local file does not exist: %v", LocalPath)
		}
	} else if RemoteURL != "" {
		_, err := p.checkURL(key, RemoteURL)
		if err != nil {
			return "", err
		}
	}

	// Return the absolute remote URL if any, or the original relative path
	// (returning the local path would be pointless as we assume it's a temporary
	// working directory)
	if RemoteURL != "" {
		return RemoteURL, nil
	}
	return file, nil
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
	if !funk.Contains(validExt, ext) {
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
	if !funk.Contains(validExt, ext) {
		return value, newErrorInvalidValue(key, "invalid file extension for: %s", value)
	}

	// Check existence of file.
	file, err := p.checkFile(key, value)
	if err != nil {
		return value, err
	}

	// Try to compute both a local absolute path and a remote URL pointing
	// to this file, if we have enough information.
	localPath, remoteURL, err := p.getAbsolutePaths(key, file)
	if err != nil {
		return "", err
	}

	// Remote. Create a temp dir, download and check the file. Remove the temp dir.
	if localPath == "" && remoteURL != "" {
		localPath, err = downloadTmpFile(remoteURL)
		if err != nil {
			return file, err
		}
		defer func() { os.Remove(path.Dir(localPath)) }()
	}

	if localPath != "" {
		// Check for image size if .png.
		if ext == ".png" {
			f, err := os.Open(localPath)
			if err != nil {
				return file, err
			}
			image, _, err := image.DecodeConfig(f)
			if err != nil {
				return file, err
			}

			if image.Width < minLogoWidth {
				return file, newErrorInvalidValue(key, "invalid image size of %d (min %dpx of width): %s", image.Width, minLogoWidth, value)
			}
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
	if !funk.Contains(validExt, ext) {
		return value, newErrorInvalidValue(key, "invalid file extension for: %s", value)
	}

	// Check existence of file.
	file, err := p.checkFile(key, value)
	if err != nil {
		return value, err
	}

	// Try to compute both a local absolute path and a remote URL pointing
	// to this file, if we have enough information.
	localPath, remoteURL, err := p.getAbsolutePaths(key, file)
	if err != nil {
		return "", err
	}

	// Remote. Create a temp dir, download and check the file. Remove the temp dir.
	if localPath == "" && remoteURL != "" {
		localPath, err = downloadTmpFile(remoteURL)
		if err != nil {
			return file, err
		}
		defer func() { os.Remove(path.Dir(localPath)) }()
	}

	if localPath != "" {
		// Check for image size if .png.
		if ext == ".png" {
			image.RegisterFormat("png", "png", png.Decode, png.DecodeConfig)

			f, err := os.Open(localPath)
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
				return file, newErrorInvalidValue(key, "invalid image size of %d (min %dpx of width): %s", width, minLogoWidth, value)
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
			data, err := ioutil.ReadFile(localPath)
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
			data, err := ioutil.ReadFile(localPath)
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

// gUnzipData g-unzip a list of bytes. (used for svgz unzip)
func gUnzipData(data []byte) (resData []byte, err error) {
	b := bytes.NewBuffer(data)

	var r io.Reader
	r, err = gzip.NewReader(b)
	if err != nil {
		return nil, err
	}

	var resB bytes.Buffer
	_, err = resB.ReadFrom(r)
	if err != nil {
		return nil, err
	}

	return resB.Bytes(), nil
}
