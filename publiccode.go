package publiccode

// SupportedVersions lists the publiccode.yml versions this parser supports.
var SupportedVersions = []string{"0", "0.2", "0.2.0", "0.2.1", "0.2.2", "0.3", "0.3.0", "0.4", "0.4.0", "0.5", "0.5.0"}

type PublicCode interface {
	Version() uint
	Url() *URL
	ToYAML() ([]byte, error)
}
