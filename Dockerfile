# Accept the Go version for the image to be set as a build argument.
ARG GO_VERSION=1.15-alpine

FROM golang:${GO_VERSION}

ENV BIN /usr/local/bin/pcvalidate

WORKDIR /go/src

COPY . .

RUN go build -o $BIN pcvalidate/pcvalidate.go \
  && chmod +x $BIN

ENTRYPOINT ["/usr/local/bin/pcvalidate"]

CMD ["files/publiccode.yml"]
