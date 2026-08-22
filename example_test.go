package publiccode_test

import (
	"fmt"
	"strings"

	publiccode "github.com/italia/publiccode-parser-go/v5"
)

func ExampleNewDefaultParser() {
	parser, err := publiccode.NewDefaultParser()
	if err != nil {
		panic(err)
	}

	_, err = parser.Parse("file:///path/to/publiccode.yml")
	if err != nil {
		fmt.Println(err)
	}
}

func ExampleNewParser_disableNetwork() {
	parser, err := publiccode.NewParser(publiccode.ParserConfig{
		DisableNetwork: true,
	})
	if err != nil {
		panic(err)
	}

	_, err = parser.Parse("file:///path/to/publiccode.yml")
	if err != nil {
		fmt.Println(err)
	}
}

func ExampleParser_ParseStream() {
	yaml := `
publiccodeYmlVersion: "0"
name: My Software
url: https://github.com/example/example
`

	parser, err := publiccode.NewParser(publiccode.ParserConfig{
		DisableNetwork: true,
	})
	if err != nil {
		panic(err)
	}

	_, err = parser.ParseStream(strings.NewReader(yaml))
	if err != nil {
		fmt.Println(err)
	}
}
