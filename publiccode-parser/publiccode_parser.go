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
	disableNetworkPtr := flags.Bool(
		"no-network", false,
		"Disables checks that require network connections (URL existence and oEmbed). This makes validation much faster.",
	)
	disableExternalChecksPtr := flags.Bool(
		"no-external-checks", false,
		"Disables ALL checks that reference external resources such as remote URLs or local file existence. "+
			"Implies --no-network",
	)
	timeoutPtr := flags.Duration(
		"timeout", 0,
		"Timeout for each HTTP request during external checks (e.g. 10s, 1m). "+
			"Defaults to 30s if not set. No effect with --no-network or --no-external-checks.",
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

	if flags.NArg() < 1 {
		flags.Usage()

		return 2
	}

	config := publiccode.ParserConfig{BaseURL: *localBasePathPtr}
	config.DisableNetwork = *disableNetworkPtr
	config.DisableExternalChecks = *disableExternalChecksPtr
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
