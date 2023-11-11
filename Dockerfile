#
# This is for local development only.
# See Dockerfile.goreleaser for the image published on release.
#

# Accept the Go version for the image to be set as a build argument.
ARG GO_VERSION=1.18-alpine

FROM docker.io/golang:${GO_VERSION} as build

WORKDIR /go/src

COPY . .

RUN cd publiccode-parser && go build -ldflags="-s -w"

FROM docker.io/alpine:3

COPY --from=build /go/src/publiccode-parser/publiccode-parser /usr/local/bin/publiccode-parser

# Keep the old name for backward compatibility
RUN ln -s /usr/local/bin/publiccode-parser /usr/local/bin/pcvalidate

ENTRYPOINT ["/usr/local/bin/publiccode-parser"]
CMD ["files/publiccode.yml"]
