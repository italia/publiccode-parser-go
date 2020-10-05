package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"runtime/debug"

	vcsurl "github.com/alranel/go-vcsurl"
	publiccode "github.com/italia/publiccode-parser-go"
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
	remoteBaseURLPtr := flag.String("remote-base-url", "", "The URL pointing to the directory where the publiccode.yml file is located.")
	localBasePathPtr := flag.String("path", "", "An absolute or relative path pointing to a locally cloned repository where the publiccode.yml is located.")
	disableNetworkPtr := flag.Bool("no-network", false, "Disables checks that require network connections (URL existence and oEmbed). This makes validation much faster.")
	exportPtr := flag.String("export", "", "Export the normalized publiccode.yml file to the given path.")
	noStrictPtr := flag.Bool("no-strict", false, "Disable strict mode.")
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

	p := publiccode.NewParser()
	p.RemoteBaseURL = *remoteBaseURLPtr
	p.LocalBasePath = *localBasePathPtr
	p.DisableNetwork = *disableNetworkPtr
	p.Strict = !*noStrictPtr

	var err error
	if ok, url := isValidURL(flag.Args()[0]); ok {
		// supplied argument looks like an URL
		rawUrl := vcsurl.GetRawFile(url)
		if rawUrl == nil {
			fmt.Fprintf(os.Stderr, "Code hosting provider not supported for %s\n", url)
			os.Exit(1)
		}
		if p.RemoteBaseURL == "" {
			p.RemoteBaseURL = vcsurl.GetRawRoot(rawUrl).String()
		}

		err = p.ParseRemoteFile(rawUrl.String())
	} else {
		// supplied argument looks like a file path
		err = p.ParseFile(flag.Args()[0])
	}
	if err != nil {
		fmt.Printf("validation ko:\n%v\n", err)
		os.Exit(1)
	}
	fmt.Println("validation ok")

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
