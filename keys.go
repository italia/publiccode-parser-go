package publiccode

import (
	"regexp"
	"strings"

	spdxValidator "github.com/r3vit/go-spdx/spdx"
	"github.com/thoas/go-funk"
)

func (p *Parser) decodeBool(key string, boolValue bool) (err error) {
	switch key {
	case "localisation/localisationReady":
		p.PublicCode.Localisation.LocalisationReady = boolValue
	case "it/conforme/lineeGuidaDesign":
		p.PublicCode.It.Conforme.LineeGuidaDesign = boolValue
	case "it/conforme/modelloInteroperabilita":
		p.PublicCode.It.Conforme.ModelloInteroperabilita = boolValue
	case "it/conforme/misureMinimeSicurezza":
		p.PublicCode.It.Conforme.MisureMinimeSicurezza = boolValue
	case "it/conforme/gdpr":
		p.PublicCode.It.Conforme.GDPR = boolValue
	case "it/piattaforme/spid":
		p.PublicCode.It.Piattaforme.Spid = boolValue
	case "it/piattaforme/pagopa":
		p.PublicCode.It.Piattaforme.Pagopa = boolValue
	case "it/piattaforme/cie":
		p.PublicCode.It.Piattaforme.Cie = boolValue
	case "it/piattaforme/anpr":
		p.PublicCode.It.Piattaforme.Anpr = boolValue

	default:
		return ErrorInvalidKey{"Unexpected boolean key: " + key}
	}
	return
}

func (p *Parser) decodeString(key string, value string) (err error) {
	switch {
	case key == "publiccodeYmlVersion":
		// strip legacy URI prefix
		value = strings.Replace(value, "http://w3id.org/publiccode/version/", "", 1)

		if !funk.Contains(SupportedVersions, value) {
			return newErrorInvalidValue(key, "version %s not supported", value)
		}
		p.PublicCode.PubliccodeYamlVersion = value
	case key == "name":
		p.PublicCode.Name = value
	case key == "applicationSuite":
		p.PublicCode.ApplicationSuite = value
	case key == "url":
		p.PublicCode.URLString, p.PublicCode.URL, err = p.checkURL(key, value)
		return err
	case key == "landingURL":
		p.PublicCode.LandingURLString, p.PublicCode.LandingURL, err = p.checkURL(key, value)
		return err
	case key == "isBasedOn":
		return p.decodeArrString(key, []string{value})
	case key == "softwareVersion":
		p.PublicCode.SoftwareVersion = value
	case key == "releaseDate":
		p.PublicCode.ReleaseDateString, p.PublicCode.ReleaseDate, err = p.checkDate(key, value)
		return err
	case key == "logo":
		p.PublicCode.Logo, err = p.checkLogo(key, value)
		return err
	case key == "monochromeLogo":
		p.PublicCode.MonochromeLogo, err = p.checkMonochromeLogo(key, value)
		return err
	case key == "platforms":
		return p.decodeArrString(key, []string{value})
	case key == "tags":
		return p.decodeArrString(key, []string{value})
	case key == "roadmap":
		p.PublicCode.RoadmapString, p.PublicCode.Roadmap, err = p.checkURL(key, value)
		return err
	case key == "developmentStatus":
		for _, v := range []string{"concept", "development", "beta", "stable", "obsolete"} {
			if v == value {
				p.PublicCode.DevelopmentStatus = value
				return nil
			}
		}
		return newErrorInvalidValue(key, "invalid value: %s", value)
	case key == "softwareType":
		// the "standalone" value was deprecated in publiccode.yml 0.2
		if value == "standalone" {
			value = "standalone/other"
		}

		var supportedTypes = []string{"standalone/mobile", "standalone/iot", "standalone/desktop", "standalone/web", "standalone/backend", "standalone/other", "addon", "library", "configurationFiles"}
		if !funk.Contains(supportedTypes, value) {
			return newErrorInvalidValue(key, "invalid value: %s", value)
		}
		p.PublicCode.SoftwareType = value

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
			if len(value) == 0 {
				return newErrorInvalidValue(key, "missing mandatory field")
			}
			if len(value) > 35 {
				return newErrorInvalidValue(key, "too long (%d), max 35 chars", len(value))
			}
			desc.GenericName = value
		}
		if attr == "longDescription" {
			if len(value) < 500 && !p.Strict {
				return newErrorInvalidValue(key, "too short (%d), min 500 chars", len(value))
			}
			if len(value) > 10000 {
				return newErrorInvalidValue(key, "too long (%d), max 10000 chars", len(value))
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
			if len(value) > 150 {
				return newErrorInvalidValue(key, "too long (%d), max 150 chars", len(value))
			}
			desc.ShortDescription = value
		}
		p.PublicCode.Description[lang] = desc
		return nil
	case key == "legal/authorsFile":
		p.PublicCode.Legal.AuthorsFile, err = p.checkFile(key, value)

		// If not running in strict mode we can tolerate this absence.
		if err != nil && p.Strict {
			return err
		}
	case key == "legal/license":
		_, err := spdxValidator.Parse(value)
		if err != nil {
			return newErrorInvalidValue(key, "invalid value %s: %v", value, err)
		}
		p.PublicCode.Legal.License = value
		return nil
	case key == "legal/mainCopyrightOwner":
		p.PublicCode.Legal.MainCopyrightOwner = value
	case key == "legal/repoOwner":
		p.PublicCode.Legal.RepoOwner = value
	case key == "maintenance/type":
		for _, v := range []string{"internal", "contract", "community", "none"} {
			if v == value {
				p.PublicCode.Maintenance.Type = value
				return nil
			}
		}
		return newErrorInvalidValue(key, "invalid value: %s", value)
	case key == "it/countryExtensionVersion":
		if !funk.Contains(ExtensionITSupportedVersions, value) {
			return newErrorInvalidValue(key, "version %s not supported for 'it' extension", value)
		}
		p.PublicCode.It.CountryExtensionVersion = value
	case key == "it/riuso/codiceIPA":
		p.PublicCode.It.Riuso.CodiceIPA, err = p.checkCodiceIPA(key, value)
		if err != nil {
			return err
		}
	default:
		return ErrorInvalidKey{"Unexpected string key: " + key}
	}
	return
}

func (p *Parser) decodeArrString(key string, value []string) error {
	switch {
	case key == "isBasedOn":
		p.PublicCode.IsBasedOn = append(p.PublicCode.IsBasedOn, value...)

	case key == "platforms":
		p.PublicCode.Platforms = append(p.PublicCode.Platforms, value...)

	case key == "categories":
		for _, v := range value {
			v, err := p.checkCategory(key, v)
			if err != nil {
				return err
			}
			p.PublicCode.Categories = append(p.PublicCode.Categories, v)
		}

	case key == "usedBy":
		p.PublicCode.UsedBy = append(p.PublicCode.UsedBy, value...)

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
		if attr == "features" {
			for _, v := range value {
				if len(v) > 100 {
					return newErrorInvalidValue(key, "too long, max 100 chars")

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
		for _, v := range value {
			if err := p.checkMIME(key, v); err != nil {
				return err
			}
			p.PublicCode.InputTypes = append(p.PublicCode.InputTypes, v)
		}

	case key == "outputTypes":
		for _, v := range value {
			if err := p.checkMIME(key, v); err != nil {
				return err
			}
			p.PublicCode.OutputTypes = append(p.PublicCode.OutputTypes, v)
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
					contact.Phone = val.(string)
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
		p.PublicCode.DependsOn.Open = deps

	case "dependsOn/proprietary":
		deps, err := p.checkDependencies(key, value)
		if err != nil {
			return err
		}
		p.PublicCode.DependsOn.Proprietary = deps

	case "dependsOn/hardware":
		deps, err := p.checkDependencies(key, value)
		if err != nil {
			return err
		}
		p.PublicCode.DependsOn.Hardware = deps

	default:
		return ErrorInvalidKey{"Unexpected array key: " + key}
	}
	return nil
}

// finalize do the cross-validation checks.
func (p *Parser) finalize() (es ErrorParseMulti) {
	// description must have at least one language
	if len(p.PublicCode.Description) == 0 {
		es = append(es, newErrorInvalidValue("description", "at least one language is required"))
	}

	// description/[lang]/genericName is mandatory
	for lang, description := range p.PublicCode.Description {
		if description.GenericName == "" {
			es = append(es, newErrorInvalidValue("description/"+lang+"/genericName", "missing mandatory key"))
		}
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
			// If we are not running in strict mode and we can tolerate the absence of this key, skip it
			if !p.Strict && funk.Contains(toleratedMandatoryKeys, k) {
				continue
			}

			es = append(es, newErrorInvalidValue(k, "missing mandatory key"))
		}
	}

	return
}
