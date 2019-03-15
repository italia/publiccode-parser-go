package publiccode

import (
	"net/url"
	"time"
)

var legacyKeys = map[string]string{
	"publiccode-yaml-version": "publiccodeYmlVersion",
	"featureList":             "features",
}

// Version of the latest PublicCode specs.
// Source https://github.com/publiccodenet/publiccode.yml
const Version = "0.2"

// SupportedVersions lists the publiccode.yml versions this parser supports.
var SupportedVersions = []string{"0.1", "0.2"}

// PublicCode is a publiccode.yml file definition.
// Reference: https://github.com/publiccodenet/publiccode.yml
type PublicCode struct {
	PubliccodeYamlVersion string `yaml:"publiccodeYmlVersion"`

	Name             string   `yaml:"name"`
	ApplicationSuite string   `yaml:"applicationSuite,omitempty"`
	URL              *url.URL `yaml:"-"`
	URLString        string   `yaml:"url"`
	LandingURL       *url.URL `yaml:"-"`
	LandingURLString string   `yaml:"landingURL,omitempty"`

	IsBasedOn         []string  `yaml:"isBasedOn,omitempty"`
	SoftwareVersion   string    `yaml:"softwareVersion,omitempty"`
	ReleaseDate       time.Time `yaml:"-"`
	ReleaseDateString string    `yaml:"releaseDate"`
	Logo              string    `yaml:"logo,omitempty"`
	MonochromeLogo    string    `yaml:"monochromeLogo,omitempty"`

	InputTypes  []string `yaml:"inputTypes,omitempty"`
	OutputTypes []string `yaml:"outputTypes,omitempty"`

	Platforms []string `yaml:"platforms"`

	Categories []string `yaml:"categories"`

	UsedBy []string `yaml:"usedBy,omitempty"`

	Roadmap       *url.URL `yaml:"-"`
	RoadmapString string   `yaml:"roadmap,omitempty"`

	DevelopmentStatus string `yaml:"developmentStatus"`

	SoftwareType string `yaml:"softwareType"`

	IntendedAudience struct {
		Scope                []string `yaml:"scope,omitempty"`
		Countries            []string `yaml:"countries,omitempty"`
		UnsupportedCountries []string `yaml:"unsupportedCountries,omitempty"`
	} `yaml:"intendedAudience"`

	Description map[string]Desc `yaml:"description"`

	Legal struct {
		License            string `yaml:"license"`
		MainCopyrightOwner string `yaml:"mainCopyrightOwner,omitempty"`
		RepoOwner          string `yaml:"repoOwner,omitempty"`
		AuthorsFile        string `yaml:"authorsFile,omitempty"`
	} `yaml:"legal"`

	Maintenance struct {
		Type        string       `yaml:"type"`
		Contractors []Contractor `yaml:"contractors,omitempty"`
		Contacts    []Contact    `yaml:"contacts,omitempty"`
	} `yaml:"maintenance"`

	Localisation struct {
		LocalisationReady  bool     `yaml:"localisationReady"`
		AvailableLanguages []string `yaml:"availableLanguages"`
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
	GenericName            string     `yaml:"genericName"`
	ShortDescription       string     `yaml:"shortDescription"`
	LongDescription        string     `yaml:"longDescription,omitempty"`
	Documentation          *url.URL   `yaml:"-"`
	DocumentationString    string     `yaml:"documentation,omitempty"`
	APIDocumentation       *url.URL   `yaml:"-"`
	APIDocumentationString string     `yaml:"apiDocumentation,omitempty"`
	Features               []string   `yaml:"features,omitempty"`
	Screenshots            []string   `yaml:"screenshots,omitempty"`
	Videos                 []*url.URL `yaml:"-"`
	VideosStrings          []string   `yaml:"videos,omitempty"`
	Awards                 []string   `yaml:"awards,omitempty"`
}

// Contractor is an entity or entities, if any, that are currently contracted for maintaining the software.
// Reference: https://github.com/publiccodenet/publiccode.yml/blob/develop/schema.md#contractor
type Contractor struct {
	Name          string    `yaml:"name"`
	Website       *url.URL  `yaml:"-"`
	WebsiteString string    `yaml:"website"`
	Until         time.Time `yaml:"-"`
	UntilString   string    `yaml:"until"`
}

// Contact is a contact info maintaining the software.
// Reference: https://github.com/publiccodenet/publiccode.yml/blob/develop/schema.md#contact
type Contact struct {
	Name        string `yaml:"name"`
	Email       string `yaml:"email"`
	Affiliation string `yaml:"affiliation"`
	Phone       string `yaml:"phone"`
}

// Dependency describe system-level dependencies required to install and use this software.
// Reference: https://github.com/publiccodenet/publiccode.yml/blob/develop/schema.md#section-dependencies
type Dependency struct {
	Name       string `yaml:"name"`
	VersionMin string `yaml:"versionMin"`
	VersionMax string `yaml:"versionMax"`
	Optional   bool   `yaml:"optional"`
	Version    string `yaml:"version"`
}
