package publiccode

import (
	"net/url"
	"cloud.google.com/go/civil"
	"github.com/go-playground/validator/v10"
)

// list of keys that were renamed in the publiccode.yml spec
var renamedKeys = map[string]string{
	"publiccode-yaml-version":    "publiccodeYmlVersion",
	"it/conforme/accessibile":    "it/conforme/lineeGuidaDesign",
	"it/conforme/interoperabile": "it/conforme/modelloInteroperabilita",
	"it/conforme/sicuro":         "it/conforme/misureMinimeSicurezza",
	"it/conforme/privacy":        "it/conforme/gdpr",
	"it/spid":                    "it/piattaforme/spid",
	"it/pagopa":                  "it/piattaforme/pagopa",
	"it/cie":                     "it/piattaforme/cie",
	"it/anpr":                    "it/piattaforme/anpr",
}

// list of keys that were removed from the publiccode.yml spec
var removedKeys = []string{
	"tags",                     // deprecated in it:0.2
	"intendedAudience/onlyFor", // deprecated in it:0.2
	"it/designKit/seo",         // deprecated in it:0.2
	"it/designKit/ui",          // deprecated in it:0.2
	"it/designKit/web",         // deprecated in it:0.2
	"it/designKit/content",     // deprecated in it:0.2
	"it/ecosistemi",            // deprecated in it:0.2
}

// Version of the latest PublicCode specs.
// Source https://github.com/publiccodenet/publiccode.yml
const Version = "0.2"

// SupportedVersions lists the publiccode.yml versions this parser supports.
var SupportedVersions = []string{"0.1", "0.2"}

// PublicCode is a publiccode.yml file definition.
// Reference: https://github.com/publiccodenet/publiccode.yml
type PublicCode struct {
	PubliccodeYamlVersion string `yaml:"publiccodeYmlVersion" validate:"required"`

	Name             string   `yaml:"name" validate:"required"`
	ApplicationSuite string   `yaml:"applicationSuite,omitempty"`
	URL              *url.URL `yaml:"url" validate="required"`
	LandingURL       *url.URL `yaml:"landingURL,omitempty"`

	IsBasedOn         []string   `yaml:"isBasedOn,omitempty"`
	SoftwareVersion   string     `yaml:"softwareVersion,omitempty"`
	ReleaseDate       civil.Date `yaml:"releaseDate" validate="required"`
	Logo              string     `yaml:"logo,omitempty"`
	MonochromeLogo    string     `yaml:"monochromeLogo,omitempty"`

	InputTypes  []string `yaml:"inputTypes,omitempty"`
	OutputTypes []string `yaml:"outputTypes,omitempty"`

	Platforms []string `yaml:"platforms" validate="gt=0"`

	Categories []string `yaml:"categories" validate="gt=0"`

	UsedBy []string `yaml:"usedBy,omitempty"`

	Roadmap *url.URL `yaml:"roadmap,omitempty"`

	DevelopmentStatus string `yaml:"developmentStatus" validate="required"`

	SoftwareType string `yaml:"softwareType" validate="required"`

	IntendedAudience struct {
		Scope                []string `yaml:"scope,omitempty"`
		Countries            []string `yaml:"countries,omitempty" validate:"dive,iso3166_1_alpha2"`
		UnsupportedCountries []string `yaml:"unsupportedCountries,omitempty" validate:"dive,iso3166_1_alpha2"`
	} `yaml:"intendedAudience"`

	Description map[string]Desc `yaml:"description" validate:"gt=0,dive,keys,iso3166_1_alpha3,end_keys"`

	Legal struct {
		License            string `yaml:"license" validate="required"`
		MainCopyrightOwner string `yaml:"mainCopyrightOwner,omitempty"`
		RepoOwner          string `yaml:"repoOwner,omitempty"`
		AuthorsFile        string `yaml:"authorsFile,omitempty"`
	} `yaml:"legal"`

	Maintenance struct {
		Type        string       `yaml:"type" validate="required"`
		Contractors []Contractor `yaml:"contractors,omitempty"`
		Contacts    []Contact    `yaml:"contacts,omitempty"`
	} `yaml:"maintenance"`

	Localisation struct {
		LocalisationReady  bool     `yaml:"localisationReady" validate="required"`
		AvailableLanguages []string `yaml:"availableLanguages" validate:"required,dive,iso3166_1_alpha3"`
	} `yaml:"localisation"`

	DependsOn struct {
		Open        []Dependency `yaml:"open,omitempty"`
		Proprietary []Dependency `yaml:"proprietary,omitempty"`
		Hardware    []Dependency `yaml:"hardware,omitempty"`
	} `yaml:"dependsOn,omitempty"`

	It ExtensionIT `yaml:"it"`
}

// Desc is a general description of the software.
// Reference: https://github.com/publiccodenet/publiccode.yml/blob/develop/schema.md#section-description
type Desc struct {
	LocalisedName          string     `yaml:"localisedName,omitempty"`
	GenericName            string     `yaml:"genericName" validate:"required"`
	ShortDescription       string     `yaml:"shortDescription" validate:"required"`
	LongDescription        string     `yaml:"longDescription,omitempty" validate:"required"`
	Documentation          *url.URL   `yaml:"documentation,omitempty"`
	APIDocumentation       *url.URL   `yaml:"apiDocumentation,omitempty"`
	Features               []string   `yaml:"features,omitempty" validate:"gt=0"`
	Screenshots            []string   `yaml:"screenshots,omitempty"`
	Videos                 []*url.URL `yaml:"videos,omitempty"`
	Awards                 []string   `yaml:"awards,omitempty"`
}

// Contractor is an entity or entities, if any, that are currently contracted for maintaining the software.
// Reference: https://github.com/publiccodenet/publiccode.yml/blob/develop/schema.md#contractor
type Contractor struct {
	Name          string     `yaml:"name" validate:"required"`
	Email         string     `yaml:"email,omitempty" validate:"email"`
	Website       *url.URL   `yaml:"website,omitempty"`
	Until         civil.Date `yaml:"until" validate:"required"`
}

// Contact is a contact info maintaining the software.
// Reference: https://github.com/publiccodenet/publiccode.yml/blob/develop/schema.md#contact
type Contact struct {
	Name        string `yaml:"name" validate:"required"`
	Email       string `yaml:"email,omitempty" validate:"email"`
	Affiliation string `yaml:"affiliation,omitempty"`
	Phone       string `yaml:"phone,omitempty"`
}

// Dependency describe system-level dependencies required to install and use this software.
// Reference: https://github.com/publiccodenet/publiccode.yml/blob/develop/schema.md#section-dependencies
type Dependency struct {
	Name       string `yaml:"name" validate:"required"`
	VersionMin string `yaml:"versionMin"`
	VersionMax string `yaml:"versionMax"`
	Optional   bool   `yaml:"optional"`
	Version    string `yaml:"version"`
}
