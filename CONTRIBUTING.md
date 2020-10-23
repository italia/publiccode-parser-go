# Contributing

🙇‍♀️ Thank you for contributing!

Example steps in order to add a key-val to parse: add `nickname` field.

* Add `nickname` in `tests/valid.yml`

```
publiccodeYmlVersion: "0.1"
name: Medusa

nickname: Meds

applicationSuite: MegaProductivitySuite
url: "https://github.com/italia/developers.italia.it.git"        # URL of this repository
landingURL: "https://developers.italia.it"
...
```

* Add it into publiccode struct in `publiccode.go` (or into the right struct in `extensions.go`)

```
type PublicCode struct {
	PubliccodeYamlVersion string `yaml:"publiccodeYmlVersion"`
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

* Catched! The `nickname` key is detected as a String, and there is no definition in the keys list.

* Open `keys.go` and search the right function that will handle this new String element.
  When found, you can add the right key to the switch case.

```
func (p *Parser) decodeString(key string, value string) (err error) {
	switch {
  ...
  case key == "nickname":
    p.PublicCode.Nickname = value
  ...
  }
}
```

* Done!

* Run go tests again. It should return `ok` and no errors.

```
ok  	publiccode.yml-parser-go	6.665s
```
