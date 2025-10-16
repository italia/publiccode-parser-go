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

// validateFields validates publiccode with additional rules not validatable
// with a simple YAML schema.
// It returns any error encountered as ValidationResults.
func validateFieldsV0(publiccode PublicCode, parser Parser, network bool) error { //nolint:maintidx
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

	if publiccodev0.Logo != nil && *publiccodev0.Logo != "" {
		if _, err := isRelativePathOrURL(*publiccodev0.Logo, "logo"); err != nil {
			vr = append(vr, err)
		} else if !parser.disableExternalChecks {
			validLogo, err := parser.validLogo(toCodeHostingURL(*publiccodev0.Logo, parser.currentBaseURL), network)
			if !validLogo {
				vr = append(vr, newValidationError("logo", err.Error()))
			}
		}
	}

	if publiccodev0.MonochromeLogo != nil && *publiccodev0.MonochromeLogo != "" {
		vr = append(vr, ValidationWarning{"monochromeLogo", "This key is DEPRECATED and will be removed in the future. Use 'logo' instead", 0, 0})

		if _, err := isRelativePathOrURL(*publiccodev0.MonochromeLogo, "monochromeLogo"); err != nil {
			vr = append(vr, err)
		} else if !parser.disableExternalChecks {
			validLogo, err := parser.validLogo(toCodeHostingURL(*publiccodev0.MonochromeLogo, parser.currentBaseURL), network)
			if !validLogo {
				vr = append(vr, newValidationError("monochromeLogo", err.Error()))
			}
		}
	}

	if publiccodev0.IntendedAudience != nil {
		// This is not ideal, but we need to revalidate the countries
		// here, because otherwise we could get a warning and the advice
		// to use uppercase on an invalid country.
		validate := publiccodeValidator.New()

		if publiccodev0.IntendedAudience.Countries != nil {
			for i, c := range *publiccodev0.IntendedAudience.Countries {
				if validate.Var(c, "iso3166_1_alpha2_lower_or_upper") == nil && c == strings.ToLower(c) {
					vr = append(vr, ValidationWarning{
						fmt.Sprintf("intendedAudience.countries[%d]", i),
						fmt.Sprintf("Lowercase country codes are DEPRECATED. Use uppercase instead ('%s')", strings.ToUpper(c)),
						0, 0,
					})
				}
			}
		}

		if publiccodev0.IntendedAudience.UnsupportedCountries != nil {
			for i, c := range *publiccodev0.IntendedAudience.UnsupportedCountries {
				if validate.Var(c, "iso3166_1_alpha2_lower_or_upper") == nil && c == strings.ToLower(c) {
					vr = append(vr, ValidationWarning{
						fmt.Sprintf("intendedAudience.unsupportedCountries[%d]", i),
						fmt.Sprintf("Lowercase country codes are DEPRECATED. Use uppercase instead ('%s')", strings.ToUpper(c)),
						0, 0,
					})
				}
			}
		}
	}

	if publiccodev0.Legal.AuthorsFile != nil {
		vr = append(vr, ValidationWarning{"legal.authorsFile", "This key is DEPRECATED and will be removed in the future. It's safe to drop it", 0, 0})

		if _, err := isRelativePathOrURL(*publiccodev0.Legal.AuthorsFile, "legal.authorsFile"); err != nil {
			vr = append(vr, err)
		} else if !parser.disableExternalChecks {
			exists, err := parser.fileExists(toCodeHostingURL(*publiccodev0.Legal.AuthorsFile, parser.currentBaseURL), network)
			if !exists {
				u := toCodeHostingURL(*publiccodev0.Legal.AuthorsFile, parser.currentBaseURL)

				vr = append(vr, newValidationError("legal.authorsFile", "'%s' does not exist: %s", urlutil.DisplayURL(&u), err.Error()))
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
		if publiccodev0.Description == nil {
			publiccodev0.Description = make(map[string]DescV0)
		}

		if len(desc.GenericName) > 0 {
			vr = append(vr, ValidationWarning{
				fmt.Sprintf("description.%s.genericName", lang),
				"This key is DEPRECATED and will be removed in the future. It's safe to drop it", 0, 0,
			})
		}

		if !parser.disableExternalChecks && network && desc.Documentation != nil {
			if reachable, err := parser.isReachable(*(*url.URL)(desc.Documentation), network); !reachable {
				vr = append(vr, newValidationError(
					fmt.Sprintf("description.%s.documentation", lang),
					"'%s' not reachable: %s", desc.Documentation, err.Error(),
				))
			}
		}

		if !parser.disableExternalChecks && network && desc.APIDocumentation != nil {
			if reachable, err := parser.isReachable(*(*url.URL)(desc.APIDocumentation), network); !reachable {
				vr = append(vr, newValidationError(
					fmt.Sprintf("description.%s.apiDocumentation", lang),
					"'%s' not reachable: %s", desc.APIDocumentation, err.Error(),
				))
			}
		}

		for i, v := range desc.Screenshots {
			keyName := fmt.Sprintf("description.%s.screenshots[%d]", lang, i)
			if _, err := isRelativePathOrURL(v, keyName); err != nil {
				vr = append(vr, err)
			} else if !parser.disableExternalChecks {
				isImage, err := parser.isImageFile(toCodeHostingURL(v, parser.currentBaseURL), network)
				if !isImage {
					vr = append(vr, newValidationError(
						keyName,
						"'%s' is not an image: %s", v, err.Error(),
					))
				}
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

	if it != nil {
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
	}

	if len(vr) == 0 {
		return nil
	}

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
