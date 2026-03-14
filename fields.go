package publiccode

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/alranel/go-vcsurl/v2"
	urlutil "github.com/italia/publiccode-parser-go/v5/internal"
	publiccodeValidator "github.com/italia/publiccode-parser-go/v5/validators"
)

type validateFn func(publiccode PublicCode, parser Parser, network bool) error

// validateReachableURL checks if a URL is reachable and appends an error if not.
// Does nothing if checksNetwork is false or u is nil.
func (p *Parser) validateReachableURL(u *URL, field string, checksNetwork bool) ValidationResults {
	if !checksNetwork || u == nil {
		return nil
	}

	var vr ValidationResults

	if reachable, err := p.isReachable(*(*url.URL)(u)); !reachable {
		vr = append(vr, newValidationErrorf(field, "'%s' not reachable: %s", u, err.Error()))
	}

	return vr
}

// validateLogoField validates a logo field (path or URL).
func (p *Parser) validateLogoField(logo *string, key string, network bool) ValidationResults {
	if logo == nil || *logo == "" {
		return nil
	}

	var vr ValidationResults

	if _, err := isRelativePathOrURL(*logo, key); err != nil {
		return ValidationResults{err}
	}

	if !p.disableExternalChecks {
		u := toAbsoluteURL(*logo, p.currentBaseURL, network)
		if u != nil {
			if valid, err := p.validLogo(*u, network); !valid {
				vr = append(vr, newValidationError(key, err.Error()))
			}
		}
	}

	return vr
}

// validateLowercaseCountries checks for deprecated lowercase country codes in a list.
func validateLowercaseCountries(prefix string, countries *[]string) ValidationResults {
	if countries == nil {
		return nil
	}

	var vr ValidationResults

	validate := publiccodeValidator.New()

	for i, c := range *countries {
		if validate.Var(c, "iso3166_1_alpha2_lower_or_upper") == nil && c == strings.ToLower(c) {
			vr = append(vr, ValidationWarning{
				fmt.Sprintf("%s[%d]", prefix, i),
				fmt.Sprintf("Lowercase country codes are DEPRECATED. Use uppercase instead ('%s')", strings.ToUpper(c)),
				0, 0,
			})
		}
	}

	return vr
}

// validateDescURLs checks the reachability of documentation URLs in a description.
func (p *Parser) validateDescURLs(lang string, documentation, apiDocumentation *URL, checksNetwork bool) ValidationResults {
	var vr ValidationResults

	vr = append(vr, p.validateReachableURL(documentation, fmt.Sprintf("description.%s.documentation", lang), checksNetwork)...)
	vr = append(vr, p.validateReachableURL(apiDocumentation, fmt.Sprintf("description.%s.apiDocumentation", lang), checksNetwork)...)

	return vr
}

// validateDescScreenshots validates screenshot file references in a description.
func (p *Parser) validateDescScreenshots(lang string, screenshots []string, network bool) ValidationResults {
	var vr ValidationResults

	for i, v := range screenshots {
		keyName := fmt.Sprintf("description.%s.screenshots[%d]", lang, i)
		if _, err := isRelativePathOrURL(v, keyName); err != nil {
			vr = append(vr, err)
		} else if !p.disableExternalChecks {
			u := toAbsoluteURL(v, p.currentBaseURL, network)
			if u != nil {
				if isImage, err := p.isImageFile(*u, network); !isImage {
					vr = append(vr, newValidationErrorf(keyName, "'%s' is not an image: %s", v, err.Error()))
				}
			}
		}
	}

	return vr
}

// validateDescVideos validates video URL references in a description.
func (p *Parser) validateDescVideos(lang string, videos []*URL) ValidationResults {
	var vr ValidationResults

	for i, v := range videos {
		if err := p.checkOEmbedURL((*url.URL)(v)); err != nil {
			vr = append(vr, newValidationErrorf(
				fmt.Sprintf("description.%s.videos[%d]", lang, i),
				"'%s' is not a valid video URL supporting oEmbed: %s", v, err.Error(),
			))
		}
	}

	return vr
}

// validateITSection validates the IT country-specific extension section.
func validateITSection(it *ITSectionV0) ValidationResults {
	if it == nil {
		return nil
	}

	var vr ValidationResults

	if it.Conforme != nil {
		vr = append(vr, ValidationWarning{
			"IT.conforme",
			"This key is DEPRECATED and will be removed in the future. It's safe to drop it", 0, 0,
		})
	}

	if it.Riuso.CodiceIPA != "" {
		validate := publiccodeValidator.New()

		if validate.Var(it.Riuso.CodiceIPA, "is_italian_ipa_code") == nil {
			vr = append(vr, ValidationWarning{
				"IT.riuso.codiceIPA",
				fmt.Sprintf(
					"This key is DEPRECATED and will be removed in the future. Use 'organisation.uri' and set it to 'urn:x-italian-pa:%s' instead",
					it.Riuso.CodiceIPA,
				),
				0, 0,
			})
		}
	}

	return vr
}

// validateFieldsV0 validates publiccode.yml with additional rules not validatable
// with go-playground/validator.
// It returns any error encountered as ValidationResults.
func validateFieldsV0(publiccode PublicCode, parser Parser, network bool) error {
	publiccodev0 := publiccode.(*PublicCodeV0)

	var vr ValidationResults

	checksNetwork := network && !parser.disableExternalChecks

	vr = append(vr, parser.validateReachableURL(publiccodev0.URL, "url", checksNetwork)...)

	if checksNetwork && publiccodev0.URL != nil && !vcsurl.IsRepo((*url.URL)(publiccodev0.URL)) {
		vr = append(vr, newValidationError("url", "is not a valid code repository"))
	}

	vr = append(vr, parser.validateReachableURL(publiccodev0.LandingURL, "landingURL", checksNetwork)...)
	vr = append(vr, parser.validateReachableURL(publiccodev0.Roadmap, "roadmap", checksNetwork)...)

	vr = append(vr, parser.validateLogoField(publiccodev0.Logo, "logo", network)...)

	if publiccodev0.MonochromeLogo != nil && *publiccodev0.MonochromeLogo != "" {
		vr = append(vr, ValidationWarning{"monochromeLogo", "This key is DEPRECATED and will be removed in the future. Use 'logo' instead", 0, 0})
		vr = append(vr, parser.validateLogoField(publiccodev0.MonochromeLogo, "monochromeLogo", network)...)
	}

	if publiccodev0.IntendedAudience != nil {
		vr = append(vr, validateLowercaseCountries("intendedAudience.countries", publiccodev0.IntendedAudience.Countries)...)
		vr = append(vr, validateLowercaseCountries("intendedAudience.unsupportedCountries", publiccodev0.IntendedAudience.UnsupportedCountries)...)
	}

	if publiccodev0.Legal.AuthorsFile != nil {
		vr = append(vr, ValidationWarning{"legal.authorsFile", "This key is DEPRECATED and will be removed in the future. It's safe to drop it", 0, 0})

		if _, err := isRelativePathOrURL(*publiccodev0.Legal.AuthorsFile, "legal.authorsFile"); err != nil {
			vr = append(vr, err)
		} else if !parser.disableExternalChecks {
			u := toAbsoluteURL(*publiccodev0.Legal.AuthorsFile, parser.currentBaseURL, network)
			if u != nil {
				if exists, err := parser.fileExists(*u, network); !exists {
					vr = append(vr, newValidationErrorf("legal.authorsFile", "'%s' does not exist: %s", urlutil.DisplayURL(u), err.Error()))
				}
			}
		}
	}

	if publiccodev0.Legal.RepoOwner != nil {
		vr = append(vr, ValidationWarning{"legal.repoOwner", "This key is DEPRECATED and will be removed in the future. Use 'organisation.name' instead", 0, 0})
	}

	if publiccodev0.InputTypes != nil {
		vr = append(vr, ValidationWarning{"inputTypes", "This key is DEPRECATED and will be removed in the future. It's safe to drop it", 0, 0})
	}

	if publiccodev0.OutputTypes != nil {
		vr = append(vr, ValidationWarning{"outputTypes", "This key is DEPRECATED and will be removed in the future. It's safe to drop it", 0, 0})
	}

	for lang, desc := range publiccodev0.Description {
		vr = append(vr, parser.validateDescV0(lang, desc, checksNetwork, network)...)
	}

	it := publiccodev0.IT

	if publiccodev0.It != nil {
		vr = append(vr, ValidationWarning{
			"it",
			"Lowercase country codes are DEPRECATED and will be removed in the future. Use 'IT' instead", 0, 0,
		})

		it = publiccodev0.It
	}

	if publiccodev0.IT != nil && publiccodev0.It != nil {
		vr = append(vr, newValidationError("it", "'IT' key already present. Remove this key"))

		it = publiccodev0.IT
	}

	vr = append(vr, validateITSection(it)...)

	if len(vr) == 0 {
		return nil
	}

	return vr
}

func (p *Parser) validateDescV0(lang string, desc DescV0, checksNetwork, network bool) ValidationResults {
	var vr ValidationResults

	if len(desc.GenericName) > 0 {
		vr = append(vr, ValidationWarning{
			fmt.Sprintf("description.%s.genericName", lang),
			"This key is DEPRECATED and will be removed in the future. It's safe to drop it", 0, 0,
		})
	}

	vr = append(vr, p.validateDescURLs(lang, desc.Documentation, desc.APIDocumentation, checksNetwork)...)
	vr = append(vr, p.validateDescScreenshots(lang, desc.Screenshots, network)...)
	vr = append(vr, p.validateDescVideos(lang, desc.Videos)...)

	return vr
}

// validateFieldsV1 validates publiccode.yml with additional rules not validatable
// with go-playground/validator.
// It returns any error encountered as ValidationResults.
func validateFieldsV1(publiccode PublicCode, parser Parser, network bool) error {
	publiccodev1 := publiccode.(*PublicCodeV1)

	var vr ValidationResults

	checksNetwork := network && !parser.disableExternalChecks

	vr = append(vr, parser.validateReachableURL(publiccodev1.URL, "url", checksNetwork)...)

	if checksNetwork && publiccodev1.URL != nil && !vcsurl.IsRepo((*url.URL)(publiccodev1.URL)) {
		vr = append(vr, newValidationError("url", "is not a valid code repository"))
	}

	vr = append(vr, parser.validateReachableURL(publiccodev1.LandingURL, "landingURL", checksNetwork)...)
	vr = append(vr, parser.validateReachableURL(publiccodev1.Roadmap, "roadmap", checksNetwork)...)

	vr = append(vr, parser.validateLogoField(publiccodev1.Logo, "logo", network)...)

	if publiccodev1.IntendedAudience != nil {
		vr = append(vr, validateLowercaseCountries("intendedAudience.countries", publiccodev1.IntendedAudience.Countries)...)
		vr = append(vr, validateLowercaseCountries("intendedAudience.unsupportedCountries", publiccodev1.IntendedAudience.UnsupportedCountries)...)
	}

	// Description
	for lang, desc := range publiccodev1.Description {
		vr = append(vr, parser.validateDescV1(lang, desc, checksNetwork, network)...)
	}

	if len(vr) == 0 {
		return nil
	}

	return vr
}

func (p *Parser) validateDescV1(lang string, desc DescV1, checksNetwork, network bool) ValidationResults {
	var vr ValidationResults

	vr = append(vr, p.validateDescURLs(lang, desc.Documentation, desc.APIDocumentation, checksNetwork)...)
	vr = append(vr, p.validateDescScreenshots(lang, desc.Screenshots, network)...)
	vr = append(vr, p.validateDescVideos(lang, desc.Videos)...)

	return vr
}

// isRelativePathOrURL checks whether the field contains either a relative filename
// or an HTTP URL
//
//nolint:unparam
func isRelativePathOrURL(content string, keyName string) (bool, error) {
	if strings.HasPrefix(content, "/") {
		return false, newValidationError(keyName, "is an absolute path. Only relative paths or HTTP(s) URLs allowed")
	}

	if strings.HasPrefix(content, "file:") {
		return false, newValidationError(keyName, "is a file:// URL. Only relative paths or HTTP(s) URLs allowed")
	}

	return true, nil
}
