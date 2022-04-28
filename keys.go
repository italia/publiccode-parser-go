package publiccode

import (
	"fmt"
	"net/url"

	spdxValidator "github.com/kyoh86/go-spdx/spdx"
	"github.com/alranel/go-vcsurl/v2"
	"github.com/thoas/go-funk"

	urlutil "github.com/italia/publiccode-parser-go/v2/internal"
)

// validateFields validates publiccode with additional rules not validatable
// with a simple YAML schema.
// It returns any error encountered as ValidationErrors.
func (p *Parser) validateFields() error {
	var ve ValidationErrors
	var err error

	if (p.PublicCode.URL != nil) {
		if reachable, err := p.isReachable(*(*url.URL)(p.PublicCode.URL)); !reachable {
			ve = append(ve, newValidationError("url", "'%s' not reachable: %s", p.PublicCode.URL, err.Error()))
		}
		if !vcsurl.IsRepo((*url.URL)(p.PublicCode.URL)) {
			ve = append(ve, newValidationError("url", "is not a valid code repository"))
		}
	}

	if p.PublicCode.LandingURL != nil {
		if reachable, err := p.isReachable(*(*url.URL)(p.PublicCode.LandingURL)); !reachable {
			ve = append(ve, newValidationError(
				"landingURL",
				"'%s' not reachable: %s", p.PublicCode.LandingURL, err.Error(),
			))
		}
	}

	if p.PublicCode.Roadmap != nil {
		if reachable, err := p.isReachable(*(*url.URL)(p.PublicCode.Roadmap)); !reachable {
			ve = append(ve, newValidationError(
				"roadmap",
				"'%s' not reachable: %s", p.PublicCode.Roadmap, err.Error(),
			))
		}
	}

	if (p.PublicCode.Logo != "") {
		if validLogo, err := p.validLogo(p.toURL(p.PublicCode.Logo)); !validLogo {
			ve = append(ve, newValidationError("logo", err.Error()))
		}
	}
	if (p.PublicCode.MonochromeLogo != "") {
		if validLogo, err := p.validLogo(p.toURL(p.PublicCode.MonochromeLogo)); !validLogo {
			ve = append(ve, newValidationError("monochromeLogo", err.Error()))
		}
	}

	if p.PublicCode.Legal.AuthorsFile != nil && !p.fileExists(p.toURL(*p.PublicCode.Legal.AuthorsFile)) {
		u := p.toURL(*p.PublicCode.Legal.AuthorsFile)

		ve = append(ve, newValidationError("legal.authorsFile", "'%s' does not exist", urlutil.DisplayURL(&u)))
	}

	if p.PublicCode.Legal.License != "" {
		_, err = spdxValidator.Parse(p.PublicCode.Legal.License)
		if err != nil {
			ve = append(ve, newValidationError(
				"legal.license",
				"invalid license '%s'", p.PublicCode.Legal.License,
			))
		}
	}

	if p.PublicCode.It.CountryExtensionVersion != "" &&
		!funk.Contains(ExtensionITSupportedVersions, p.PublicCode.It.CountryExtensionVersion) {

		ve = append(ve, newValidationError(
			"it.countryExtensionVersion",
			"version %s not supported for 'it' extension", p.PublicCode.It.CountryExtensionVersion,
		))
	}

	for i, mimeType := range p.PublicCode.InputTypes {
		if !p.isMIME(mimeType) {
			ve = append(ve, newValidationError(
				fmt.Sprintf("inputTypes[%d]", i), "'%s' is not a MIME type", mimeType,
			))
		}
	}
	for i, mimeType := range p.PublicCode.OutputTypes {
		if !p.isMIME(mimeType) {
			ve = append(ve, newValidationError(
				fmt.Sprintf("outputTypes[%d]", i), "'%s' is not a MIME type", mimeType,
			))
		}
	}

	for lang, desc := range p.PublicCode.Description {
		if p.PublicCode.Description == nil {
			p.PublicCode.Description = make(map[string]Desc)
		}

		if (desc.Documentation != nil) {
			if reachable, err := p.isReachable(*(*url.URL)(desc.Documentation)); !reachable {
				ve = append(ve, newValidationError(
					fmt.Sprintf("description.%s.documentation", lang),
					"'%s' not reachable: %s", desc.Documentation, err.Error(),
				))
			}
		}
		if desc.APIDocumentation != nil {
			if reachable, err := p.isReachable(*(*url.URL)(desc.APIDocumentation)); !reachable {
				ve = append(ve, newValidationError(
					fmt.Sprintf("description.%s.apiDocumentation", lang),
					"'%s' not reachable: %s", desc.APIDocumentation, err.Error(),
				))
			}
		}

		for i, v := range desc.Screenshots {
			if isImage, err := p.isImageFile(p.toURL(v)); !isImage {
				ve = append(ve, newValidationError(
					fmt.Sprintf("description.%s.screenshots[%d]", lang, i),
					"'%s' is not an image: %s", v, err.Error(),
				))
			}
		}
		for i, v := range desc.Videos {
			_, err := p.isOembedURL((*url.URL)(v))
			if err != nil {
				ve = append(ve, newValidationError(
					fmt.Sprintf("description.%s.videos[%d]", lang, i),
					"'%s' is not a valid video URL supporting oEmbed: %s", v, err.Error(),
				))
			}
		}
	}

	if (len(ve) == 0) {
		return nil
	}

	return ve
}
