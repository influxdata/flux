#!/bin/bash

FLATBUFFERS_VERSION=2.0.0
FLATBUFFERS_CHECKSUM=9ddb9031798f4f8754d00fca2f1a68ecf9d0f83dfac7239af1311e4fd9a565c4
curl -LS https://github.com/google/flatbuffers/archive/v${FLATBUFFERS_VERSION}.tar.gz -O && \
    echo "${FLATBUFFERS_CHECKSUM} v${FLATBUFFERS_VERSION}.tar.gz" | sha256sum --check -- && \
    tar xvzf v${FLATBUFFERS_VERSION}.tar.gz && \
    mkdir flatbuffers-${FLATBUFFERS_VERSION}/build && \
    cd flatbuffers-${FLATBUFFERS_VERSION}/build && \
    cmake -G "Unix Makefiles" .. && \
    make && make install && \
    cd ../.. && rm -rf flatbuffers-${FLATBUFFERS_VERSION} v${FLATBUFFERS_VERSION}.tar.gz
