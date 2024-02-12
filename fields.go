package publiccode

import (
	"fmt"
	"net/url"
	"slices"

	spdxValidator "github.com/kyoh86/go-spdx/spdx"
	"github.com/alranel/go-vcsurl/v2"

	urlutil "github.com/italia/publiccode-parser-go/v3/internal"
)

// validateFields validates publiccode with additional rules not validatable
// with a simple YAML schema.
// It returns any error encountered as ValidationResults.
func (p *Parser) validateFields() error {
	var vr ValidationResults
	var err error

	if (p.PublicCode.URL != nil) {
		if reachable, err := p.isReachable(*(*url.URL)(p.PublicCode.URL)); !reachable {
			vr =  append(vr, newValidationError("url", "'%s' not reachable: %s", p.PublicCode.URL, err.Error()))
		}
		if !vcsurl.IsRepo((*url.URL)(p.PublicCode.URL)) {
			vr =  append(vr, newValidationError("url", "is not a valid code repository"))
		}
	}

	if p.PublicCode.LandingURL != nil {
		if reachable, err := p.isReachable(*(*url.URL)(p.PublicCode.LandingURL)); !reachable {
			vr =  append(vr, newValidationError(
				"landingURL",
				"'%s' not reachable: %s", p.PublicCode.LandingURL, err.Error(),
			))
		}
	}

	if p.PublicCode.Roadmap != nil {
		if reachable, err := p.isReachable(*(*url.URL)(p.PublicCode.Roadmap)); !reachable {
			vr =  append(vr, newValidationError(
				"roadmap",
				"'%s' not reachable: %s", p.PublicCode.Roadmap, err.Error(),
			))
		}
	}

	if (p.PublicCode.Logo != "") {
		if validLogo, err := p.validLogo(p.toURL(p.PublicCode.Logo)); !validLogo {
			vr =  append(vr, newValidationError("logo", err.Error()))
		}
	}
	if (p.PublicCode.MonochromeLogo != "") {
		vr =  append(vr, ValidationWarning{"monochromeLogo", "This key is DEPRECATED and will be removed in the future", 0, 0})

		if validLogo, err := p.validLogo(p.toURL(p.PublicCode.MonochromeLogo)); !validLogo {
			vr =  append(vr, newValidationError("monochromeLogo", err.Error()))
		}
	}

	if p.PublicCode.Legal.AuthorsFile != nil && !p.fileExists(p.toURL(*p.PublicCode.Legal.AuthorsFile)) {
		u := p.toURL(*p.PublicCode.Legal.AuthorsFile)

		vr =  append(vr, newValidationError("legal.authorsFile", "'%s' does not exist", urlutil.DisplayURL(&u)))
	}

	if p.PublicCode.Legal.License != "" {
		_, err = spdxValidator.Parse(p.PublicCode.Legal.License)
		if err != nil {
			vr =  append(vr, newValidationError(
				"legal.license",
				"invalid license '%s'", p.PublicCode.Legal.License,
			))
		}
	}

	if p.PublicCode.It.CountryExtensionVersion != "" &&
		!slices.Contains(ExtensionITSupportedVersions, p.PublicCode.It.CountryExtensionVersion) {

		vr =  append(vr, newValidationError(
			"it.countryExtensionVersion",
			"version %s not supported for 'it' extension", p.PublicCode.It.CountryExtensionVersion,
		))
	}

	if len(p.PublicCode.InputTypes) > 0 {
		vr =  append(vr, ValidationWarning{"inputTypes", "This key is DEPRECATED and will be removed in the future", 0, 0})
	}
	for i, mimeType := range p.PublicCode.InputTypes {
		if !p.isMIME(mimeType) {
			vr =  append(vr, newValidationError(
				fmt.Sprintf("inputTypes[%d]", i), "'%s' is not a MIME type", mimeType,
			))
		}
	}

	if len(p.PublicCode.OutputTypes) > 0 {
		vr =  append(vr, ValidationWarning{"outputTypes", "This key is DEPRECATED and will be removed in the future", 0, 0})
	}
	for i, mimeType := range p.PublicCode.OutputTypes {
		if !p.isMIME(mimeType) {
			vr =  append(vr, newValidationError(
				fmt.Sprintf("outputTypes[%d]", i), "'%s' is not a MIME type", mimeType,
			))
		}
	}

	for lang, desc := range p.PublicCode.Description {
		if p.PublicCode.Description == nil {
			p.PublicCode.Description = make(map[string]Desc)
		}

		if len(desc.GenericName) > 0 {
			vr =  append(vr, ValidationWarning{
				fmt.Sprintf("description.%s.genericName", lang),
				"This key is DEPRECATED and will be removed in the future", 0, 0,
			})
		}

		if (desc.Documentation != nil) {
			if reachable, err := p.isReachable(*(*url.URL)(desc.Documentation)); !reachable {
				vr =  append(vr, newValidationError(
					fmt.Sprintf("description.%s.documentation", lang),
					"'%s' not reachable: %s", desc.Documentation, err.Error(),
				))
			}
		}
		if desc.APIDocumentation != nil {
			if reachable, err := p.isReachable(*(*url.URL)(desc.APIDocumentation)); !reachable {
				vr =  append(vr, newValidationError(
					fmt.Sprintf("description.%s.apiDocumentation", lang),
					"'%s' not reachable: %s", desc.APIDocumentation, err.Error(),
				))
			}
		}

		for i, v := range desc.Screenshots {
			if isImage, err := p.isImageFile(p.toURL(v)); !isImage {
				vr =  append(vr, newValidationError(
					fmt.Sprintf("description.%s.screenshots[%d]", lang, i),
					"'%s' is not an image: %s", v, err.Error(),
				))
			}
		}
		for i, v := range desc.Videos {
			_, err := p.isOEmbedURL((*url.URL)(v))
			if err != nil {
				vr =  append(vr, newValidationError(
					fmt.Sprintf("description.%s.videos[%d]", lang, i),
					"'%s' is not a valid video URL supporting oEmbed: %s", v, err.Error(),
				))
			}
		}
	}

	if (len(vr) == 0) {
		return nil
	}

	return vr
}
