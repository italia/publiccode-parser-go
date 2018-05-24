package publiccode

import (
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// checkEmail tells whether email is well formatted.
// In general an email is valid if compile the regex: ^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$
func (p *parser) checkEmail(key string, fn string) error {
	re := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	if !re.MatchString(fn) {
		return newErrorInvalidValue(key, "invalid email: %v", fn)
	}

	return nil
}

// checkUrl tells whether the URL resource is well formatted and reachable and return it as *url.URL.
// An URL resource is well formatted if it's' a valid URL and the scheme is not empty.
// An URL resource is reachable if returns an http Status = 200 OK.
func (p *parser) checkUrl(key string, value string) (*url.URL, error) {
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
		if _, err := os.Stat(fn); err != nil {
			return "", newErrorInvalidValue(key, "file does not exist: %v", fn)
		}
	} else {
		//Remote bitbucket
		_, err := p.checkUrl(key, BaseDir+fn)

		//_, err := p.checkUrl(key, "https://bitbucket.org/marco-capobussi/publiccode-example/raw/master/"+fn)
		if err != nil {
			return "", newErrorInvalidValue(key, "file does not exist on remote: %v", BaseDir+fn)
		}
	}
	return fn, nil
}

// checkDate tells whether the string in input is a date in the
// format "YYYY-MM-DD", which is one of the ISO8601 allowed encoding, and return it as time.Time.
func (p *parser) checkDate(key string, value string) (time.Time, error) {
	if t, err := time.Parse("2006-01-02", value); err != nil {
		return t, newErrorInvalidValue(key, "cannot parse date: %v", err)
	} else {
		return t, nil
	}
}

// checkImage tells whether the string in a valid image. It also checks if the file exists.
// Reference: https://github.com/publiccodenet/publiccode.yml/blob/develop/schema.mdù
// TODO: check image size.
func (p *parser) checkImage(key string, value string) (string, error) {
	validExt := []string{".svg", ".svgz", ".png"}
	ext := strings.ToLower(filepath.Ext(value))

	// Check for valid extension.
	if !contains(validExt, ext) {
		return value, newErrorInvalidValue(key, "invalid file extension for: %s", value)
	}

	// Check existence of file.
	file, err := p.checkFile(key, value)
	if err != nil {
		return file, err
	}

	return file, nil
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
