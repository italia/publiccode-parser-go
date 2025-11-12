<!-- markdownlint-disable MD033 -->
<!-- MD033/no-inline-html -->
# publiccode.yml parser for Go

[![Join the #publiccode channel](https://img.shields.io/badge/Slack%20channel-%23publiccode-blue.svg?logo=slack)](https://developersitalia.slack.com/messages/CAM3F785T)
[![Get invited](https://slack.developers.italia.it/badge.svg)](https://slack.developers.italia.it/)

A Go parser and validator for [publiccode.yml](https://github.com/publiccodeyml/publiccode.yml)
files.

`publiccode.yml` is an international standard for describing public software, which
should be placed at the root of Free/Libre and Open Source software repositories.

This parser performs syntactic and semantic validation according to the
[official spec](https://yml.publiccode.tools).

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

publiccode, err := parser.Parse("file:///path/to/local/dir/publiccode.yml")
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
publiccode.yml:36:1: error: developmentStatus: developmentStatus must be one of the following: "concept", "development", "beta", "stable" or "obsolete"
publiccode.yml:48:3: warning: legal.authorsFile: This key is DEPRECATED and will be removed in the future. It's safe to drop it
publiccode.yml:12:5: warning: description.en.genericName: This key is DEPRECATED and will be removed in the future. It's safe to drop it
```

Run `publiccode-parser --help` for the available command line flags.

The tool returns 0 in case of successful validation, 1 otherwise.

## With Docker

You can easily validate your files using Docker on your local machine or in your
CI pipeline:

```shell
docker run -i italia/publiccode-parser-go /dev/stdin < publiccode.yml
```

The image is available on [Dockerhub](https://hub.docker.com/repository/docker/italia/publiccode-parser-go).
You can also build your own running:

```sh
docker build -t italia/publiccode-parser-go .
```

### Examples

<details>
  <summary>Click to expand</summary>

The examples assume that your `publiccode.yml` file is on your local machine,
at `/opt/publiccodes/publiccode.yml`

- Validate and print the canonicalized file

  ```shell
  docker run -i italia/publiccode-parser-go -export /dev/stdout /dev/stdin < publiccode.yml
  ```

- Validate a publiccode file named `publiccode.yml` in `/home/user`

  ```shell
  docker run -v /home/user:/go/src/files italia/publiccode-parser-go
  ```

- Validate a publiccode file named `/opt/publiccodes/my-amazing-code.yaml`

  ```shell
  docker run -v /opt/publiccodes:/go/src/files italia/publiccode-parser-go my-amazing-code.yaml
  ```

- Validate `publiccode.yml` without using the network (fe. checking URLs are reachable)

  ```shell
  docker run -v /opt/publiccodes/publiccodes:/files italia/publiccode-parser-go -no-network publiccode.yml
  ```

- Debugging, access the container interactive shell, overriding the entrypoint

  ```shell
  docker run -it --entrypoint /bin/sh italia/publiccode-parser-go
  ```

</details>

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
