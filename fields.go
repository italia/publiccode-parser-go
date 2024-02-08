package publiccode

import (
	"fmt"
	"net/url"

	"github.com/alranel/go-vcsurl/v2"
	urlutil "github.com/italia/publiccode-parser-go/v4/internal"
)

type validateFn func(publiccode PublicCode, parser Parser, network bool) error

// validateFields validates publiccode with additional rules not validatable
// with a simple YAML schema.
// It returns any error encountered as ValidationResults.
func validateFieldsV0(publiccode PublicCode, parser Parser, network bool) error {
	publiccodev0 := publiccode.(PublicCodeV0)

	var vr ValidationResults

	if publiccodev0.URL != nil && network {
		if reachable, err := parser.isReachable(*(*url.URL)(publiccodev0.URL), network); !reachable {
			vr = append(vr, newValidationError("url", "'%s' not reachable: %s", publiccodev0.URL, err.Error()))
		}

		if !vcsurl.IsRepo((*url.URL)(publiccodev0.URL)) {
			vr = append(vr, newValidationError("url", "is not a valid code repository"))
		}
	}

	if publiccodev0.LandingURL != nil {
		if reachable, err := parser.isReachable(*(*url.URL)(publiccodev0.LandingURL), network); !reachable {
			vr = append(vr, newValidationError(
				"landingURL",
				"'%s' not reachable: %s", publiccodev0.LandingURL, err.Error(),
			))
		}
	}

	if publiccodev0.Roadmap != nil {
		if reachable, err := parser.isReachable(*(*url.URL)(publiccodev0.Roadmap), network); !reachable {
			vr = append(vr, newValidationError(
				"roadmap",
				"'%s' not reachable: %s", publiccodev0.Roadmap, err.Error(),
			))
		}
	}

	if publiccodev0.Logo != "" {
		if validLogo, err := parser.validLogo(toCodeHostingURL(publiccodev0.Logo, parser.baseURL), network); !validLogo {
			vr = append(vr, newValidationError("logo", err.Error()))
		}
	}

	if publiccodev0.MonochromeLogo != "" {
		vr = append(vr, ValidationWarning{"monochromeLogo", "This key is DEPRECATED and will be removed in the future", 0, 0})

		if validLogo, err := parser.validLogo(toCodeHostingURL(publiccodev0.MonochromeLogo, parser.baseURL), network); !validLogo {
			vr = append(vr, newValidationError("monochromeLogo", err.Error()))
		}
	}

	if publiccodev0.Legal.AuthorsFile != nil {
		vr = append(vr, ValidationWarning{"legal.authorsFile", "This key is DEPRECATED and will be removed in the future", 0, 0})

		if !parser.fileExists(toCodeHostingURL(*publiccodev0.Legal.AuthorsFile, parser.baseURL), network) {
			u := toCodeHostingURL(*publiccodev0.Legal.AuthorsFile, parser.baseURL)

			vr = append(vr, newValidationError("legal.authorsFile", "'%s' does not exist", urlutil.DisplayURL(&u)))
		}
	}

	if len(publiccodev0.InputTypes) > 0 {
		vr = append(vr, ValidationWarning{"inputTypes", "This key is DEPRECATED and will be removed in the future", 0, 0})
	}

	if len(publiccodev0.OutputTypes) > 0 {
		vr = append(vr, ValidationWarning{"outputTypes", "This key is DEPRECATED and will be removed in the future", 0, 0})
	}

	for lang, desc := range publiccodev0.Description {
		if publiccodev0.Description == nil {
			publiccodev0.Description = make(map[string]DescV0)
		}

		if len(desc.GenericName) > 0 {
			vr = append(vr, ValidationWarning{
				fmt.Sprintf("description.%s.genericName", lang),
				"This key is DEPRECATED and will be removed in the future", 0, 0,
			})
		}

		if network && desc.Documentation != nil {
			if reachable, err := parser.isReachable(*(*url.URL)(desc.Documentation), network); !reachable {
				vr = append(vr, newValidationError(
					fmt.Sprintf("description.%s.documentation", lang),
					"'%s' not reachable: %s", desc.Documentation, err.Error(),
				))
			}
		}

		if network && desc.APIDocumentation != nil {
			if reachable, err := parser.isReachable(*(*url.URL)(desc.APIDocumentation), network); !reachable {
				vr = append(vr, newValidationError(
					fmt.Sprintf("description.%s.apiDocumentation", lang),
					"'%s' not reachable: %s", desc.APIDocumentation, err.Error(),
				))
			}
		}

		for i, v := range desc.Screenshots {
			if isImage, err := parser.isImageFile(toCodeHostingURL(v, parser.baseURL), network); !isImage {
				vr = append(vr, newValidationError(
					fmt.Sprintf("description.%s.screenshots[%d]", lang, i),
					"'%s' is not an image: %s", v, err.Error(),
				))
			}
		}

		for i, v := range desc.Videos {
			_, err := parser.isOEmbedURL((*url.URL)(v))
			if err != nil {
				vr = append(vr, newValidationError(
					fmt.Sprintf("description.%s.videos[%d]", lang, i),
					"'%s' is not a valid video URL supporting oEmbed: %s", v, err.Error(),
				))
			}
		}
	}

	if len(vr) == 0 {
		return nil
	}

	return vr
}
