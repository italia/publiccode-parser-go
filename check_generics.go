package publiccode

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"image"
	"image/png"
	"io"
	"io/ioutil"
	"math/rand"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	httpclient "github.com/italia/httpclient-lib-go"
	"github.com/thoas/go-funk"
)

// Despite the spec requires at least 1000px, we temporarily release this constraint to 120px.
const minLogoWidth = 120

func getBasicAuth(domain Domain) string {
	if len(domain.BasicAuth) > 0 {
		auth := domain.BasicAuth[rand.Intn(len(domain.BasicAuth))]
		return "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
	}
	return ""
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func isHostInDomain(domain Domain, u string) bool {
	if len(domain.UseTokenFor) == 0 {
		return false
	}
	urlP, _ := url.Parse(u)
	if stringInSlice(urlP.Host, domain.UseTokenFor) {
		return true
	}
	return false
}

func getHeaderFromDomain(domain Domain, url string) map[string]string {
	if !isHostInDomain(domain, url) {
		return nil
	}
	// Set BasicAuth header
	headers := make(map[string]string)
	headers["Authorization"] = getBasicAuth(domain)
	return headers
}

// TODO checkURL tells whether the URL resource is well formatted and reachable and return it as *url.URL.
// An URL resource is well formatted if it's a valid URL and the scheme is not empty.
// An URL resource is reachable if returns an http Status = 200 OK.
func (p *Parser) checkURL(u *url.URL) (error) {
	if u.Scheme == "" {
		return ParseError{"missing URL scheme"}
	}

	if !p.DisableNetwork {
		// Check whether URL is reachable
		r, err := httpclient.GetURL(u.String(), getHeaderFromDomain(p.Domain, u.String()))
		if err != nil {
			return fmt.Errorf("HTTP GET failed for %s: %v", u.String(), err)
		}
		if r.Status.Code != 200 {
			return fmt.Errorf("HTTP GET returned %d for %s; 200 was expected", r.Status.Code, u.String())
		}
	}

	return nil
}

// getAbsolutePaths tries to compute both a local absolute path and a remote
// URL pointing to the given file, if we have enough information.
func (p *Parser) getAbsolutePaths(file string) (string, string, error) {
	var LocalPath, RemoteURL string

	// Check if file is an absolute URL
	if _, err := url.ParseRequestURI(file); err == nil {
		// If the base URL is set, we can perform validation and try to compute the local path
		if p.RemoteBaseURL != "" {
			// Let's be tolerant: turn GitHub non-raw URLs to raw URLs
			var re = regexp.MustCompile(`^https://github.com/(.+?)/(.+?)/blob/(.+)$`)
			file = re.ReplaceAllString(file, `https://raw.githubusercontent.com/$1/$2/$3`)

			// Check if the URL matches the base URL.
			// We don't allow absolute URLs not pointing to the same repository as the
			// publiccode.yml file
			if strings.Index(file, p.RemoteBaseURL) != 0 {
				return "", "", fmt.Errorf("Absolute URL (%s) is outside the repository (%s)", file, p.RemoteBaseURL)
			}

			// We can compute the local path by stripping the base URL.
			if p.LocalBasePath != "" {
				LocalPath = path.Join(p.LocalBasePath, strings.Replace(file, p.RemoteBaseURL, "", 1))
			}
		}
		RemoteURL = file
	} else {
		// If file is a relative path, let's try to compute its absolute filesystem path
		// and remote URL by prepending the base paths, if provided.
		if p.LocalBasePath != "" {
			LocalPath = path.Join(p.LocalBasePath, file)
		}
		if p.RemoteBaseURL != "" {
			u, err := url.Parse(p.RemoteBaseURL)
			if err != nil {
				return "", "", err
			}
			u.Path = path.Join(u.Path, file)
			RemoteURL = u.String()
		}
	}

	return LocalPath, RemoteURL, nil
}

// checkFile tells whether the file resource exists and return it.
func (p *Parser) checkFile(file string) (string, error) {
	// Try to compute both a local absolute path and a remote URL pointing
	// to this file, if we have enough information.
	LocalPath, RemoteURL, err := p.getAbsolutePaths(file)
	if err != nil {
		return "", err
	}

	// If we have an absolute local path, perform validation on it, otherwise do it
	// on the remote URL if any. If none are available, validation is skipped.
	if LocalPath != "" {
		if _, err := os.Stat(LocalPath); err != nil {
			return "", fmt.Errorf("local file does not exist: %v", LocalPath)
		}
	} else if RemoteURL != "" {
		url, err := url.Parse(RemoteURL)
		if err != nil {
			return "", err
		}
		err = p.checkURL(url)
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

// checkImage tells whether the string in a valid image. It also checks if the file exists.
// Reference: https://github.com/publiccodenet/publiccode.yml/blob/develop/schema.md
func (p *Parser) checkImage(value string) (string, error) {
	validExt := []string{".jpg", ".png"}
	ext := strings.ToLower(filepath.Ext(value))

	// Check for valid extension.
	if !funk.Contains(validExt, ext) {
		return value, fmt.Errorf("invalid file extension for: %s", value)
	}

	// Check existence of file.
	file, err := p.checkFile(value)

	return file, err
}

// checkLogo tells whether the string in a valid logo. It also checks if the file exists.
// Reference: https://github.com/publiccodenet/publiccode.yml/blob/develop/schema.md
func (p *Parser) checkLogo(value string) (string, error) {
	validExt := []string{".svg", ".svgz", ".png"}
	ext := strings.ToLower(filepath.Ext(value))

	// Check for valid extension.
	if !funk.Contains(validExt, ext) {
		return value, fmt.Errorf("invalid file extension for: %s", value)
	}

	// Check existence of file.
	file, err := p.checkFile(value)
	if err != nil {
		return value, err
	}

	// Try to compute both a local absolute path and a remote URL pointing
	// to this file, if we have enough information.
	localPath, remoteURL, err := p.getAbsolutePaths(file)
	if err != nil {
		return "", err
	}

	// Remote. Create a temp dir, download and check the file. Remove the temp dir.
	if localPath == "" && remoteURL != "" {
		if p.DisableNetwork {
			return file, nil
		}

		localPath, err = downloadTmpFile(remoteURL, getHeaderFromDomain(p.Domain, remoteURL))
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
				return file, fmt.Errorf("invalid image size of %d (min %dpx of width): %s", image.Width, minLogoWidth, value)
			}
		}
	}

	return file, nil
}

// checkLogo tells whether the string in a valid logo. It also checks if the file exists.
// Reference: https://github.com/publiccodenet/publiccode.yml/blob/develop/schema.md
func (p *Parser) checkMonochromeLogo(value string) (string, error) {
	validExt := []string{".svg", ".svgz", ".png"}
	ext := strings.ToLower(filepath.Ext(value))

	// Check for valid extension.
	if !funk.Contains(validExt, ext) {
		return value, fmt.Errorf("invalid file extension for: %s", value)
	}

	// Check existence of file.
	file, err := p.checkFile(value)
	if err != nil {
		return value, err
	}

	// Try to compute both a local absolute path and a remote URL pointing
	// to this file, if we have enough information.
	localPath, remoteURL, err := p.getAbsolutePaths(file)
	if err != nil {
		return "", err
	}

	// Remote. Create a temp dir, download and check the file. Remove the temp dir.
	if localPath == "" && remoteURL != "" {
		if p.DisableNetwork {
			return file, nil
		}
		localPath, err = downloadTmpFile(remoteURL, getHeaderFromDomain(p.Domain, remoteURL))
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
				return file, fmt.Errorf("invalid image size of %d (min %dpx of width): %s", width, minLogoWidth, value)
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
						return file, fmt.Errorf("the monochromeLogo is not monochrome (black): %s", value)
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
					return file, fmt.Errorf("the monochromeLogo is not monochrome (black): %s", value)
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
					return file, fmt.Errorf("the monochromeLogo is not monochrome (black): %s", value)
				}
			}
		}
	}

	return file, nil
}

// checkMIME tells whether the string in input is a well formatted MIME or not.
func (p *Parser) isMIME(value string) bool {
	// Regex for MIME.
	// Reference: https://github.com/jshttp/media-typer/
	re := regexp.MustCompile("^ *([A-Za-z0-9][A-Za-z0-9!#$&^_-]{0,126})/([A-Za-z0-9][A-Za-z0-9!#$&^_.+-]{0,126}) *$")

	return re.MatchString(value)
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
