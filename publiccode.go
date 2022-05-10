package publiccode

// Version of the latest PublicCode specs.
// Source https://github.com/publiccodenet/publiccode.yml
const Version = "0.3"

// SupportedVersions lists the publiccode.yml versions this parser supports.
var SupportedVersions = []string{"0.2", "0.2.0", "0.2.1", "0.2.2", "0.3", "0.3.0"}

// PublicCode is a publiccode.yml file definition.
// Reference: https://github.com/publiccodenet/publiccode.yml
type PublicCode struct {
	PubliccodeYamlVersion string `yaml:"publiccodeYmlVersion" validate:"required,oneof=0.2 0.2.0 0.2.1 0.2.2 0.3 0.3.0"`

	Name             string   `yaml:"name" validate:"required"`
	ApplicationSuite string   `yaml:"applicationSuite,omitempty"`
	URL              *URL 	  `yaml:"url" validate:"required"`
	LandingURL       *URL     `yaml:"landingURL,omitempty"`

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

	Roadmap *URL `yaml:"roadmap,omitempty"`

	DevelopmentStatus string `yaml:"developmentStatus" validate:"required,oneof=concept development beta stable obsolete"`

	SoftwareType string `yaml:"softwareType" validate:"required,oneof=standalone/mobile standalone/iot standalone/desktop standalone/web standalone/backend standalone/other addon library configurationFiles"`

	IntendedAudience *struct {
		Scope                *[]string `yaml:"scope,omitempty" validate:"omitempty,dive,is_scope_v0_2"`
		Countries            *[]string `yaml:"countries,omitempty" validate:"omitempty,dive,iso3166_1_alpha2_lowercase"`
		UnsupportedCountries *[]string `yaml:"unsupportedCountries,omitempty" validate:"omitempty,dive,iso3166_1_alpha2_lowercase"`
	} `yaml:"intendedAudience,omitempty"`

	Description map[string]Desc `yaml:"description" validate:"gt=0,dive,keys,bcp47_language_tag,endkeys,dive"`

	Legal struct {
		License            string  `yaml:"license" validate:"required"`
		MainCopyrightOwner *string `yaml:"mainCopyrightOwner,omitempty"`
		RepoOwner          *string `yaml:"repoOwner,omitempty"`
		AuthorsFile        *string `yaml:"authorsFile,omitempty"`
	} `yaml:"legal" validate:"required"`

	Maintenance struct {
		Type        string        `yaml:"type" validate:"required,oneof=internal contract community none"`
		Contractors []Contractor `yaml:"contractors,omitempty" validate:"required_if=Type contract,dive"`
		Contacts    []Contact    `yaml:"contacts,omitempty" validate:"required_if=Type community,required_if=Type internal,dive"`
	} `yaml:"maintenance"`

	Localisation struct {
		LocalisationReady  *bool    `yaml:"localisationReady" validate:"required"`
		AvailableLanguages []string `yaml:"availableLanguages" validate:"required,gt=0,dive,bcp47_language_tag"`
	} `yaml:"localisation" validate:"required"`

	DependsOn *struct {
		Open        *[]Dependency `yaml:"open,omitempty" validate:"omitempty,dive"`
		Proprietary *[]Dependency `yaml:"proprietary,omitempty" validate:"omitempty,dive"`
		Hardware    *[]Dependency `yaml:"hardware,omitempty" validate:"omitempty,dive"`
	} `yaml:"dependsOn,omitempty"`

	It ExtensionIT `yaml:"it"`
}

// Desc is a general description of the software.
// Reference: https://github.com/publiccodenet/publiccode.yml/blob/develop/schema.md#section-description
type Desc struct {
	LocalisedName          *string   `yaml:"localisedName,omitempty"`
	GenericName            string    `yaml:"genericName" validate:"umax=35"`
	ShortDescription       string    `yaml:"shortDescription" validate:"required,umax=150"`
	LongDescription        string    `yaml:"longDescription,omitempty" validate:"required,umin=150,umax=10000"`
	Documentation          *URL      `yaml:"documentation,omitempty"`
	APIDocumentation       *URL      `yaml:"apiDocumentation,omitempty"`
	Features               *[]string `yaml:"features,omitempty" validate:"gt=0,dive"`
	Screenshots            []string  `yaml:"screenshots,omitempty"`
	Videos                 []*URL    `yaml:"videos,omitempty"`
	Awards                 []string  `yaml:"awards,omitempty"`
}

// Contractor is an entity or entities, if any, that are currently contracted for maintaining the software.
// Reference: https://github.com/publiccodenet/publiccode.yml/blob/develop/schema.md#contractor
type Contractor struct {
	Name          string  `yaml:"name" validate:"required"`
	Email         *string `yaml:"email,omitempty" validate:"omitempty,email"`
	Website       *URL    `yaml:"website,omitempty"`
	Until         string  `yaml:"until" validate:"required,date"`
}

// Contact is a contact info maintaining the software.
// Reference: https://github.com/publiccodenet/publiccode.yml/blob/develop/schema.md#contact
type Contact struct {
	Name        string  `yaml:"name" validate:"required"`
	Email       *string `yaml:"email,omitempty" validate:"omitempty,email"`
	Affiliation *string `yaml:"affiliation,omitempty"`
	Phone       *string `yaml:"phone,omitempty" validate:"omitempty"`
}

// Dependency describe system-level dependencies required to install and use this software.
// Reference: https://github.com/publiccodenet/publiccode.yml/blob/develop/schema.md#section-dependencies
type Dependency struct {
	Name       string  `yaml:"name" validate:"required,gt=0"`
	VersionMin *string `yaml:"versionMin,omitempty"`
	VersionMax *string `yaml:"versionMax,omitempty"`
	Optional   *bool   `yaml:"optional,omitempty"`
	Version    *string `yaml:"version,omitempty"`
}
