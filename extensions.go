package publiccode

// Country-specific extensions
//
// While the standard is structured to be meaningful on an international level,
// there are additional information that can be added that makes sense in specific
// countries, such as declaring compliance with local laws or regulations.
// The provided extension mechanism is the usage of country-specific sections.
//
// All country-specific sections are contained in a section named with
// the two-letter lowercase ISO 3166-1 alpha-2 country code.
//
// For instance "spid" is a property for Italian software declaring whether
// the software is integrated with the Italian Public Identification System.
// Reference: https://github.com/publiccodenet/publiccode.yml/blob/develop/schema.md

// ExtensionITVersion declares the latest supported version of the 'it' extension
var ExtensionITVersion = "0.2"

// ExtensionITSupportedVersions declares the versions of the 'it' extension
// supported by this parser. We also support legacy publiccode.yml files
// which did not contain the it/countryExtensionVersion key.
var ExtensionITSupportedVersions = []string{"0.2"}

// ExtensionIT is the country-specific extension for Italy.
// Reference: https://github.com/publiccodenet/publiccode.yml/blob/develop/schema.it.md
type ExtensionIT struct {
	CountryExtensionVersion string `yaml:"countryExtensionVersion"`

	Conforme struct {
		LineeGuidaDesign        bool `yaml:"lineeGuidaDesign"`
		ModelloInteroperabilita bool `yaml:"modelloInteroperabilita"`
		MisureMinimeSicurezza   bool `yaml:"misureMinimeSicurezza"`
		GDPR                    bool `yaml:"gdpr"`
	} `yaml:"conforme"`

	Riuso struct {
		CodiceIPA string `yaml:"codiceIPA,omitempty"`
	} `yaml:"riuso"`

	Piattaforme struct {
		Spid   bool `yaml:"spid"`
		Pagopa bool `yaml:"pagopa"`
		Cie    bool `yaml:"cie"`
		Anpr   bool `yaml:"anpr"`
		Io     bool `yaml:"io"`
	} `yaml:"piattaforme"`
}
