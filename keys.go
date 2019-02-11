package publiccode

import (
	"regexp"
	"strings"

	"github.com/thoas/go-funk"
)

var mandatoryKeys = []string{
	"publiccodeYmlVersion",
	"name",
	"url",
	"releaseDate",
	"platforms",
	"tags",
	"softwareType",
	"legal/license",
	"maintenance/type",
	"localisation/localisationReady",
	"localisation/availableLanguages",
}

func (p *Parser) decodeBool(key string, boolValue bool) (err error) {
	switch key {
	case "localisation/localisationReady":
		p.PublicCode.Localisation.LocalisationReady = boolValue
	case "it/conforme/accessibile":
		p.PublicCode.It.Conforme.Accessibile = boolValue
	case "it/conforme/interoperabile":
		p.PublicCode.It.Conforme.Interoperabile = boolValue
	case "it/conforme/sicuro":
		p.PublicCode.It.Conforme.Sicuro = boolValue
	case "it/conforme/privacy":
		p.PublicCode.It.Conforme.Privacy = boolValue
	case "it/spid":
		p.PublicCode.It.Spid = boolValue
	case "it/pagopa":
		p.PublicCode.It.Pagopa = boolValue
	case "it/cie":
		p.PublicCode.It.Cie = boolValue
	case "it/anpr":
		p.PublicCode.It.Anpr = boolValue
	case "it/designKit/seo":
		p.PublicCode.It.DesignKit.Seo = boolValue
	case "it/designKit/ui":
		p.PublicCode.It.DesignKit.UI = boolValue
	case "it/designKit/web":
		p.PublicCode.It.DesignKit.Web = boolValue
	case "it/designKit/content":
		p.PublicCode.It.DesignKit.Content = boolValue

	default:
		return ErrorInvalidKey{key + " : Boolean"}
	}
	return
}

func (p *Parser) decodeString(key string, value string) (err error) {
	switch {
	case key == "publiccodeYmlVersion":
		// strip legacy URI prefix
		value = strings.Replace(value, "http://w3id.org/publiccode/version/", "", 1)

		p.PublicCode.PubliccodeYamlVersion = value
		if p.PublicCode.PubliccodeYamlVersion != Version {
			return newErrorInvalidValue(key, "version %s not supported", p.PublicCode.PubliccodeYamlVersion)
		}
	case key == "name":
		p.PublicCode.Name = value
	case key == "applicationSuite":
		p.PublicCode.ApplicationSuite = value
	case key == "url":
		p.PublicCode.URL, err = p.checkURL(key, value)
		return err
	case key == "landingURL":
		p.PublicCode.LandingURL, err = p.checkURL(key, value)
		return err
	case key == "isBasedOn":
		return p.decodeArrString(key, []string{value})
	case key == "softwareVersion":
		p.PublicCode.SoftwareVersion = value
	case key == "releaseDate":
		p.PublicCode.ReleaseDate, err = p.checkDate(key, value)
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
		p.PublicCode.Roadmap, err = p.checkURL(key, value)
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
		for _, v := range []string{"standalone", "addon", "library", "configurationFiles"} {
			if v == value {
				p.PublicCode.SoftwareType = value
				return nil
			}
		}
		return newErrorInvalidValue(key, "invalid value: %s", value)
	case regexp.MustCompile(`^description/[a-z]{3}`).MatchString(key):
		if p.PublicCode.Description == nil {
			p.PublicCode.Description = make(map[string]Desc)
		}
		k := strings.Split(key, "/")[1]
		attr := strings.Split(key, "/")[2]
		var desc = p.PublicCode.Description[k]
		if attr == "localisedName" {
			desc.LocalisedName = value
			p.PublicCode.Description[k] = desc
		}
		if attr == "genericName" {
			if len(value) == 0 || len(value) > 35 {
				return newErrorInvalidValue(key, "\"%s\" has an invalid number of characters: %d.  (mandatory and max 35 chars)", key, len(value))
			}
			desc.GenericName = value
			p.PublicCode.Description[k] = desc
		}
		if attr == "longDescription" {
			if len(value) < 500 || len(value) > 10000 {
				return newErrorInvalidValue(key, "\"%s\" has an invalid number of characters: %d.  (min 500 chars, max 10.000 chars)", key, len(value))
			}
			desc.LongDescription = value
			p.PublicCode.Description[k] = desc
		}
		if attr == "documentation" {
			desc.Documentation, err = p.checkURL(key, value)
			if err != nil {
				return err
			}
			p.PublicCode.Description[k] = desc
		}
		if attr == "apiDocumentation" {
			desc.APIDocumentation, err = p.checkURL(key, value)
			if err != nil {
				return err
			}
			p.PublicCode.Description[k] = desc
		}
		if attr == "shortDescription" {
			if len(value) > 150 {
				return newErrorInvalidValue(key, "\"%s\" has an invalid number of characters: %d.  (max 150 chars)", key, len(value))
			}
			desc.ShortDescription = value
			p.PublicCode.Description[k] = desc
		}
		return p.checkLanguageCodes3(key, k)
	case key == "legal/authorsFile":
		p.PublicCode.Legal.AuthorsFile, err = p.checkFile(key, value)
		return err
	case key == "legal/license":
		p.PublicCode.Legal.License = value
		return p.checkSpdx(key, value)
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
	case key == "it/riuso/codiceIPA":
		p.PublicCode.It.Riuso.CodiceIPA, err = p.checkCodiceIPA(key, value)
		if err != nil {
			return err
		}
	default:
		return ErrorInvalidKey{key + " : String"}
	}
	return
}

func (p *Parser) decodeArrString(key string, value []string) error {
	switch {
	case key == "isBasedOn":
		p.PublicCode.IsBasedOn = append(p.PublicCode.IsBasedOn, value...)

	case key == "platforms":
		p.PublicCode.Platforms = append(p.PublicCode.Platforms, value...)

	case key == "tags":
		for _, v := range value {
			v, err := p.checkTag(key, v)
			if err != nil {
				return err
			}
			p.PublicCode.Tags = append(p.PublicCode.Tags, v)
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

	case key == "intendedAudience/onlyFor":
		for _, v := range value {
			v, err := p.checkPaTypes(key, v)
			if err != nil {
				return err
			}
			p.PublicCode.IntendedAudience.OnlyFor = append(p.PublicCode.IntendedAudience.OnlyFor, v)
		}

	case regexp.MustCompile(`^description/[a-z]{3}`).MatchString(key):
		if p.PublicCode.Description == nil {
			p.PublicCode.Description = make(map[string]Desc)
		}
		k := strings.Split(key, "/")[1]
		attr := strings.Split(key, "/")[2]
		var desc = p.PublicCode.Description[k]
		if attr == "awards" {
			desc.Awards = append(desc.Awards, value...)
			p.PublicCode.Description[k] = desc
		}
		if attr == "freeTags" {
			desc.FreeTags = append(desc.FreeTags, value...)
			p.PublicCode.Description[k] = desc
		}
		if attr == "features" {
			for _, v := range value {
				if len(v) > 100 {
					return newErrorInvalidValue(key, " %s is too long.  (max 100 chars)", key)

				}
				desc.Features = append(desc.Features, v)
			}
			p.PublicCode.Description[k] = desc
		}
		if attr == "screenshots" {
			for _, v := range value {
				i, err := p.checkImage(key, v)
				if err != nil {
					return err
				}
				desc.Screenshots = append(desc.Screenshots, i)
			}
			p.PublicCode.Description[k] = desc
		}
		if attr == "videos" {
			for _, v := range value {
				v, err := p.checkOembed(key, v)
				if err != nil {
					return err
				}
				desc.Videos = append(desc.Videos, v)
			}
			p.PublicCode.Description[k] = desc
		}
		return p.checkLanguageCodes3(key, k)

	case key == "localisation/availableLanguages":
		for _, v := range value {
			if err := p.checkLanguageCodes3(key, v); err != nil {
				return err
			}
			p.PublicCode.Localisation.AvailableLanguages = append(p.PublicCode.Localisation.AvailableLanguages, v)
		}

	case key == "it/ecosistemi":
		for _, v := range value {
			ecosistemi := []string{"sanita", "welfare", "finanza-pubblica", "scuola", "istruzione-superiore-ricerca",
				"difesa-sicurezza-soccorso-legalita", "giustizia", "infrastruttura-logistica", "sviluppo-sostenibilita",
				"beni-culturali-turismo", "agricoltura", "italia-europa-mondo"}

			if !funk.Contains(ecosistemi, v) {
				return newErrorInvalidValue(key, "unknown it/ecosistemi: %s", v)
			}
			p.PublicCode.It.Ecosistemi = append(p.PublicCode.It.Ecosistemi, v)
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
		return ErrorInvalidKey{key + " : Array of Strings"}

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
				} else if k.(string) == "until" {
					date, err := p.checkDate(key, val.(string))
					if err != nil {
						return err
					}
					contractor.Until = date
				} else if k.(string) == "website" {
					u, err := p.checkURL(key, val.(string))
					if err != nil {
						return err
					}
					contractor.Website = u
				} else {
					return newErrorInvalidValue(key, " %s contains an invalid value", k)
				}
			}
			if contractor.Name == "" {
				return newErrorInvalidValue(key, " name field is mandatory.")
			}
			if contractor.Until.IsZero() {
				return newErrorInvalidValue(key, " until field is mandatory.")
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
					return newErrorInvalidValue(key, " %s contains an invalid value", k)
				}
			}
			if contact.Name == "" {
				return newErrorInvalidValue(key, " name field is mandatory.")
			}

			p.PublicCode.Maintenance.Contacts = append(p.PublicCode.Maintenance.Contacts, contact)
		}

	case "dependsOn/open":
		for _, v := range value {
			var dep Dependency

			for k, val := range v.(map[interface{}]interface{}) {
				if k.(string) == "name" {
					dep.Name = val.(string)
				} else if k.(string) == "optional" {
					dep.Optional = val.(bool)
				} else if k.(string) == "version" {
					dep.Version = val.(string)
				} else if k.(string) == "versionMin" {
					dep.VersionMin = val.(string)
				} else if k.(string) == "versionMax" {
					dep.VersionMax = val.(string)
				} else {
					return newErrorInvalidValue(key, " %s contains an invalid value", k)
				}
			}
			if dep.Name == "" {
				return newErrorInvalidValue(key, " name field is mandatory.")
			}

			p.PublicCode.DependsOn.Open = append(p.PublicCode.DependsOn.Open, dep)
		}

	case "dependsOn/proprietary":
		for _, v := range value {
			var dep Dependency

			for k, val := range v.(map[interface{}]interface{}) {
				if k.(string) == "name" {
					dep.Name = val.(string)
				} else if k.(string) == "optional" {
					dep.Optional = val.(bool)
				} else if k.(string) == "version" {
					dep.Version = val.(string)
				} else if k.(string) == "versionMin" {
					dep.VersionMin = val.(string)
				} else if k.(string) == "versionMax" {
					dep.VersionMax = val.(string)
				} else {
					return newErrorInvalidValue(key, " %s contains an invalid value", k)
				}
			}
			if dep.Name == "" {
				return newErrorInvalidValue(key, " name field is mandatory.")
			}

			p.PublicCode.DependsOn.Proprietary = append(p.PublicCode.DependsOn.Proprietary, dep)
		}

	case "dependsOn/hardware":
		for _, v := range value {
			var dep Dependency

			for k, val := range v.(map[interface{}]interface{}) {
				if k.(string) == "name" {
					dep.Name = val.(string)
				} else if k.(string) == "optional" {
					dep.Optional = val.(bool)
				} else if k.(string) == "version" {
					dep.Version = val.(string)
				} else if k.(string) == "versionMin" {
					dep.VersionMin = val.(string)
				} else if k.(string) == "versionMax" {
					dep.VersionMax = val.(string)
				} else {
					return newErrorInvalidValue(key, " %s contains an invalid value", k)
				}
			}
			if dep.Name == "" {
				return newErrorInvalidValue(key, " name field is mandatory.")
			}

			p.PublicCode.DependsOn.Hardware = append(p.PublicCode.DependsOn.Hardware, dep)
		}

	default:
		return ErrorInvalidKey{key + " : Array of Objects"}
	}
	return nil
}

// finalize do the cross-validation checks.
func (p *Parser) finalize() (es ErrorParseMulti) {
	// description must have at least one language
	if len(p.PublicCode.Description) == 0 {
		es = append(es, newErrorInvalidValue("description", "must have at least one language."))
	}

	// description/[lang]/genericName is mandatory
	for lang, description := range p.PublicCode.Description {
		if description.GenericName == "" {
			es = append(es, newErrorInvalidValue("description/"+lang+"/genericName", "must have GenericName key."))
		}
	}

	// "maintenance/contractors" presence is mandatory (if maintainance/type is contract).
	if p.PublicCode.Maintenance.Type == "contract" && len(p.PublicCode.Maintenance.Contractors) == 0 {
		es = append(es, newErrorInvalidValue("maintenance/contractors", "not found, mandatory for \"contract\" maintenance"))
	}

	// "maintenance/contractors" presence is mandatory (if maintainance/type is internal or community).
	if (p.PublicCode.Maintenance.Type == "internal" || p.PublicCode.Maintenance.Type == "community") && len(p.PublicCode.Maintenance.Contacts) == 0 {
		es = append(es, newErrorInvalidValue("maintenance/contacts", "not found, mandatory for \"internal\" or \"community\" maintenance"))
	}

	// maintenance/contacts/name is always mandatory
	if len(p.PublicCode.Maintenance.Contacts) > 0 {
		for _, c := range p.PublicCode.Maintenance.Contacts {
			if c.Name == "" {
				es = append(es, newErrorInvalidValue("maintenance/contacts/name", "not found. It's mandatory."))
			}
		}
	}
	// maintenance/contractors/name is always mandatory
	if len(p.PublicCode.Maintenance.Contractors) > 0 {
		for _, c := range p.PublicCode.Maintenance.Contractors {
			if c.Name == "" {
				es = append(es, newErrorInvalidValue("maintenance/contractors/name", "not found. It's mandatory."))
			}
		}
	}
	// maintenance/contractors/until is always mandatory
	if len(p.PublicCode.Maintenance.Contractors) > 0 {
		for _, c := range p.PublicCode.Maintenance.Contractors {
			if c.Until.IsZero() {
				es = append(es, newErrorInvalidValue("maintenance/contractors/until", "not found. It's mandatory."))
			}
		}
	}

	// mandatoryKeys check
	for k, v := range p.missing {
		if v {
			es = append(es, newErrorInvalidValue(k, k+" is a mandatory key."))
		}
	}

	return
}
