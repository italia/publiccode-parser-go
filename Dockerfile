# Accept the Go version for the image to be set as a build argument.
ARG GO_VERSION=1.16-alpine

FROM golang:${GO_VERSION} as build

WORKDIR /go/src

COPY . .

RUN cd pcvalidate && \
  go build -ldflags="-s -w" pcvalidate.go

FROM alpine:3

COPY --from=build /go/src/pcvalidate/pcvalidate /usr/local/bin/pcvalidate

# git and openssh-client are needed by CircleCI when using
# publiccode-parser-orb, which uses on this image.
RUN apk --no-cache add git openssh-client

ENTRYPOINT ["/usr/local/bin/pcvalidate"]

CMD ["files/publiccode.yml"]
