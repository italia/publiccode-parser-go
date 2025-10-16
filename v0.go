package publiccode

import (
	urlutil "github.com/italia/publiccode-parser-go/v4/internal"
	"gopkg.in/yaml.v3"
)

// PublicCodeV0 defines how a publiccode.yml v0.x is structured
type PublicCodeV0 struct {
	PubliccodeYamlVersion string `json:"publiccodeYmlVersion" validate:"required,oneof=0 0.2 0.2.0 0.2.1 0.2.2 0.3 0.3.0 0.4 0.4.0 0.5 0.5.0" yaml:"publiccodeYmlVersion"`

	Name             string `json:"name"                       validate:"required"               yaml:"name"`
	ApplicationSuite string `json:"applicationSuite,omitempty" yaml:"applicationSuite,omitempty"`
	URL              *URL   `json:"url"                        validate:"required,url_url"       yaml:"url"`
	LandingURL       *URL   `json:"landingURL,omitempty"       validate:"omitnil,url_http_url"   yaml:"landingURL,omitempty"`

	IsBasedOn       UrlOrUrlArray `json:"isBasedOn,omitempty"       validate:"omitempty,dive,url_url" yaml:"isBasedOn,omitempty"`
	SoftwareVersion string        `json:"softwareVersion,omitempty" yaml:"softwareVersion,omitempty"`
	ReleaseDate     *string       `json:"releaseDate"               validate:"omitnil,date"           yaml:"releaseDate"`
	Logo            string        `json:"logo,omitempty"            yaml:"logo,omitempty"`
	MonochromeLogo  string        `json:"monochromeLogo,omitempty"  yaml:"monochromeLogo,omitempty"`

	Organisation *struct {
		Name string  `json:"name"          validate:"required"                   yaml:"name"`
		URI  *string `json:"uri,omitempty" validate:"omitempty,organisation_uri" yaml:"uri,omitempty"`
	} `json:"organisation,omitempty" yaml:"organisation,omitempty"`

	InputTypes  []string `json:"inputTypes,omitempty"  validate:"omitempty,dive,is_mime_type" yaml:"inputTypes,omitempty"`
	OutputTypes []string `json:"outputTypes,omitempty" validate:"omitempty,dive,is_mime_type" yaml:"outputTypes,omitempty"`

	Platforms []string `json:"platforms" validate:"gt=0" yaml:"platforms"`

	Categories *[]string `json:"categories,omitempty" validate:"omitempty,dive,is_category_v0" yaml:"categories,omitempty"`

	UsedBy *[]string `json:"usedBy,omitempty" yaml:"usedBy,omitempty"`

	FundedBy *[]struct {
		Name string  `json:"name"          validate:"required"                   yaml:"name"`
		URI  *string `json:"uri,omitempty" validate:"omitempty,organisation_uri" yaml:"uri,omitempty"`
	} `json:"fundedBy,omitempty" validate:"omitempty,dive" yaml:"fundedBy,omitempty"`

	Roadmap *URL `json:"roadmap,omitempty" validate:"omitnil,url_http_url" yaml:"roadmap,omitempty"`

	DevelopmentStatus string `json:"developmentStatus" validate:"required,oneof=concept development beta stable obsolete" yaml:"developmentStatus"`

	SoftwareType string `json:"softwareType" validate:"required,oneof=standalone/mobile standalone/iot standalone/desktop standalone/web standalone/backend standalone/other addon library configurationFiles" yaml:"softwareType"`

	IntendedAudience *struct {
		Scope                *[]string `json:"scope,omitempty"                validate:"omitempty,dive,is_scope_v0"                     yaml:"scope,omitempty"`
		Countries            *[]string `json:"countries,omitempty"            validate:"omitempty,dive,iso3166_1_alpha2_lower_or_upper" yaml:"countries,omitempty"`
		UnsupportedCountries *[]string `json:"unsupportedCountries,omitempty" validate:"omitempty,dive,iso3166_1_alpha2_lower_or_upper" yaml:"unsupportedCountries,omitempty"`
	} `json:"intendedAudience,omitempty" yaml:"intendedAudience,omitempty"`

	Description map[string]DescV0 `json:"description" validate:"gt=0,bcp47_keys,dive" yaml:"description"`

	Legal struct {
		License            string  `json:"license"                      validate:"required,is_spdx_expression" yaml:"license"`
		MainCopyrightOwner *string `json:"mainCopyrightOwner,omitempty" yaml:"mainCopyrightOwner,omitempty"`
		RepoOwner          *string `json:"repoOwner,omitempty"          yaml:"repoOwner,omitempty"`
		AuthorsFile        *string `json:"authorsFile,omitempty"        yaml:"authorsFile,omitempty"`
	} `yaml:"legal" json:"legal"`

	Maintenance struct {
		Type        string         `json:"type"                  validate:"required,oneof=internal contract community none"              yaml:"type"`
		Contractors []ContractorV0 `json:"contractors,omitempty" validate:"required_if=Type contract,excluded_unless=Type contract,dive" yaml:"contractors,omitempty"`
		Contacts    []ContactV0    `json:"contacts,omitempty"    validate:"required_if=Type community,required_if=Type internal,dive"    yaml:"contacts,omitempty"`
	} `yaml:"maintenance" json:"maintenance"`

	Localisation struct {
		LocalisationReady  *bool    `json:"localisationReady"  validate:"required"                              yaml:"localisationReady"`
		AvailableLanguages []string `json:"availableLanguages" validate:"required,gt=0,dive,bcp47_language_tag" yaml:"availableLanguages"`
	} `yaml:"localisation" json:"localisation"`

	DependsOn *struct {
		Open        *[]DependencyV0 `json:"open,omitempty"        validate:"omitempty,dive" yaml:"open,omitempty"`
		Proprietary *[]DependencyV0 `json:"proprietary,omitempty" validate:"omitempty,dive" yaml:"proprietary,omitempty"`
		Hardware    *[]DependencyV0 `json:"hardware,omitempty"    validate:"omitempty,dive" yaml:"hardware,omitempty"`
	} `json:"dependsOn,omitempty" yaml:"dependsOn,omitempty"`

	IT *ITSectionV0 `json:"IT,omitempty" yaml:"IT,omitempty"`

	// Don't use, this is provided for backwards compatibility.
	// IT always has the same data.
	It *ITSectionV0 `json:"it,omitempty" yaml:"it,omitempty"`
}

// DescV0 is a general description of the software.
type DescV0 struct {
	LocalisedName    *string   `json:"localisedName,omitempty"    yaml:"localisedName,omitempty"`
	GenericName      string    `json:"genericName"                validate:"umax=35"                      yaml:"genericName"`
	ShortDescription string    `json:"shortDescription"           validate:"required,umax=150"            yaml:"shortDescription"`
	LongDescription  string    `json:"longDescription,omitempty"  validate:"required,umin=150,umax=10000" yaml:"longDescription,omitempty"`
	Documentation    *URL      `json:"documentation,omitempty"    validate:"omitnil,url_http_url"         yaml:"documentation,omitempty"`
	APIDocumentation *URL      `json:"apiDocumentation,omitempty" validate:"omitnil,url_http_url"         yaml:"apiDocumentation,omitempty"`
	Features         *[]string `json:"features,omitempty"         validate:"gt=0,dive"                    yaml:"features,omitempty"`
	Screenshots      []string  `json:"screenshots,omitempty"      yaml:"screenshots,omitempty"`
	Videos           []*URL    `json:"videos,omitempty"           validate:"dive,omitnil,url_http_url"    yaml:"videos,omitempty"`
	Awards           []string  `json:"awards,omitempty"           yaml:"awards,omitempty"`
}

// ContractorV0 is an entity or entities, if any, that are currently contracted for maintaining the software.
type ContractorV0 struct {
	Name    string  `json:"name"              validate:"required"             yaml:"name"`
	Email   *string `json:"email,omitempty"   validate:"omitempty,email"      yaml:"email,omitempty"`
	Website *URL    `json:"website,omitempty" validate:"omitnil,url_http_url" yaml:"website,omitempty"`
	Until   string  `json:"until"             validate:"required,date"        yaml:"until"`
}

// ContactV0 is a contact info maintaining the software.
type ContactV0 struct {
	Name        string  `json:"name"                  validate:"required"          yaml:"name"`
	Email       *string `json:"email,omitempty"       validate:"omitempty,email"   yaml:"email,omitempty"`
	Affiliation *string `json:"affiliation,omitempty" yaml:"affiliation,omitempty"`
	Phone       *string `json:"phone,omitempty"       validate:"omitempty"         yaml:"phone,omitempty"`
}

// DependencyV0 describes system-level dependencies required to install and use this software.
type DependencyV0 struct {
	Name       string  `json:"name"                 validate:"required,gt=0"    yaml:"name"`
	VersionMin *string `json:"versionMin,omitempty" yaml:"versionMin,omitempty"`
	VersionMax *string `json:"versionMax,omitempty" yaml:"versionMax,omitempty"`
	Optional   *bool   `json:"optional,omitempty"   yaml:"optional,omitempty"`
	Version    *string `json:"version,omitempty"    yaml:"version,omitempty"`
}

// Country-specific sections
//
// While the standard is structured to be meaningful on an international level,
// there are additional information that can be added that makes sense in specific
// countries, such as declaring compliance with local laws or regulations.

type ITSectionV0 struct {
	CountryExtensionVersion *string `json:"countryExtensionVersion" validate:"omitnil,oneof=0.2 1.0" yaml:"countryExtensionVersion"`

	Conforme *struct {
		LineeGuidaDesign        *bool `json:"lineeGuidaDesign,omitempty"        yaml:"lineeGuidaDesign,omitempty"`
		ModelloInteroperabilita *bool `json:"modelloInteroperabilita,omitempty" yaml:"modelloInteroperabilita,omitempty"`
		MisureMinimeSicurezza   *bool `json:"misureMinimeSicurezza,omitempty"   yaml:"misureMinimeSicurezza,omitempty"`
		GDPR                    *bool `json:"gdpr,omitempty"                    yaml:"gdpr,omitempty"`
	} `json:"conforme,omitempty" yaml:"conforme,omitempty"`

	Riuso struct {
		CodiceIPA string `json:"codiceIPA,omitempty" validate:"omitempty,is_italian_ipa_code" yaml:"codiceIPA,omitempty"`
	} `yaml:"riuso,omitempty" json:"riuso,omitempty"`

	Piattaforme struct {
		SPID   bool `json:"spid"   yaml:"spid"`
		PagoPa bool `json:"pagopa" yaml:"pagopa"`
		CIE    bool `json:"cie"    yaml:"cie"`
		ANPR   bool `json:"anpr"   yaml:"anpr"`
		Io     bool `json:"io"     yaml:"io"`
	} `yaml:"piattaforme" json:"piattaforme"`
}

func (p PublicCodeV0) Version() uint {
	return 0
}

func (p PublicCodeV0) ToYAML() ([]byte, error) {
	return yaml.Marshal(p)
}

func (p PublicCodeV0) Url() *URL {
	if p.URL == nil {
		return nil
	}

	if ok, _ := urlutil.IsValidURL(p.URL.String()); ok {
		return p.URL
	}

	return nil
}
