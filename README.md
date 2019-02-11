# publiccode.yml parser for Go

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

## Example

```go
parser := publiccode.NewParser()
parser.LocalBasePath = "/path/to/local/clone"
parser.RemoteBaseURL = "https://raw.githubusercontent.com/gith002/Medusa/master"

err := parser.ParseRemoteFile(url)
pc := parser.PublicCode
```

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
