package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime/debug"

	publiccode "github.com/italia/publiccode-parser-go/v5"
)

var (
	version string
	date    string
)

var (
	errValidModes       = errors.New("valid modes are none, local and network")
	errCloneMode        = errors.New("the clone mode is not implemented yet")
	errConflictingModes = errors.New("conflicting external check modes")
)

func init() {
	if version == "" {
		version = "devel"
		if info, ok := debug.ReadBuildInfo(); ok {
			version = info.Main.Version
		}
	}

	if date == "" {
		date = "(latest)"
	}
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("publiccode-parser", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.Usage = func() {
		_, _ = fmt.Fprintf(flags.Output(), "Usage: %s [ OPTIONS ] publiccode.yml\n", flags.Name())

		flags.PrintDefaults()
	}
	localBasePathPtr := flags.String(
		"path", "",
		"Use this local directory as base path when checking for files existence "+
			"instead of using the `url` key in publiccode.yml",
	)
	externalChecksPtr := flags.String(
		"external-checks", "network",
		"Which `mode` of checks on resources external to the publiccode.yml to run: "+
			"none (nothing external is checked), "+
			"local (only local file existence and images), "+
			"network (also URL existence and remote images).",
	)
	disableNetworkPtr := flags.Bool(
		"no-network", false,
		"Deprecated, use -external-checks=local. "+
			"Disables checks that require network connections (URL existence and oEmbed).",
	)
	disableExternalChecksPtr := flags.Bool(
		"no-external-checks", false,
		"Deprecated, use -external-checks=none. "+
			"Disables ALL checks that reference external resources such as remote URLs or local file existence.",
	)
	timeoutPtr := flags.Duration(
		"timeout", 0,
		"Timeout for each HTTP request during external checks (e.g. 10s, 1m). "+
			"Defaults to 30s if not set. No effect with -external-checks=none or -external-checks=local.",
	)
	jsonOutputPtr := flags.Bool("json", false, "Output the validation errors as a JSON list.")
	helpPtr := flags.Bool("help", false, "Display command line usage.")
	versionPtr := flags.Bool("version", false, "Display current software version.")

	if err := flags.Parse(args); err != nil {
		return 2
	}

	if *versionPtr {
		fmt.Fprintln(stdout, version, date)

		return 0
	}

	if *helpPtr {
		flags.SetOutput(stdout)
		flags.Usage()

		return 0
	}

	mode, err := resolveCheckMode(checkModeFlags{
		mode:             *externalChecksPtr,
		modeSet:          flagPassed(flags, "external-checks"),
		noNetwork:        *disableNetworkPtr,
		noExternalChecks: *disableExternalChecksPtr,
	}, stderr)
	if err != nil {
		fmt.Fprintln(stderr, err)

		return 2
	}

	if flags.NArg() < 1 {
		flags.Usage()

		return 2
	}

	config := publiccode.ParserConfig{BaseURL: *localBasePathPtr}
	config.ExternalChecks = mode
	config.Timeout = *timeoutPtr

	p, err := publiccode.NewParser(config)
	if err != nil {
		fmt.Fprintf(stderr, "Error creating Parser: %s\n", err.Error())

		return 1
	}

	_, parseErr := p.Parse(flags.Arg(0))

	if *jsonOutputPtr {
		return reportJSON(parseErr, stdout, stderr)
	}

	if parseErr != nil {
		fmt.Fprintln(stdout, parseErr)
	}

	if hasValidationErrors(parseErr) {
		return 1
	}

	return 0
}

// checkModeFlags is the state of the flags selecting which checks on external
// resources run.
type checkModeFlags struct {
	mode    string
	modeSet bool

	noNetwork        bool
	noExternalChecks bool
}

// resolveCheckMode turns the flags into the single mode the parser understands,
// warning about the deprecated ones.
func resolveCheckMode(opts checkModeFlags, stderr io.Writer) (publiccode.CheckMode, error) {
	parsed, err := parseCheckMode(opts.mode)
	if err != nil {
		return parsed, err
	}

	alias, aliasMode := "", parsed

	// -no-external-checks wins over -no-network, the precedence the two
	// booleans have in the parser configuration.
	switch {
	case opts.noExternalChecks:
		alias, aliasMode = "-no-external-checks", publiccode.CheckNone
	case opts.noNetwork:
		alias, aliasMode = "-no-network", publiccode.CheckLocal
	}

	if alias == "" {
		return parsed, nil
	}

	if opts.modeSet && aliasMode != parsed {
		return parsed, fmt.Errorf("%w: %s and -external-checks=%s", errConflictingModes, alias, opts.mode)
	}

	if opts.noNetwork {
		fmt.Fprintln(stderr, "-no-network is deprecated, use -external-checks=local")
	}

	if opts.noExternalChecks {
		fmt.Fprintln(stderr, "-no-external-checks is deprecated, use -external-checks=none")
	}

	return aliasMode, nil
}

func parseCheckMode(mode string) (publiccode.CheckMode, error) {
	switch mode {
	case "none":
		return publiccode.CheckNone, nil
	case "local":
		return publiccode.CheckLocal, nil
	case "network":
		return publiccode.CheckNetwork, nil
	case "clone":
		// Reserved for the mode checking the files in a clone of the repository.
		return publiccode.CheckNetwork, errCloneMode
	default:
		return publiccode.CheckNetwork, fmt.Errorf("invalid value %q for -external-checks: %w", mode, errValidModes)
	}
}

func flagPassed(flags *flag.FlagSet, name string) bool {
	passed := false

	flags.Visit(func(f *flag.Flag) {
		if f.Name == name {
			passed = true
		}
	})

	return passed
}

func reportJSON(parseErr error, stdout, stderr io.Writer) int {
	if parseErr == nil {
		fmt.Fprintln(stdout, "[]")

		return 0
	}

	// Failures that are not validation results, such as an unreadable file, carry
	// unexported fields only and would marshal to an empty JSON object.
	var results publiccode.ValidationResults
	if !errors.As(parseErr, &results) {
		results = publiccode.ValidationResults{publiccode.ValidationError{Description: parseErr.Error()}}
	}

	out, jsonerr := json.MarshalIndent(results, "", "    ")
	if jsonerr != nil {
		fmt.Fprintf(stderr, "Error encoding JSON\n")

		return 1
	}

	fmt.Fprintln(stdout, string(out))

	if hasValidationErrors(parseErr) {
		return 1
	}

	return 0
}

func hasValidationErrors(results error) bool {
	if results == nil {
		return false
	}

	var vr publiccode.ValidationResults
	if errors.As(results, &vr) {
		for _, res := range vr {
			var ve publiccode.ValidationError
			if errors.As(res, &ve) {
				return true
			}
		}

		return false
	}

	return true
}
