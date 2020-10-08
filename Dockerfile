FROM antegulin/protobuf-builder:master

COPY protodist /
ENTRYPOINT ["/protodist"]
