# Accept the Go version for the image to be set as a build argument.
# Defaults to Go 1.13-alpine
ARG GO_VERSION=1.13-alpine

# First stage: build the executable.
FROM golang:${GO_VERSION} AS builder

# Set environment vars
ENV USER=nobody
ENV BASE_PATH=/go/bin
ENV MOUNT_PATH=${BASE_PATH}/files
ENV PC_VALIDATE_URL=github.com/italia/publiccode-parser-go/pcvalidate

# Install required software
RUN apk add git

# Get deps for both components
RUN go get ${PC_VALIDATE_URL}

# Create destination folders and external mount points
RUN mkdir -p ${MOUNT_PATH}

# Fix ownership to run the binary as non root
RUN chown -R ${USER}:${USER} ${BASE_PATH}

# Perform any further action as an unprivileged user.
USER ${USER}:${USER}

# Change workdir
WORKDIR ${BASE_PATH}

# Final stage: the running container.
FROM alpine AS final

COPY --from=builder /go/bin/pcvalidate /

# Run the compiled binary.
ENTRYPOINT ["/pcvalidate"]
CMD ["files/publiccode.yml"]