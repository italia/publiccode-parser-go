package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"runtime/debug"

	"github.com/alranel/go-vcsurl/v2"

	publiccode "github.com/italia/publiccode-parser-go/v2"
	urlutil "github.com/italia/publiccode-parser-go/v2/internal"
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
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [ OPTIONS ] publiccode.yml\n", os.Args[0])
		flag.PrintDefaults()
	}
	localBasePathPtr := flag.String("path", "", "Use this local directory as base path when checking for files existence instead of using the `url` key in publiccode.yml")
	disableNetworkPtr := flag.Bool("no-network", false, "Disables checks that require network connections (URL existence and oEmbed). This makes validation much faster.")
	exportPtr := flag.String("export", "", "Export the normalized publiccode.yml file to the given path.")
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

	var publiccodeFile = flag.Args()[0]

	if ok, url := urlutil.IsValidURL(publiccodeFile); ok {
		// supplied argument looks like an URL
		if vcsurl.GetRawFile(url) == nil {
			fmt.Fprintf(os.Stderr, "Code hosting provider not supported for %s\n", url)
			os.Exit(1)
		}
	}

	var p *publiccode.Parser
	var err error
	if *localBasePathPtr != "" {
		p, err = publiccode.NewParserWithPath(publiccodeFile, *localBasePathPtr)
	} else {
		p, err = publiccode.NewParser(publiccodeFile)
	}
	p.DisableNetwork = *disableNetworkPtr

	err = p.Parse()

	if *jsonOutputPtr {
		if err == nil {
			fmt.Println("[]")
			os.Exit(0)
		}

		out, jsonerr := json.MarshalIndent(err, "", "    ")
		if (jsonerr != nil) {
			fmt.Fprintf(os.Stderr, "Error encoding JSON\n")
			os.Exit(1)
		}

		fmt.Println(string(out))

		return
	} else {
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	if *exportPtr != "" {
		yaml, err := p.ToYAML()
		err = ioutil.WriteFile(*exportPtr, yaml, 0644)
		if err != nil {
			panic(err)
		}
		fmt.Printf("publiccode written to %s\n", *exportPtr)
	}
}

// isValidURL tests a string to determine if it is a well-structured url or not.
func isValidURL(toTest string) (bool, *url.URL) {
	_, err := url.ParseRequestURI(toTest)
	if err != nil {
		return false, nil
	}

	u, err := url.Parse(toTest)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return false, nil
	}

	// Check it's an acceptable scheme
	switch u.Scheme {
	case "http":
	case "https":
	default:
		return false, nil
	}

	return true, u
}
