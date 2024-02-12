package publiccode

import (
	"gopkg.in/yaml.v3"

	urlutil "github.com/italia/publiccode-parser-go/v4/internal"
)

// PublicCode is a publiccode.yml file definition.
type PublicCodeV0 struct {
	PubliccodeYamlVersion string `yaml:"publiccodeYmlVersion" validate:"required,oneof=0.2 0.2.0 0.2.1 0.2.2 0.3 0.3.0"`

	Name             string   `yaml:"name" validate:"required"`
	ApplicationSuite string   `yaml:"applicationSuite,omitempty"`
	URL              *URL     `yaml:"url" validate:"required,url_url"`
	LandingURL       *URL     `yaml:"landingURL,omitempty" validate:"omitnil,url_http_url"`

	IsBasedOn         UrlOrUrlArray `yaml:"isBasedOn,omitempty"`
	SoftwareVersion   string        `yaml:"softwareVersion,omitempty"`
	ReleaseDate       string        `yaml:"releaseDate" validate:"required,date"`
	Logo              string        `yaml:"logo,omitempty"`
	MonochromeLogo    string        `yaml:"monochromeLogo,omitempty"`

	InputTypes  []string `yaml:"inputTypes,omitempty"`
	OutputTypes []string `yaml:"outputTypes,omitempty"`

	Platforms []string `yaml:"platforms" validate:"gt=0"`

	Categories []string `yaml:"categories" validate:"required,gt=0,dive,is_category_v0_2"`

	UsedBy *[]string `yaml:"usedBy,omitempty"`

	Roadmap *URL `yaml:"roadmap,omitempty" validate:"omitnil,url_http_url"`

	DevelopmentStatus string `yaml:"developmentStatus" validate:"required,oneof=concept development beta stable obsolete"`

	SoftwareType string `yaml:"softwareType" validate:"required,oneof=standalone/mobile standalone/iot standalone/desktop standalone/web standalone/backend standalone/other addon library configurationFiles"`

	IntendedAudience *struct {
		Scope                *[]string `yaml:"scope,omitempty" validate:"omitempty,dive,is_scope_v0_2"`
		Countries            *[]string `yaml:"countries,omitempty" validate:"omitempty,dive,iso3166_1_alpha2_lowercase"`
		UnsupportedCountries *[]string `yaml:"unsupportedCountries,omitempty" validate:"omitempty,dive,iso3166_1_alpha2_lowercase"`
	} `yaml:"intendedAudience,omitempty"`

	Description map[string]DescV0 `yaml:"description" validate:"gt=0,dive,keys,bcp47_language_tag,endkeys,required"`

	Legal struct {
		License            string  `yaml:"license" validate:"required"`
		MainCopyrightOwner *string `yaml:"mainCopyrightOwner,omitempty"`
		RepoOwner          *string `yaml:"repoOwner,omitempty"`
		AuthorsFile        *string `yaml:"authorsFile,omitempty"`
	} `yaml:"legal" validate:"required"`

	Maintenance struct {
		Type        string         `yaml:"type" validate:"required,oneof=internal contract community none"`
		Contractors []ContractorV0 `yaml:"contractors,omitempty" validate:"required_if=Type contract,dive"`
		Contacts    []ContactV0    `yaml:"contacts,omitempty" validate:"required_if=Type community,required_if=Type internal,dive"`
	} `yaml:"maintenance"`

	Localisation struct {
		LocalisationReady  *bool    `yaml:"localisationReady" validate:"required"`
		AvailableLanguages []string `yaml:"availableLanguages" validate:"required,gt=0,dive,bcp47_language_tag"`
	} `yaml:"localisation" validate:"required"`

	DependsOn *struct {
		Open        *[]DependencyV0 `yaml:"open,omitempty" validate:"omitempty,dive"`
		Proprietary *[]DependencyV0 `yaml:"proprietary,omitempty" validate:"omitempty,dive"`
		Hardware    *[]DependencyV0 `yaml:"hardware,omitempty" validate:"omitempty,dive"`
	} `yaml:"dependsOn,omitempty"`

	It ITSectionV0 `yaml:"it"`
}

// Desc is a general description of the software.
type DescV0 struct {
	LocalisedName          *string   `yaml:"localisedName,omitempty"`
	GenericName            string    `yaml:"genericName" validate:"umax=35"`
	ShortDescription       string    `yaml:"shortDescription" validate:"required,umax=150"`
	LongDescription        string    `yaml:"longDescription,omitempty" validate:"required,umin=150,umax=10000"`
	Documentation          *URL      `yaml:"documentation,omitempty" validate:"omitnil,url_http_url"`
	APIDocumentation       *URL      `yaml:"apiDocumentation,omitempty" validate:"omitnil,url_http_url"`
	Features               *[]string `yaml:"features,omitempty" validate:"gt=0,dive"`
	Screenshots            []string  `yaml:"screenshots,omitempty"`
	Videos                 []*URL    `yaml:"videos,omitempty" validate:"dive,omitnil,url_http_url"`
	Awards                 []string  `yaml:"awards,omitempty"`
}

// Contractor is an entity or entities, if any, that are currently contracted for maintaining the software.
type ContractorV0 struct {
	Name          string  `yaml:"name" validate:"required"`
	Email         *string `yaml:"email,omitempty" validate:"omitempty,email"`
	Website       *URL    `yaml:"website,omitempty" validate:"omitnil,url_http_url"`
	Until         string  `yaml:"until" validate:"required,date"`
}

// Contact is a contact info maintaining the software.
type ContactV0 struct {
	Name        string  `yaml:"name" validate:"required"`
	Email       *string `yaml:"email,omitempty" validate:"omitempty,email"`
	Affiliation *string `yaml:"affiliation,omitempty"`
	Phone       *string `yaml:"phone,omitempty" validate:"omitempty"`
}

// Dependency describe system-level dependencies required to install and use this software.
type DependencyV0 struct {
	Name       string  `yaml:"name" validate:"required,gt=0"`
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
//
// All country-specific sections are contained in a section named with
// the two-letter lowercase ISO 3166-1 alpha-2 country code.

type ITSectionV0 struct {
	CountryExtensionVersion string `yaml:"countryExtensionVersion"`

	Conforme struct {
		LineeGuidaDesign        bool `yaml:"lineeGuidaDesign,omitempty"`
		ModelloInteroperabilita bool `yaml:"modelloInteroperabilita"`
		MisureMinimeSicurezza   bool `yaml:"misureMinimeSicurezza"`
		GDPR                    bool `yaml:"gdpr"`
	} `yaml:"conforme"`

	Riuso struct {
		CodiceIPA string `yaml:"codiceIPA,omitempty" validate:"omitempty,is_italian_ipa_code"`
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
	return 0;
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
