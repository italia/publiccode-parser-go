# publiccode.yml Go parser

A Go parser for publiccode.yml files

**Features**

* Parse and validate a standard publiccode.yml (no extensions)

**Contributing**

Example steps in order to add a key-val (add `nickname` field).

* Add `nickname` in `valid.yml`

```
publiccode-yaml-version: "http://w3id.org/publiccode/version/0.1"
name: Medusa

nickname: Meds

applicationSuite: MegaProductivitySuite
url: "https://github.com/italia/developers.italia.it.git"        # URL of this repository
landingURL: "https://developers.italia.it"
...
```

* Add it into publiccode struct in `publiccode.go` (or its `extensions.go`)

```
type PublicCode struct {
	PubliccodeYamlVersion string `yaml:"publiccode-yaml-version"`
	...

  Nickname         string   `yaml:"nickname"`

  ...
}
```

* Run go tests.

```
go test -race .

--- FAIL: TestDecodeValueErrors (5.22s)
    --- FAIL: TestDecodeValueErrors/#00 (5.22s)
    	parser_test.go:54: unexpected error:
    		 invalid key: nickname : String
FAIL
FAIL	publiccode.yml-parser-go	5.255s
```

* Catched! `nickname` key is detected as String, and there is no definition in the keys list.

* Open `keys.go` and search the right function that will handle the new String element.
  When found, add the right key to the switch case.

```
func (p *parser) decodeString(key string, value string) (err error) {
	switch {
  ...
  case key == "nickname":
    p.pc.Nickname = value
  ...
  }
}
```

* Done!

* Run go tests again. It should return `ok` and no errors.

```
ok  	publiccode.yml-parser-go	6.665s
```

## License

© 2018 Foundation For Public Code and all contributors – Licensed under the EUPL
