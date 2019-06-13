# publiccode.yml parser for Go

[![Join the #publiccode channel](https://img.shields.io/badge/Slack%20channel-%23publiccode-blue.svg?logo=slack)](https://developersitalia.slack.com/messages/CAM3F785T)
[![Get invited](https://slack.developers.italia.it/badge.svg)](https://slack.developers.italia.it/) [![CircleCI](https://circleci.com/gh/italia/publiccode-parser-go.svg?style=svg)](https://circleci.com/gh/italia/publiccode-parser-go)

This is a Go parser and validator for [publiccode.yml](https://github.com/italia/publiccode.yml) files.

publiccode.yml is an international standard for describing public software. It is expected to be published in the root of open source repositories. This parser performs syntactic and semantic validation according to the official spec.

## Features

- Support for the Italian extension
- Check ISO 3166-1 alpha-2 and alpha-3 country codes
- Validate emails, URLs (http scheme, valid status code), local and remote files, dates as "YYYY-MM-DD", images (colors, mimes)
- Check pa-types
- Validate oembed video links and retrieve HTML for easy embedding
- Validate SPDX licenses. Without WITH keyword.
- Check tags
- Strict and non-strict modes (use non-strict when you want to be tolerant, such as in a crawler, but use strict in editors and validators)

## Example

```go
parser := publiccode.NewParser()

// all these settings are optional:
parser.LocalBasePath = "/path/to/local/clone"
parser.RemoteBaseURL = "https://raw.githubusercontent.com/gith002/Medusa/master"
parser.DisableNetwork = false

err := parser.ParseRemoteFile(url)
pc := parser.PublicCode
```

## Validation from command line

This repository also contains an executable tool which can be used for validating a publiccode.yml file locally.

```
$ go get github.com/italia/publiccode-parser-go/pcvalidate
$ pcvalidate mypubliccode.yml
```

Run `pcvalidate --help` for the available command line flags.

## Assets

- [data/amministrazioni.txt](data/amministrazioni.txt) updated on: _2018-07-12_.
- [data/oembed_providers.json](data/oembed_providers.json) updated on: _2018-07-12_.

In order to update the assets, run this command:

`go-bindata -o assets.go data/`

And change the package name into `publiccode`

## Contributing

Contributing is always appreciated.
Feel free to open issues, fork or submit a Pull Request.
If you want to know more about how to add new fields, check out [CONTRIBUTING.md](CONTRIBUTING.md). In order to support other country-specific extensions in addition to Italy some refactoring might be needed.

## See also

* [Developers Italia backend & crawler](https://github.com/italia/developers-italia-backend) - a Go crawler that uses this library

## Maintainers

This software is maintained by the [Developers Italia](https://developers.italia.it/) team.

## License

© 2018-2019 Team per la Trasformazione Digitale - Presidenza del Consiglio dei Minstri

Licensed under the EUPL.
The version control system provides attribution for specific lines of code.
