package publiccode

import (
	"fmt"
	"runtime/debug"
	"strings"
)

const (
	// productName is the token identifying this software in the User-Agent.
	productName = "publiccode-parser-go"

	// projectURL is part of the User-Agent so that whoever finds these requests
	// in their logs can tell what made them, and allow them explicitly instead
	// of having to allow every Go program.
	projectURL = "https://github.com/italia/publiccode-parser-go"

	// modulePath is this module's import path. The version is looked up under it
	// in the build information: the parser is the main module when its own CLI
	// runs, and a dependency when the library is embedded in another program.
	modulePath = "github.com/italia/publiccode-parser-go/v5"

	// unknownVersion stands in when the build information carries no usable
	// version, as happens with "go run", "go test" and unstamped builds.
	unknownVersion = "devel"
)

// userAgent is resolved once: the build information doesn't change at runtime.
var userAgent = buildUserAgent(buildInfo())

// UserAgent returns the value the parser sends in the User-Agent header of the
// requests it makes during the external checks, e.g.
//
//	publiccode-parser-go/5.4.3 (+https://github.com/italia/publiccode-parser-go)
//
// The version comes from the build information of the running binary, and is
// "devel" when there is none to be found.
//
// Programs embedding the parser should identify themselves instead, through
// [ParserConfig.UserAgent]. This function is exported for the ones that want to
// keep this token and add their own to it.
func UserAgent() string {
	return userAgent
}

// UserAgentForVersion returns the same User-Agent as [UserAgent], for the
// version passed instead of the one in the build information.
//
// It exists for binaries that have their version stamped in at link time, as
// the publiccode-parser CLI does: they know it, while the build information of
// a binary built from a list of files carries no module version at all.
func UserAgentForVersion(version string) string {
	return fmt.Sprintf("%s/%s (+%s)", productName, normalizeVersion(version), projectURL)
}

// buildInfo returns the build information of the running binary, or nil when it
// is unavailable.
func buildInfo() *debug.BuildInfo {
	info, _ := debug.ReadBuildInfo()

	return info
}

// buildUserAgent formats the User-Agent for the version of this module recorded
// in info.
func buildUserAgent(info *debug.BuildInfo) string {
	return UserAgentForVersion(moduleVersion(info))
}

// moduleVersion digs the version of this module out of info. It answers an empty
// string when info records none, which is the case for a binary built from a
// list of files (as the released CLI is) and when info itself is nil.
func moduleVersion(info *debug.BuildInfo) string {
	if info == nil {
		return ""
	}

	// The main module is this one when one of its own commands runs: its path is
	// the module path for the module itself, and below it for a command.
	if info.Main.Path == modulePath || strings.HasPrefix(info.Main.Path, modulePath+"/") {
		return info.Main.Version
	}

	// The parser is a dependency when the library is embedded in another program.
	for _, dep := range info.Deps {
		if dep != nil && dep.Path == modulePath {
			return dep.Version
		}
	}

	return ""
}

// normalizeVersion turns a Go module version into what goes in the User-Agent:
// a bare version without the "v" prefix, or unknownVersion when there is
// nothing usable. A pseudo-version is kept as it is, since it identifies the
// commit.
func normalizeVersion(version string) string {
	version = strings.TrimPrefix(version, "v")

	// An unstamped main module reports "(devel)".
	if version == "" || strings.HasPrefix(version, "(") {
		return unknownVersion
	}

	return version
}
