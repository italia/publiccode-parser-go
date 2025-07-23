package publiccode

import (
	"github.com/goccy/go-yaml"
	urlutil "github.com/italia/publiccode-parser-go/v4/internal"
)

// PublicCodeV0 defines how a publiccode.yml v0.x is structured
type PublicCodeV0 struct {
	PubliccodeYamlVersion string `validate:"required,oneof=0.2 0.2.0 0.2.1 0.2.2 0.3 0.3.0 0.4 0.4.0" yaml:"publiccodeYmlVersion"`

	Name             string `validate:"required"               yaml:"name"`
	ApplicationSuite string `yaml:"applicationSuite,omitempty"`
	URL              *URL   `validate:"required,url_url"       yaml:"url"`
	LandingURL       *URL   `validate:"omitnil,url_http_url"   yaml:"landingURL,omitempty"`

	IsBasedOn       UrlOrUrlArray `yaml:"isBasedOn,omitempty"`
	SoftwareVersion string        `yaml:"softwareVersion,omitempty"`
	ReleaseDate     *string       `validate:"omitnil,date"          yaml:"releaseDate"`
	Logo            string        `yaml:"logo,omitempty"`
	MonochromeLogo  string        `yaml:"monochromeLogo,omitempty"`

	InputTypes  []string `yaml:"inputTypes,omitempty"`
	OutputTypes []string `yaml:"outputTypes,omitempty"`

	Platforms []string `validate:"gt=0" yaml:"platforms"`

	Categories []string `validate:"required,gt=0,dive,is_category_v0" yaml:"categories"`

	UsedBy *[]string `yaml:"usedBy,omitempty"`

	Roadmap *URL `validate:"omitnil,url_http_url" yaml:"roadmap,omitempty"`

	DevelopmentStatus string `validate:"required,oneof=concept development beta stable obsolete" yaml:"developmentStatus"`

	SoftwareType string `validate:"required,oneof=standalone/mobile standalone/iot standalone/desktop standalone/web standalone/backend standalone/other addon library configurationFiles" yaml:"softwareType"`

	IntendedAudience *struct {
		Scope                *[]string `validate:"omitempty,dive,is_scope_v0"                yaml:"scope,omitempty"`
		Countries            *[]string `validate:"omitempty,dive,iso3166_1_alpha2_lowercase" yaml:"countries,omitempty"`
		UnsupportedCountries *[]string `validate:"omitempty,dive,iso3166_1_alpha2_lowercase" yaml:"unsupportedCountries,omitempty"`
	} `yaml:"intendedAudience,omitempty"`

	Description map[string]DescV0 `validate:"gt=0,bcp47_keys,dive" yaml:"description"`

	Legal struct {
		License            string  `validate:"required,is_spdx_expression" yaml:"license"`
		MainCopyrightOwner *string `yaml:"mainCopyrightOwner,omitempty"`
		RepoOwner          *string `yaml:"repoOwner,omitempty"`
		AuthorsFile        *string `yaml:"authorsFile,omitempty"`
	} `yaml:"legal"`

	Maintenance struct {
		Type        string         `validate:"required,oneof=internal contract community none"           yaml:"type"`
		Contractors []ContractorV0 `validate:"required_if=Type contract,dive"                            yaml:"contractors,omitempty"`
		Contacts    []ContactV0    `validate:"required_if=Type community,required_if=Type internal,dive" yaml:"contacts,omitempty"`
	} `yaml:"maintenance"`

	Localisation struct {
		LocalisationReady  *bool    `validate:"required"                              yaml:"localisationReady"`
		AvailableLanguages []string `validate:"required,gt=0,dive,bcp47_language_tag" yaml:"availableLanguages"`
	} `yaml:"localisation"`

	DependsOn *struct {
		Open        *[]DependencyV0 `validate:"omitempty,dive" yaml:"open,omitempty"`
		Proprietary *[]DependencyV0 `validate:"omitempty,dive" yaml:"proprietary,omitempty"`
		Hardware    *[]DependencyV0 `validate:"omitempty,dive" yaml:"hardware,omitempty"`
	} `yaml:"dependsOn,omitempty"`

	It ITSectionV0 `yaml:"it"`
}

// DescV0 is a general description of the software.
type DescV0 struct {
	LocalisedName    *string   `yaml:"localisedName,omitempty"`
	GenericName      string    `validate:"umax=35"                      yaml:"genericName"`
	ShortDescription string    `validate:"required,umax=150"            yaml:"shortDescription"`
	LongDescription  string    `validate:"required,umin=150,umax=10000" yaml:"longDescription,omitempty"`
	Documentation    *URL      `validate:"omitnil,url_http_url"         yaml:"documentation,omitempty"`
	APIDocumentation *URL      `validate:"omitnil,url_http_url"         yaml:"apiDocumentation,omitempty"`
	Features         *[]string `validate:"gt=0,dive"                    yaml:"features,omitempty"`
	Screenshots      []string  `yaml:"screenshots,omitempty"`
	Videos           []*URL    `validate:"dive,omitnil,url_http_url"    yaml:"videos,omitempty"`
	Awards           []string  `yaml:"awards,omitempty"`
}

// ContractorV0 is an entity or entities, if any, that are currently contracted for maintaining the software.
type ContractorV0 struct {
	Name    string  `validate:"required"             yaml:"name"`
	Email   *string `validate:"omitempty,email"      yaml:"email,omitempty"`
	Website *URL    `validate:"omitnil,url_http_url" yaml:"website,omitempty"`
	Until   string  `validate:"required,date"        yaml:"until"`
}

// ContactV0 is a contact info maintaining the software.
type ContactV0 struct {
	Name        string  `validate:"required"          yaml:"name"`
	Email       *string `validate:"omitempty,email"   yaml:"email,omitempty"`
	Affiliation *string `yaml:"affiliation,omitempty"`
	Phone       *string `validate:"omitempty"         yaml:"phone,omitempty"`
}

// DependencyV0 describes system-level dependencies required to install and use this software.
type DependencyV0 struct {
	Name       string  `validate:"required,gt=0"    yaml:"name"`
	VersionMin *string `yaml:"versionMin,omitempty"`
	VersionMax *string `yaml:"versionMax,omitempty"`
	Optional   *bool   `yaml:"optional,omitempty"`
	Version    *string `yaml:"version,omitempty"`
}

// Country-specific sections
//
// While the standard is structured to be meaningful on an international level,
// there are additional information that can be added that makes sense in specific
// countries, such as declaring compliance with local laws or regulations.

type ITSectionV0 struct {
	CountryExtensionVersion *string `validate:"omitnil,oneof=0.2 1.0" yaml:"countryExtensionVersion"`

	Conforme struct {
		LineeGuidaDesign        bool `yaml:"lineeGuidaDesign,omitempty"`
		ModelloInteroperabilita bool `yaml:"modelloInteroperabilita"`
		MisureMinimeSicurezza   bool `yaml:"misureMinimeSicurezza"`
		GDPR                    bool `yaml:"gdpr"`
	} `yaml:"conforme"`

	Riuso struct {
		CodiceIPA string `validate:"omitempty,is_italian_ipa_code" yaml:"codiceIPA,omitempty"`
	} `yaml:"riuso,omitempty"`

	Piattaforme struct {
		SPID   bool `yaml:"spid"`
		PagoPa bool `yaml:"pagopa"`
		CIE    bool `yaml:"cie"`
		ANPR   bool `yaml:"anpr"`
		Io     bool `yaml:"io"`
	} `yaml:"piattaforme"`
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
