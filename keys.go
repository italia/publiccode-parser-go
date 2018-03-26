package publiccode

var mandatoryKeys = []string{
	"version", "url",
	"legal/license", "legal/repo-owner",
	"maintenance/type", "maintenance/maintainer",
}

func (p *parser) decodeString(key string, value string) (err error) {
	switch key {
	case "version":
		p.pc.Version = value
		if p.pc.Version != Version {
			return newErrorInvalidValue(key, "version %s not supported", p.pc.Version)
		}
	case "url":
		p.pc.Url, err = p.checkUrl(key, value)
	case "upstream-url":
		return p.decodeArrString(key, []string{value})
	case "legal/license":
		p.pc.Legal.License = value
		return p.checkSpdx(key, value)
	case "legal/main-copyright-owner":
		p.pc.Legal.MainCopyrightOwner = value
	case "legal/authors-file":
		p.pc.Legal.AuthorsFile = value
		return p.checkFile(key, value)
	case "legal/repo-owner":
		p.pc.Legal.RepoOwner = value
	case "maintenance/type":
		for _, v := range []string{"community", "commercial", "none"} {
			if v == value {
				p.pc.Maintenance.Type = value
				return nil
			}
		}
		return newErrorInvalidValue(key, "invalid value: %s", value)
	case "maintenance/until":
		p.pc.Maintenance.Until, err = p.checkDate(key, value)
	case "maintenance/maintainer":
		return p.decodeArrString(key, []string{value})
	default:
		return ErrorInvalidKey{key}
	}
	return
}

func (p *parser) decodeArrString(key string, value []string) error {
	switch key {
	case "upstream-url":
		for _, v := range value {
			if u, err := p.checkUrl(key, v); err != nil {
				return err
			} else {
				p.pc.UpstreamUrl = append(p.pc.UpstreamUrl, u)
			}
		}
	case "maintenance/maintainer":
		p.pc.Maintenance.Maintainer = value
	default:
		return ErrorInvalidKey{key}
	}
	return nil
}

func (p *parser) finalize() (es ErrorParseMulti) {
	if p.pc.Maintenance.Type == "commercial" && p.pc.Maintenance.Until.IsZero() {
		es = append(es, newErrorInvalidValue("maintenance/until", "not found, mandatory for a commercial maintenance"))
	}
	return
}
