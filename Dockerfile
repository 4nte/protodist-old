FROM golang:alpine as protodist-builder

FROM scratch
COPY protodist /
ENTRYPOINT ["/protodist"]
