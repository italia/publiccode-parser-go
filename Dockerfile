# Accept the Go version for the image to be set as a build argument.
ARG GO_VERSION=1.15-alpine

FROM golang:${GO_VERSION}

ENV BIN /usr/local/bin/pcvalidate

WORKDIR /go/src

COPY . .

# git and openssh-client are needed by CircleCI when using
# publiccode-parser-orb, which uses on this image.
RUN apk --no-cache add git openssh-client \
  && go build -o $BIN pcvalidate/pcvalidate.go \
  && chmod +x $BIN

ENTRYPOINT ["/usr/local/bin/pcvalidate"]

CMD ["files/publiccode.yml"]
