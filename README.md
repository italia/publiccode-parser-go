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

To get latest development version use:
```
$ go get github.com/italia/publiccode-parser-go/pcvalidate
$ pcvalidate mypubliccode.yml
```

To get latest stable version go to [release page](https://github.com/italia/publiccode-parser-go/releases/latest) and grab which will match your arch.

Run `pcvalidate --help` for the available command line flags.

Run `pcvalidate --version` to print actual version.

The tool returns 0 in case of successful validation, 1 otherwise.

## Easy validation with Docker

You may not want to install Go on your environment, or you'd just want something quick and easy that allows you to validate *publiccode.yml* files in your CI pipeline. With this goal in mind, a Docker container for the parser has also been published.

The image is available on [Dockerhub](https://hub.docker.com/repository/docker/italia/publiccode-parser-go).
You can also build your own starting from the [Dockerfile](Dockerfile) in this repository, running from the root `goreleaser release --skip-publish --skip-validate` (you need goreleaser installed) or use the docker image instead as follow:

```sh
$ docker run --rm --privileged \
  -v $PWD:/go/src/github.com/user/repo \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -w /go/src/github.com/user/repo \
  goreleaser/goreleaser release --skip-publish --skip-validate
```
Or for more dev purpose you can use `docker build -f Dockerfile.dev -t italia/publiccode-parser-go .`

Below some validation examples are reported. The examples assume that your publiccode.yml file is on your local machine, at `/home/my-user/publiccodes/publiccode.yml`

```shell
# Validate a publiccode file called publiccode.yml (default)
docker run -v /home/my-user/publiccodes:/files italia/publiccode-parser-go

# Validate a publiccode file called /home/my-user/publiccodes/my-amazing-code.yaml
docker run -v /home/my-user/publiccodes:/files italia/publiccode-parser-go my-amazing-code.yaml

# Do extra validations of the local publiccode.yml file against the corresponding remote repository
docker run -v /home/my-user/publiccodes:/files italia/publiccode-parser-go -remote-base-url https://github.com/YOUR_GITHUB_USER/YOUR_REPO files/publiccode.yml

# For debugging, access the container interactive shell (/bin/sh), overriding the entrypoint
docker run -it --entrypoint "/bin/sh" italia/publiccode-parser-go
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

© 2018-2020 Team per la Trasformazione Digitale - Presidenza del Consiglio dei Minstri

Licensed under the EUPL.
The version control system provides attribution for specific lines of code.
