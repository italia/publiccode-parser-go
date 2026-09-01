<!-- markdownlint-disable MD033 -->
<!-- MD033/no-inline-html -->
# libpubliccode

[![Join the #publiccode channel](https://img.shields.io/badge/Slack%20channel-%23publiccode-blue.svg?logo=slack)](https://developersitalia.slack.com/messages/CAM3F785T)
[![Get invited](https://slack.developers.italia.it/badge.svg)](https://slack.developers.italia.it/)

The core parser and validator for
[publiccode.yml](https://github.com/publiccodeyml/publiccode.yml), implemented
in Go and available both as a Go library and as the `publiccode-parser` command.

`publiccode.yml` is an international standard for describing public software, which
should be placed at the root of Free/Libre and Open Source software repositories.

This parser performs syntactic and semantic validation according to the
[official spec](https://yml.publiccode.tools).

## Backwards compatibility

This project was previously named `publiccode-parser-go`. During the rename,
the existing integration points remain unchanged:

- Go module: `github.com/italia/publiccode-parser-go/v5`
- command: `publiccode-parser`
- Docker image: `italia/publiccode-parser-go`

Existing consumers do not need to change their imports, installation commands,
or deployment configuration. New canonical integration names will only be
introduced in a future major release with a documented migration path.

## Features

- Go library and CLI tool (`publiccode-parser`)
- Supports the latest version of the `publiccode.yml` Standard
- `publiccode-parser` can output validation errors as JSON or in
  [errorformat](https://vim-jp.org/vimdoc-en/quickfix.html#error-file-format)
  friendly way
- Verifies the existence of URLs by checking the response for URL fields
  (can be disabled)

## As a library

```go
parser, err := publiccode.NewDefaultParser()

// error handling

publiccode, err := parser.Parse("/path/to/local/dir/publiccode.yml")
// OR
// parse.Parse("https://github.com/example/example/publiccode.yml")
```

[![Go Reference](https://pkg.go.dev/badge/github.com/italia/publiccode-parser-go/v5.svg)](https://pkg.go.dev/github.com/italia/publiccode-parser-go/v5)

## From command line

The `publiccode-parser` binary can be used to validate a `publiccode.yml`
from the command line.

To get the latest version use:

```shell
go install github.com/italia/publiccode-parser-go/v5/publiccode-parser@latest
```

Or get a precompiled package from the [release page](https://github.com/italia/publiccode-parser-go/releases/latest)

Example:

```shell
$ publiccode-parser mypubliccode.yml
mypubliccode.yml:36:1: error: developmentStatus: developmentStatus must be one of the following: "concept", "development", "beta", "stable" or "obsolete"
mypubliccode.yml:48:3: warning: legal.authorsFile: This key is DEPRECATED and will be removed in the future. It's safe to drop it
mypubliccode.yml:12:5: warning: description.en.genericName: This key is DEPRECATED and will be removed in the future. It's safe to drop it
```

Run `publiccode-parser --help` for the available command line flags.

The tool returns 0 in case of successful validation, 1 when the file has
validation errors or can't be read and 2 when it's invoked with no file
argument or an unknown flag.

With `--json` the same exit codes apply and the output is always a
JSON list, empty when there is nothing to report.

## Contributing

Contributing is always appreciated.
Feel free to open issues, fork or submit a Pull Request.
If you want to know more about how to add new fields, check out the
[publiccode.yml project](https://github.com/publiccodeyml/publiccode.yml)
and its [CONTRIBUTING.md](https://github.com/publiccodeyml/publiccode.yml/blob/main/CONTRIBUTING.md)
guidelines.

## See also

- [publiccode-parser-php](https://github.com/bfabio/publiccode-parser-php) - PHP
  bindings for this library
- [publiccode-crawler](https://github.com/italia/publiccode-crawler) - a Go
  crawler that uses this library

## Maintainers

This software is maintained by community contributors.

## License

© 2018-present Team per la Trasformazione Digitale - Presidenza del Consiglio
dei Ministri

Licensed under the EUPL 1.2.
The version control system provides attribution for specific lines of code.
