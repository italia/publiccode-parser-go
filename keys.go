package publiccode

import (
	"regexp"
	"strings"
)

var mandatoryKeys = []string{
	"publiccode-yaml-version",
	"name",
	"url",
	"softwareVersion",
	"releaseDate",
	"platforms",
	"tags",
	"softwareType/type",
	"legal/license",
	"legal/repoOwner",
	"maintenance/type",
	"maintenance/contacts",
	"localisation/localisationReady",
	"localisation/availableLanguages",
}

func (p *parser) decodeBool(key string, boolValue bool) (err error) {
	switch key {
	case "localisation/localisationReady":
		p.pc.Localisation.LocalisationReady = boolValue
	case "ita/spid":
		p.pc.Ita.Spid = boolValue
	case "ita/pagopa":
		p.pc.Ita.Pagopa = boolValue
	case "ita/cie":
		p.pc.Ita.Cie = boolValue
	case "ita/anpr":
		p.pc.Ita.Anpr = boolValue
	case "ita/designKit/seo":
		p.pc.Ita.DesignKit.Seo = boolValue
	case "ita/designKit/ui":
		p.pc.Ita.DesignKit.UI = boolValue
	case "ita/designKit/web":
		p.pc.Ita.DesignKit.Web = boolValue
	case "ita/designKit/content":
		p.pc.Ita.DesignKit.Content = boolValue
	case "ita/accessibile":
		p.pc.Ita.Accessibile = boolValue
	case "ita/interoperabile":
		p.pc.Ita.Interoperabile = boolValue
	case "ita/privacy":
		p.pc.Ita.Privacy = boolValue
	case "ita/sicurezza":
		p.pc.Ita.Sicurezza = boolValue

	default:
		return ErrorInvalidKey{key + " : Boolean"}
	}
	return
}

func (p *parser) decodeString(key string, value string) (err error) {
	switch {
	case key == "publiccode-yaml-version":
		p.pc.PubliccodeYamlVersion = value
		if p.pc.PubliccodeYamlVersion != Version {
			return newErrorInvalidValue(key, "version %s not supported", p.pc.PubliccodeYamlVersion)
		}
	case key == "name":
		p.pc.Name = value
	case key == "applicationSuite":
		p.pc.ApplicationSuite = value
	case key == "url":
		p.pc.URL, err = p.checkUrl(key, value)
		return err
	case key == "landingURL":
		p.pc.LandingURL, err = p.checkUrl(key, value)
		return err
	case key == "isBasedOn":
		return p.decodeArrString(key, []string{value})
	case key == "softwareVersion":
		p.pc.SoftwareVersion = value
	case key == "releaseDate":
		p.pc.ReleaseDate, err = p.checkDate(key, value)
		return err
	case key == "logo":
		p.pc.Logo, err = p.checkImage(key, value)
		return err
	case key == "monochromeLogo":
		p.pc.MonochromeLogo, err = p.checkImage(key, value)
		return err
	case key == "platforms":
		return p.decodeArrString(key, []string{value})
	case key == "api":
		p.pc.API, err = p.checkUrl(key, value)
		return err
	case key == "tags":
		return p.decodeArrString(key, []string{value})
	case key == "roadmap":
		p.pc.Roadmap, err = p.checkUrl(key, value)
		return err
	case key == "developmentStatus":
		for _, v := range []string{"concept", "development", "beta", "stable", "obsolete"} {
			if v == value {
				p.pc.DevelopmentStatus = value
				return nil
			}
		}
		return newErrorInvalidValue(key, "invalid value: %s", value)
	case key == "softwareType/isRelatedTo":
		return p.decodeArrString(key, []string{value})
	case key == "softwareType/type":
		for _, v := range []string{"standalone", "addon", "library", "configurationFiles"} {
			if v == value {
				p.pc.SoftwareType.Type = value
				return nil
			}
		}
		return newErrorInvalidValue(key, "invalid value: %s", value)
	case regexp.MustCompile(`^description/[a-z]{3}`).MatchString(key):
		if p.pc.Description == nil {
			p.pc.Description = make(map[string]Desc)
		}
		k := strings.Split(key, "/")[1]
		attr := strings.Split(key, "/")[2]
		var desc = p.pc.Description[k]
		if attr == "localisedName" {
			desc.LocalisedName = value
			return
		}
		if attr == "longDescription" {
			if len(value) < 500 || len(value) > 10000 {
				return newErrorInvalidValue(key, "\"%s\" has an invalid number of characters: %d.  (min 500 chars, max 10.000 chars)", key, len(value))
			}
			desc.LongDescription = value
			return
		}
		if attr == "documentation" {
			desc.Documentation, err = p.checkUrl(key, value)
			return err
		}
		if attr == "shortDescription" {
			if len(value) > 150 {
				return newErrorInvalidValue(key, "\"%s\" has an invalid number of characters: %d.  (max 150 chars)", key, len(value))
			}
			desc.ShortDescription = value
			return err
		}
		return ErrorInvalidKey{key + " : String"}
	case key == "legal/authorsFile":
		p.pc.Legal.AuthorsFile, err = p.checkFile(key, value)
		return err
	case key == "legal/license":
		p.pc.Legal.License = value
		return p.checkSpdx(key, value)
	case key == "legal/mainCopyrightOwner":
		p.pc.Legal.MainCopyrightOwner = value
	case key == "legal/repoOwner":
		p.pc.Legal.RepoOwner = value
	case key == "maintenance/type":
		for _, v := range []string{"internal", "contract", "community", "none"} {
			if v == value {
				p.pc.Maintenance.Type = value
				return nil
			}
		}
		return newErrorInvalidValue(key, "invalid value: %s", value)
	case key == "ita/riuso/codiceIPA":
		// TODO: check valid codiceIPA
		p.pc.Ita.Riuso.CodiceIPA = value
	default:
		return ErrorInvalidKey{key + " : String"}
	}
	return
}

func (p *parser) decodeArrString(key string, value []string) error {
	switch {
	case key == "isBasedOn":
		for _, v := range value {
			p.pc.IsBasedOn = append(p.pc.IsBasedOn, v)
		}
	case key == "platforms":
		for _, v := range value {
			p.pc.Platforms = append(p.pc.Platforms, v)
		}
	case key == "tags":
		for _, v := range value {
			v, err := p.checkTag(key, v)
			if err != nil {
				return err
			}
			p.pc.Tags = append(p.pc.Tags, v)
		}
	case regexp.MustCompile(`^freeTags/`).MatchString(key):
		if p.pc.FreeTags == nil {
			p.pc.FreeTags = make(map[string][]string)
		}
		k := strings.Split(key, "/")[1]
		for _, v := range value {
			p.pc.FreeTags[k] = append(p.pc.FreeTags[k], v)
		}
	case key == "usedBy":
		for _, v := range value {
			p.pc.UsedBy = append(p.pc.UsedBy, v)
		}
	case key == "softwareType/isRelatedTo":
		for _, v := range value {
			u, err := p.checkUrl(key, v)
			if err != nil {
				return err
			}
			p.pc.SoftwareType.IsRelatedTo = append(p.pc.SoftwareType.IsRelatedTo, u)
		}
	case key == "intendedAudience/countries":
		for _, v := range value {
			p.pc.IntendedAudience.Countries = append(p.pc.IntendedAudience.Countries, v)
		}
	case key == "intendedAudience/unsupportedCountries":
		for _, v := range value {
			p.pc.IntendedAudience.UnsupportedCountries = append(p.pc.IntendedAudience.UnsupportedCountries, v)
		}
	case key == "intendedAudience/onlyFor":
		for _, v := range value {
			v, err := p.checkPaTypes(key, v)
			if err != nil {
				return err
			}
			p.pc.IntendedAudience.OnlyFor = append(p.pc.IntendedAudience.OnlyFor, v)
		}
	case regexp.MustCompile(`^description/[a-z]{3}`).MatchString(key):
		if p.pc.Description == nil {
			p.pc.Description = make(map[string]Desc)
		}
		k := strings.Split(key, "/")[1]
		attr := strings.Split(key, "/")[2]
		var desc = p.pc.Description[k]
		if attr == "awards" {
			for _, v := range value {
				desc.Awards = append(desc.Awards, v)
			}
		}
		if attr == "featureList" {
			for _, v := range value {
				if len(v) > 100 {
					return newErrorInvalidValue(key, " %s is too long.  (max 100 chars)", key)

				}
				desc.FeatureList = append(desc.FeatureList, v)
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
				u, err := p.checkUrl(key, v)
				if err != nil {
					return err
				}
				desc.Videos = append(desc.Videos, u)
			}
		}
		return nil
	case key == "localisation/availableLanguages":
		for _, v := range value {
			p.pc.Localisation.AvailableLanguages = append(p.pc.Localisation.AvailableLanguages, v)
		}

	default:
		return ErrorInvalidKey{key + " : Array of Strings"}

	}
	return nil
}

func (p *parser) decodeArrObj(key string, value map[interface{}]interface{}) error {
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
					u, err := p.checkUrl(key, val.(string))
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
			p.pc.Maintenance.Contractors = append(p.pc.Maintenance.Contractors, contractor)
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

			p.pc.Maintenance.Contacts = append(p.pc.Maintenance.Contacts, contact)
		}
	case "dependencies/open":
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

			p.pc.Dependencies.Open = append(p.pc.Dependencies.Open, dep)
		}
	case "dependencies/proprietary":
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

			p.pc.Dependencies.Proprietary = append(p.pc.Dependencies.Proprietary, dep)
		}
	case "dependencies/hardware":
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

			p.pc.Dependencies.Hardware = append(p.pc.Dependencies.Hardware, dep)
		}
	default:
		return ErrorInvalidKey{key + " : Array of Objects"}
	}
	return nil
}

func (p *parser) finalize() (es ErrorParseMulti) {
	// "maintenance/contractors" presence is mandatory (if maintainance/type is contract).
	if p.pc.Maintenance.Type == "contract" && len(p.pc.Maintenance.Contractors) == 0 {
		es = append(es, newErrorInvalidValue("maintenance/contractors", "not found, mandatory for a \"contract\" maintenance"))
	}

	// maintenance/contacts/name is always mandatory
	if len(p.pc.Maintenance.Contacts) > 0 {
		for _, c := range p.pc.Maintenance.Contacts {
			if c.Name == "" {
				es = append(es, newErrorInvalidValue("maintenance/contacts/name", "not found. It's mandatory."))
			}
		}
	}
	// maintenance/contractors/name is always mandatory
	if len(p.pc.Maintenance.Contractors) > 0 {
		for _, c := range p.pc.Maintenance.Contractors {
			if c.Name == "" {
				es = append(es, newErrorInvalidValue("maintenance/contractors/name", "not found. It's mandatory."))
			}
		}
	}
	// maintenance/contractors/until is always mandatory
	if len(p.pc.Maintenance.Contractors) > 0 {
		for _, c := range p.pc.Maintenance.Contractors {
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
