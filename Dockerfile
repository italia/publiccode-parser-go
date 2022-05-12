#
# This is for local development only.
# See Dockerfile.goreleaser for the image published on release.
#

# Accept the Go version for the image to be set as a build argument.
ARG GO_VERSION=1.16-alpine

FROM golang:${GO_VERSION} as build

WORKDIR /go/src

COPY . .

RUN cd pcvalidate && \
  go build -ldflags="-s -w" pcvalidate.go

FROM alpine:3

COPY --from=build /go/src/pcvalidate/pcvalidate /usr/local/bin/pcvalidate

ENTRYPOINT ["/usr/local/bin/pcvalidate"]
CMD ["files/publiccode.yml"]
