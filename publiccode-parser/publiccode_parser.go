package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
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
	flag.Usage = func() {
		_, _ = fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [ OPTIONS ] publiccode.yml\n", os.Args[0])

		flag.PrintDefaults()
	}
	localBasePathPtr := flag.String(
		"path", "",
		"Use this local directory as base path when checking for files existence "+
			"instead of using the `url` key in publiccode.yml",
	)
	disableNetworkPtr := flag.Bool(
		"no-network", false,
		"Disables checks that require network connections (URL existence and oEmbed). This makes validation much faster.",
	)
	disableExternalChecksPtr := flag.Bool(
		"no-external-checks", false,
		"Disables ALL checks that reference external resources such as remote URLs or local file existence. "+
			"Implies --no-network",
	)
	timeoutPtr := flag.Duration(
		"timeout", 0,
		"Timeout for each HTTP request during external checks (e.g. 10s, 1m). "+
			"Defaults to 30s if not set. No effect with --no-network or --no-external-checks.",
	)
	jsonOutputPtr := flag.Bool("json", false, "Output the validation errors as a JSON list.")
	helpPtr := flag.Bool("help", false, "Display command line usage.")
	versionPtr := flag.Bool("version", false, "Display current software version.")

	flag.Parse()

	if *versionPtr {
		println(version, date)

		return
	}

	if *helpPtr || len(flag.Args()) < 1 {
		flag.Usage()

		return
	}

	publiccodeFile := flag.Args()[0]

	config := publiccode.ParserConfig{BaseURL: *localBasePathPtr}
	config.DisableNetwork = *disableNetworkPtr
	config.DisableExternalChecks = *disableExternalChecksPtr
	config.Timeout = *timeoutPtr

	p, err := publiccode.NewParser(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating Parser: %s\n", err.Error())
		os.Exit(1)
	}

	_, err = p.Parse(publiccodeFile)

	if *jsonOutputPtr {
		if err == nil {
			fmt.Println("[]")
			os.Exit(0)
		}

		out, jsonerr := json.MarshalIndent(err, "", "    ")
		if jsonerr != nil {
			fmt.Fprintf(os.Stderr, "Error encoding JSON\n")
			os.Exit(1)
		}

		fmt.Println(string(out))

		return
	} else {
		if err != nil {
			fmt.Println(err)
		}

		if hasValidationErrors(err) {
			os.Exit(1)
		}

		os.Exit(0)
	}
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
