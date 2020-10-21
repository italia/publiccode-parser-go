package publiccode

import (
	"fmt"
	"net/url"
	"regexp"
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
		return newErrorInvalidValue(key, "invalid repository URL: %s", value)
	}
	// Check that the supplied URL matches the source repository, if known.
	if p.RemoteBaseURL != "" {
		url1, err := url.Parse(p.RemoteBaseURL)
		if err != nil {
			return err
		}
		url2, err := url.Parse(value)
		if err != nil {
			return err
		}
		repo1 := vcsurl.GetRepo(url1)
		repo2 := vcsurl.GetRepo(url2)
		if repo1 == nil {
			return fmt.Errorf("failed to detect repo for remote-base-url: %s\n", url1.String())
		}
		if repo2 == nil {
			return newErrorInvalidValue(key, "failed to detect repo for %s\n", url2.String())
		}

		// Let's ignore the schema when checking for equality.
		//
		// This is mainly to match repos regardless of whether they are served
		// through HTTPS or HTTP.
		repo1.Scheme, repo2.Scheme = "", ""

		if !strings.EqualFold(repo1.String(), repo2.String()) {
			return newErrorInvalidValue(
				key,
				"declared url (%s) and actual publiccode.yml source URL (%s) "+
				"are not in the same repo: '%s' vs '%s'",
				value, p.RemoteBaseURL, repo2, repo1)
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

	err = p.checkURL(value)
	if err != nill {
		return err
	}

	if !funk.Contains([]string{"concept", "development", "beta", "stable", "obsolete"}, p.PublicCode.DevelopmentStatus) {
		return fmt.Errorf("invalid developmentStatus: %s", p.PublicCode.DevelopmentStatus)
	}

	// the "standalone" value was deprecated in publiccode.yml 0.2
	if p.PublicCode.SoftwareType == "standalone" {
		p.PublicCode.SoftwareType = "standalone/other"
	}
	var supportedTypes = []string{"standalone/mobile", "standalone/iot", "standalone/desktop", "standalone/web", "standalone/backend", "standalone/other", "addon", "library", "configurationFiles"}
	if !funk.Contains(supportedTypes, value) {
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

	if !funk.Contains([]string{"internal", "contract", "community", "none"}, p.Publiccode.Maintenance.Type) {
		return fmt.Errorf("invalid maintenanceType: %s", p.PublicCode.Maintenance.Type)
	}

	if !funk.Contains(ExtensionITSupportedVersions, p.Publiccode.It.CcountryExtensionVersion) {
		return fmt.Errorf("version %s not supported for 'it' extension", p.Publiccode.It.CcountryExtensionVersion)
	}

	p.PublicCode.It.Riuso.CodiceIPA, err = p.checkCodiceIPA(p.PublicCode.It.Riuso.CodiceIPA)
	if err != nil {
		return err
	}

	switch {
	case regexp.MustCompile(`^description/.+/`).MatchString(key):
		if p.PublicCode.Description == nil {
			p.PublicCode.Description = make(map[string]Desc)
		}
		lang := strings.Split(key, "/")[1]
		attr := strings.Split(key, "/")[2]

		// check lang validity and canonicalize it
		lang, err = p.checkLanguageCode(key, lang)
		if err != nil {
			return err
		}

		desc := p.PublicCode.Description[lang]
		if attr == "localisedName" {
			desc.LocalisedName = value
		}
		if attr == "genericName" {
			length := uniseg.GraphemeClusterCount(value)
			if length == 0 {
				return newErrorInvalidValue(key, "missing mandatory field")
			}
			if length > 35 {
				return newErrorInvalidValue(key, "too long (%d), max 35 chars", length)
			}
			desc.GenericName = value
		}
		if attr == "longDescription" {
			length := uniseg.GraphemeClusterCount(value)
			if length < 500 {
				return newErrorInvalidValue(key, "too short (%d), min 500 chars", length)
			}
			if length > 10000 {
				return newErrorInvalidValue(key, "too long (%d), max 10000 chars", length)
			}
			desc.LongDescription = value
		}
		if attr == "documentation" {
			desc.DocumentationString, desc.Documentation, err = p.checkURL(key, value)
			if err != nil {
				return err
			}
		}
		if attr == "apiDocumentation" {
			desc.APIDocumentationString, desc.APIDocumentation, err = p.checkURL(key, value)
			if err != nil {
				return err
			}
		}
		if attr == "shortDescription" {
			length := uniseg.GraphemeClusterCount(value)
			if length > 150 {
				return newErrorInvalidValue(key, "too long (%d), max 150 chars", length)
			}
			desc.ShortDescription = value
		}
		p.PublicCode.Description[lang] = desc
		return nil
	}
}

func (p *Parser) decodeArrString(key string, value []string) error {
	switch {

	case key == "intendedAudience/countries":
		for _, v := range value {
			if err := p.checkCountryCodes2(key, v); err != nil {
				return err
			}
			p.PublicCode.IntendedAudience.Countries = append(p.PublicCode.IntendedAudience.Countries, v)
		}

	case key == "intendedAudience/unsupportedCountries":
		for _, v := range value {
			if err := p.checkCountryCodes2(key, v); err != nil {
				return err
			}
			p.PublicCode.IntendedAudience.UnsupportedCountries = append(p.PublicCode.IntendedAudience.UnsupportedCountries, v)
		}

	case key == "intendedAudience/scope":
		for _, v := range value {
			v, err := p.checkScope(key, v)
			if err != nil {
				return err
			}
			p.PublicCode.IntendedAudience.Scope = append(p.PublicCode.IntendedAudience.Scope, v)
		}

	case regexp.MustCompile(`^description/.+/`).MatchString(key):
		if p.PublicCode.Description == nil {
			p.PublicCode.Description = make(map[string]Desc)
		}
		lang := strings.Split(key, "/")[1]
		attr := strings.Split(key, "/")[2]

		// check lang validity and canonicalize it
		lang, err := p.checkLanguageCode(key, lang)
		if err != nil {
			return err
		}

		desc := p.PublicCode.Description[lang]
		if attr == "awards" {
			desc.Awards = append(desc.Awards, value...)
		}
		if attr == "features" || (attr == "featureList" && !p.Strict) {
			for _, v := range value {
				length := uniseg.GraphemeClusterCount(v)
				if length > 100 {
					return newErrorInvalidValue(key, "too long (%d), max 100 chars", length)

				}
				desc.Features = append(desc.Features, v)
			}
		}
		if attr == "screenshots" {
			for _, v := range value {
				i, err := p.checkImage(key, v)
				if err != nil {
					return err
				}
				desc.Screenshots = append(desc.Screenshots, i)
			}
		}
		if attr == "videos" {
			for _, v := range value {
				v, u, err := p.checkOembed(key, v)
				if err != nil {
					return err
				}
				desc.Videos = append(desc.Videos, u)
				desc.VideosStrings = append(desc.VideosStrings, v)
			}
		}

		p.PublicCode.Description[lang] = desc
		return nil

	case key == "localisation/availableLanguages":
		for _, lang := range value {
			// check language and canonicalize it
			lang, err := p.checkLanguageCode(key, lang)
			if err != nil {
				return err
			}
			p.PublicCode.Localisation.AvailableLanguages = append(p.PublicCode.Localisation.AvailableLanguages, lang)
		}

	case key == "inputTypes":
		if p.Strict {
			for _, v := range value {
				if err := p.checkMIME(key, v); err != nil {
					return err
				}
				p.PublicCode.InputTypes = append(p.PublicCode.InputTypes, v)
			}
		}

	case key == "outputTypes":
		if p.Strict {
			for _, v := range value {
				if err := p.checkMIME(key, v); err != nil {
					return err
				}
				p.PublicCode.OutputTypes = append(p.PublicCode.OutputTypes, v)
			}
		}

	default:
		return ErrorInvalidKey{"Unexpected array key: " + key}

	}
	return nil
}

func (p *Parser) decodeArrObj(key string, value map[interface{}]interface{}) error {
	switch key {
	case "maintenance/contractors":
		for _, v := range value {
			var contractor Contractor

			for k, val := range v.(map[interface{}]interface{}) {
				if k.(string) == "name" {
					contractor.Name = val.(string)
				} else if k.(string) == "email" {
					err := p.checkEmail(key, val.(string))
					if err != nil {
						return err
					}
					contractor.Email = val.(string)
				} else if k.(string) == "until" {
					var err error
					contractor.UntilString, contractor.Until, err = p.checkDate(key, val.(string))
					if err != nil {
						return err
					}
				} else if k.(string) == "website" {
					var err error
					contractor.WebsiteString, contractor.Website, err = p.checkURL(key, val.(string))
					if err != nil {
						return err
					}
				} else {
					return newErrorInvalidValue(key, "invalid value for '%s", k)
				}
			}
			if contractor.Name == "" {
				return newErrorInvalidValue(key, "missing mandatory key 'name'")
			}
			if contractor.Until.IsZero() {
				return newErrorInvalidValue(key, "missing mandatory key 'until'")
			}
			p.PublicCode.Maintenance.Contractors = append(p.PublicCode.Maintenance.Contractors, contractor)
		}

	case "maintenance/contacts":
		for _, v := range value {
			var contact Contact

			for k, val := range v.(map[interface{}]interface{}) {
				if k.(string) == "name" {
					contact.Name = val.(string)
				} else if k.(string) == "email" {
					err := p.checkEmail(key, val.(string))
					if err != nil {
						return err
					}
					contact.Email = val.(string)
				} else if k.(string) == "phone" {
					contact.Phone = fmt.Sprint(val) //forcing cast to string, phone number must always be a string
				} else if k.(string) == "affiliation" {
					contact.Affiliation = val.(string)
				} else {
					return newErrorInvalidValue(key, "invalid value for '%s'", k)
				}
			}
			if contact.Name == "" {
				return newErrorInvalidValue(key, "missing mandatory key 'name'")
			}
			p.PublicCode.Maintenance.Contacts = append(p.PublicCode.Maintenance.Contacts, contact)
		}

	case "dependsOn/open":
		deps, err := p.checkDependencies(key, value)
		if err != nil {
			return err
		}
		p.PublicCode.DependsOn.Open = append(p.PublicCode.DependsOn.Open, deps...)

	case "dependsOn/proprietary":
		deps, err := p.checkDependencies(key, value)
		if err != nil {
			return err
		}
		p.PublicCode.DependsOn.Proprietary = append(p.PublicCode.DependsOn.Proprietary, deps...)

	case "dependsOn/hardware":
		deps, err := p.checkDependencies(key, value)
		if err != nil {
			return err
		}
		p.PublicCode.DependsOn.Hardware = append(p.PublicCode.DependsOn.Hardware, deps...)

	default:
		return ErrorInvalidKey{"Unexpected array key: " + key}
	}
	return nil
}

// finalize do the cross-validation checks.
func (p *Parser) validate() (es ErrorParseMulti) {
	// description must have at least one language
	if len(p.PublicCode.Description) == 0 {
		es = append(es, newErrorInvalidValue("description", "at least one language is required"))
	}

	// description/[lang]/{genericName,shortDescription} are mandatory for all languages
	// description/[lang]/{features,longDescription} are mandatory for at least one language
	haveFeatures := false
	haveLongDescription := false
	for lang, description := range p.PublicCode.Description {
		if description.GenericName == "" {
			es = append(es, newErrorInvalidValue("description/"+lang+"/genericName", "missing mandatory key"))
		}
		if description.ShortDescription == "" {
			es = append(es, newErrorInvalidValue("description/"+lang+"/shortDescription", "missing mandatory key"))
		}
		if len(description.Features) > 0 {
			haveFeatures = true
		}
		if description.LongDescription != "" {
			haveLongDescription = true
		}
	}
	if haveFeatures == false {
		es = append(es, newErrorInvalidValue("description/*/features", "missing mandatory key"))
	}
	if haveLongDescription == false {
		es = append(es, newErrorInvalidValue("description/*/longDescription", "missing mandatory key"))
	}

	// "maintenance/contractors" presence is mandatory (if maintainance/type is contract).
	if p.PublicCode.Maintenance.Type == "contract" && len(p.PublicCode.Maintenance.Contractors) == 0 {
		es = append(es, newErrorInvalidValue("maintenance/contractors", "missing but mandatory for \"contract\" maintenance"))
	}

	// "maintenance/contractors" presence is mandatory (if maintainance/type is internal or community).
	if (p.PublicCode.Maintenance.Type == "internal" || p.PublicCode.Maintenance.Type == "community") && len(p.PublicCode.Maintenance.Contacts) == 0 {
		es = append(es, newErrorInvalidValue("maintenance/contacts", "missing but mandatory for \"internal\" or \"community\" maintenance"))
	}

	// maintenance/contacts/name is always mandatory
	if len(p.PublicCode.Maintenance.Contacts) > 0 {
		for _, c := range p.PublicCode.Maintenance.Contacts {
			if c.Name == "" {
				es = append(es, newErrorInvalidValue("maintenance/contacts/name", "missing mandatory key"))
			}
		}
	}
	// maintenance/contractors/name is always mandatory
	if len(p.PublicCode.Maintenance.Contractors) > 0 {
		for _, c := range p.PublicCode.Maintenance.Contractors {
			if c.Name == "" {
				es = append(es, newErrorInvalidValue("maintenance/contractors/name", "missing mandatory key"))
			}
		}
	}
	// maintenance/contractors/until is always mandatory
	if len(p.PublicCode.Maintenance.Contractors) > 0 {
		for _, c := range p.PublicCode.Maintenance.Contractors {
			if c.Until.IsZero() {
				es = append(es, newErrorInvalidValue("maintenance/contractors/until", "missing mandatory key"))
			}
		}
	}

	// mandatoryKeys check
	for k, v := range p.missing {
		// If this is not the latest version, skip mandatority check for some keys
		if p.PublicCode.PubliccodeYamlVersion == "0.1" && k == "categories" {
			continue
		}

		if v {
			es = append(es, newErrorInvalidValue(k, "missing mandatory key"))
		}
	}

	return
}
