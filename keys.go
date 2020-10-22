package publiccode

import (
	"fmt"
	"net/url"
	"strings"

	spdxValidator "github.com/alranel/go-spdx/spdx"
	"github.com/alranel/go-vcsurl"
	"github.com/rivo/uniseg"
	"github.com/thoas/go-funk"
)

func (p *Parser) validateFields(publiccode PublicCode) (err error) {
	// strip legacy URI prefix
	publiccode.PubliccodeYamlVersion = strings.Replace(publiccode.PubliccodeYamlVersion, "http://w3id.org/publiccode/version/", "", 1)
	if !funk.Contains(SupportedVersions, publiccode.PubliccodeYamlVersion) {
		return newErrorInvalidValue("version %s not supported", publiccode.PubliccodeYamlVersion)
	}

	// Check that the supplied URL is valid and exists.
	err = p.checkURL(p.PublicCode.URL)
	if err != nil {
		return err
	}
	// Check that the supplied URL points to a repository.
	if !vcsurl.IsRepo(p.PublicCode.URL) {
		return fmt.Errorf("invalid repository URL: %s", p.PublicCode.URL)
	}
	// Check that the supplied URL matches the source repository, if known.
	if p.RemoteBaseURL != "" {
		url1, err := url.Parse(p.RemoteBaseURL)
		if err != nil {
			return err
		}
		repo1 := vcsurl.GetRepo(url1)
		repo2 := vcsurl.GetRepo(publiccode.URL)
		if repo1 == nil {
			return fmt.Errorf("failed to detect repo for remote-base-url: %s\n", url1.String())
		}
		if repo2 == nil {
			return fmt.Errorf("failed to detect repo for %s\n", publiccode.URL)
		}

		// Let's ignore the schema when checking for equality.
		//
		// This is mainly to match repos regardless of whether they are served
		// through HTTPS or HTTP.
		repo1.Scheme, repo2.Scheme = "", ""

		if !strings.EqualFold(repo1.String(), repo2.String()) {
			return fmt.Errorf(
				"declared url (%s) and actual publiccode.yml source URL (%s) "+
				"are not in the same repo: '%s' vs '%s'",
				publiccode.URL, p.RemoteBaseURL, repo2, repo1)
		}
	}

	if p.PublicCode.LandingURL != nil {
		err = p.checkURL(p.PublicCode.LandingURL)
		return err
	}

	// TODO Check for nil
	for _, c := range p.PublicCode.Categories {
		if !p.isCategory(c) {
			return fmt.Errorf("unknown category: %s", c)
		}
	}

	p.PublicCode.Logo, err = p.checkLogo(p.PublicCode.Logo)
	if err != nil {
		return err
	}
	
	p.PublicCode.MonochromeLogo, err = p.checkMonochromeLogo(p.PublicCode.MonochromeLogo)
	if err != nil {
		return err
	}

	// TODO err = p.checkURL(value)
	// if err != nill {
	// 	return err
	// }

	if !funk.Contains([]string{"concept", "development", "beta", "stable", "obsolete"}, p.PublicCode.DevelopmentStatus) {
		return fmt.Errorf("invalid developmentStatus: %s", p.PublicCode.DevelopmentStatus)
	}

	// the "standalone" value was deprecated in publiccode.yml 0.2
	if p.PublicCode.SoftwareType == "standalone" {
		p.PublicCode.SoftwareType = "standalone/other"
	}
	var supportedTypes = []string{"standalone/mobile", "standalone/iot", "standalone/desktop", "standalone/web", "standalone/backend", "standalone/other", "addon", "library", "configurationFiles"}
	if !funk.Contains(supportedTypes, p.PublicCode.SoftwareType) {
		return fmt.Errorf("invalid softwareType: %s", p.PublicCode.SoftwareType)
	}

	p.PublicCode.Legal.AuthorsFile, err = p.checkFile(p.PublicCode.Legal.AuthorsFile)
	// If not running in strict mode we can tolerate this absence.
	if err != nil && p.Strict {
		return err
	}

	_, err = spdxValidator.Parse(p.PublicCode.Legal.License)
	if err != nil {
		return fmt.Errorf("invalid license %s: %v", p.PublicCode.Legal.License, err)
	}

	if !funk.Contains([]string{"internal", "contract", "community", "none"}, p.PublicCode.Maintenance.Type) {
		return fmt.Errorf("invalid maintenanceType: %s", p.PublicCode.Maintenance.Type)
	}

	if !funk.Contains(ExtensionITSupportedVersions, p.PublicCode.It.CountryExtensionVersion) {
		return fmt.Errorf("version %s not supported for 'it' extension", p.PublicCode.It.CountryExtensionVersion)
	}

	p.PublicCode.It.Riuso.CodiceIPA, err = p.checkCodiceIPA(p.PublicCode.It.Riuso.CodiceIPA)
	if err != nil {
		return err
	}

	for _, scope := range p.PublicCode.IntendedAudience.Scope {
		if ! p.isScope(scope) {
			return fmt.Errorf("not a scope %s", scope)
		}
	}

	if p.Strict {
		for _, mimeType := range p.PublicCode.InputTypes {
			if !p.isMIME(mimeType) {
				return fmt.Errorf("not a MIME type: %s", mimeType)
			}
		}
		for _, mimeType := range p.PublicCode.OutputTypes {
			if !p.isMIME(mimeType) {
				return fmt.Errorf("not a MIME type: %s", mimeType)
			}
		}
	}

	for lang, desc := range p.PublicCode.Description {
		if p.PublicCode.Description == nil {
			p.PublicCode.Description = make(map[string]Desc)
		}

		// check lang validity and canonicalize it
		// XXX TODO lang, err = p.checkLanguageCode(lang)
		// if err != nil {
		// 	return err
		// }

		length := uniseg.GraphemeClusterCount(desc.GenericName)
		if length == 0 {
			return fmt.Errorf("genericName missing")
		}
		if length > 35 {
			return fmt.Errorf("genericName too long (%d), max 35 chars", length)
		}

		length = uniseg.GraphemeClusterCount(desc.LongDescription)
		if length < 500 {
			return fmt.Errorf("longDescription too short (%d), min 500 chars", length)
		}
		if length > 10000 {
			return fmt.Errorf("longDescription too long (%d), max 10000 chars", length)
		}

		err = p.checkURL(desc.Documentation)
		if err != nil {
			return err
		}
		err = p.checkURL(desc.APIDocumentation)
		if err != nil {
			return err
		}

		length = uniseg.GraphemeClusterCount(desc.ShortDescription)
		if length > 150 {
			return fmt.Errorf("shortDescription too long (%d), max 150 chars", length)
		}
		
		// XXX if attr == "features" || (attr == "featureList" && !p.Strict) {
		if !p.Strict {
			for _, feature := range desc.Features {
				length := uniseg.GraphemeClusterCount(feature)
				if length > 100 {
					return fmt.Errorf("feature too long (%d), max 100 chars", length)
				}
			}
		}
		for _, v := range desc.Screenshots {
			// TODO i, err := p.checkImage(v)
			if err != nil {
				return err
			}
		}
		for _, v := range desc.Videos {
			// TODO v, u, err := p.checkOembed(v)
			if err != nil {
				return err
			}
		}
	}
	return
}

// finalize do the cross-validation checks.
func (p *Parser) validate() (es ErrorParseMulti) {
	// "maintenance/contractors" presence is mandatory (if maintainance/type is contract).
	if p.PublicCode.Maintenance.Type == "contract" && len(p.PublicCode.Maintenance.Contractors) == 0 {
		es = append(es, newErrorInvalidValue("maintenance/contractors", "missing but mandatory for \"contract\" maintenance"))
	}

	// "maintenance/contractors" presence is mandatory (if maintainance/type is internal or community).
	if (p.PublicCode.Maintenance.Type == "internal" || p.PublicCode.Maintenance.Type == "community") && len(p.PublicCode.Maintenance.Contacts) == 0 {
		es = append(es, newErrorInvalidValue("maintenance/contacts", "missing but mandatory for \"internal\" or \"community\" maintenance"))
	}

	// TODO version // If this is not the latest version, skip mandatority check for some keys
	// if p.PublicCode.PubliccodeYamlVersion == "0.1" && k == "categories" {
	// 	continue
	// }
	return
}
