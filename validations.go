package publiccode

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/png"
	"math/rand"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"slices"
	"strings"

	"github.com/alranel/go-vcsurl/v2"
	"github.com/dyatlov/go-oembed/oembed"
	httpclient "github.com/italia/httpclient-lib-go"
	"github.com/italia/publiccode-parser-go/v5/data"
	"github.com/italia/publiccode-parser-go/v5/internal/git"
	netutil "github.com/italia/publiccode-parser-go/v5/internal"
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

	return stringInSlice(urlP.Host, domain.UseTokenFor)
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

// isReachable checks whether the URL resource is reachable.
// An URL resource is reachable if it returns HTTP 200.
func (p *Parser) isReachable(u url.URL, network bool) (bool, error) {
	// Don't check if the network checks are disabled or if we are running in WASM
	// because we'd most likely fail due to CORS errors.
	if !network || runtime.GOARCH == "wasm" {
		return true, nil
	}

	if u.Scheme == "" {
		return false, fmt.Errorf("missing URL scheme")
	}

	r, err := httpclient.GetURL(u.String(), getHeaderFromDomain(p.domain, u.String()))
	if err != nil {
		return false, fmt.Errorf("HTTP GET failed for %s: %v", u.String(), err)
	}

	if r.Status.Code != 200 {
		return false, fmt.Errorf("HTTP GET returned %d for %s; 200 was expected", r.Status.Code, u.String())
	}

	return true, nil
}

// toURL turns the passed string into an URL, trying to resolve
// code hosting URLs to their raw URL.
//
// It supports relative paths and turns them into remote URLs or file:// URLs
// depending on the value of baseURL.
// If a URL is not supported by vcsurl, but it looks like a Git repository,
// it returns the URL as-is.
func toCodeHostingURL(file string, baseURL *url.URL, allowGitClone bool) (url.URL, bool) {
	// Check if file is an absolute URL
	if uri, err := url.ParseRequestURI(file); err == nil {
		if raw := vcsurl.GetRawFile(uri); raw != nil {
			return *raw, false
		}

		// Check if it's a Git repository URL that we can clone
		if allowGitClone && git.IsGitURL(uri) {
			return *uri, true
		}

		return *uri, false
	}

	// We always pass the computed base URL here, with a fallback to the cwd,
	// so baseURL won't be nil. I'd be a programming mistake in case it is, so
	// it'll be good to crash here.
	u := *baseURL
	u.Path = path.Join(u.Path, file)

	// Check if the base URL itself it's a Git repository
	if allowGitClone && git.IsGitURL(&u) {
		return u, true
	}

	return u, false
}

// Returns true if the file resource exists.
func (p *Parser) fileExists(u url.URL, network bool, isGitRepo bool) (bool, error) {
	// Don't check if we are running in WASM because there's no stat(2) there
	if runtime.GOARCH == "wasm" {
		return true, nil
	}

	// If we have an absolute local path, perform validation on it, otherwise do it
	// on the remote URL if any. If none are available, validation is skipped.
	if u.Scheme == "file" {
		_, err := os.Stat(u.Path)
		if err != nil {
			err = fmt.Errorf("no such file: %s", netutil.DisplayURL(&u))
		}

		return err == nil, err
	}

	// Check file existence via Git clone
	if isGitRepo && p.allowLocalGitClone && network {
		exists, _, err := p.checkFileInGitRepo(&u)
		if err == nil {
			return exists, nil
		}
	}

	// Check file existence via HTTP request
	if network {
		reachable, err := p.isReachable(u, network)

		return reachable, err
	}

	return true, nil
}

// Checks whether the string is a valid image. It also checks if the file exists.
// It returns true if it is an image or false if it's not and an error, if any.
func (p *Parser) isImageFile(u url.URL, network bool, isGitRepo bool) (bool, error) {
	validExt := []string{".jpg", ".png"}
	ext := strings.ToLower(filepath.Ext(u.Path))

	if !slices.Contains(validExt, ext) {
		return false, fmt.Errorf("invalid file extension for: %s", netutil.DisplayURL(&u))
	}

	return p.fileExists(u, network, isGitRepo)
}

// Returns true if the file path in value is a valid logo.
// It also checks if the file exists.
func (p *Parser) validLogo(u url.URL, network bool, isGitRepo bool) (bool, error) {
	validExt := []string{".svg", ".svgz", ".png"}
	ext := strings.ToLower(filepath.Ext(u.Path))

	// Check for valid extension.
	if !slices.Contains(validExt, ext) {
		return false, fmt.Errorf("invalid file extension for: %s", netutil.DisplayURL(&u))
	}

	if exists, err := p.fileExists(u, network, isGitRepo); !exists {
		return false, err
	}

	var localPath string

	// If it's a Git repository, try to get the file locally
	if isGitRepo && p.allowLocalGitClone && network {
		exists, gitLocalPath, err := p.checkFileInGitRepo(&u)
		if err == nil && exists {
			localPath = gitLocalPath
		}
	}

	// If we don't have a local path yet and it's a remote file, download it
	if localPath == "" && u.Scheme != "file" {
		var err error

		if !network {
			return true, nil
		}

		localPath, err = netutil.DownloadTmpFile(&u, getHeaderFromDomain(p.domain, u.String()))
		if err != nil {
			return false, err
		}

		defer func() {
			if localPath == "" {
				return
			}

			if err := os.Remove(localPath); err != nil {
				fmt.Fprintf(os.Stderr, "failed to remove %s: %v\n", localPath, err)
			}

			dir := path.Dir(localPath)
			if err = os.Remove(dir); err != nil {
				fmt.Fprintf(os.Stderr, "failed to remove %s: %v\n", dir, err)
			}
		}()
	} else if localPath == "" && u.Scheme == "file" {
		localPath = u.Path
	}

	if localPath != "" && ext == ".png" {
		image.RegisterFormat("png", "png", png.Decode, png.DecodeConfig)

		f, err := os.Open(localPath)
		if err != nil {
			return false, err
		}

		image, _, err := image.DecodeConfig(f)
		if err != nil {
			return false, err
		}

		if image.Width < minLogoWidth {
			return false, fmt.Errorf("invalid image size of %d (min %dpx of width): %s", image.Width, minLogoWidth, netutil.DisplayURL(&u))
		}
	}

	return true, nil
}

// isOembedURL returns whether the link is from a valid oEmbed provider.
// Reference: https://oembed.com/providers.json
func (p *Parser) isOEmbedURL(url *url.URL) (bool, error) {
	b := data.OembedProviders
	oe := oembed.NewOembed()
	_ = oe.ParseProviders(bytes.NewReader(b))

	link := url.String()
	if item := oe.FindItem(link); item == nil {
		return false, fmt.Errorf("invalid oEmbed link: %s", link)
	}

	return true, nil
}
