FROM alpine

COPY pcvalidate /

# Run the compiled binary.
ENTRYPOINT ["/pcvalidate"]
CMD ["files/publiccode.yml"]