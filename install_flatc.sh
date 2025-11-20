#!/bin/bash
FLATBUFFERS_VERSION=23.5.26
FLATBUFFERS_CHECKSUM=1cce06b17cddd896b6d73cc047e36a254fb8df4d7ea18a46acf16c4c0cd3f3f3
curl -LS https://github.com/google/flatbuffers/archive/v${FLATBUFFERS_VERSION}.tar.gz -O && \
    echo "${FLATBUFFERS_CHECKSUM} v${FLATBUFFERS_VERSION}.tar.gz" | sha256sum --check -- && \
    tar xvzf v${FLATBUFFERS_VERSION}.tar.gz && \
    mkdir flatbuffers-${FLATBUFFERS_VERSION}/build && \
    cd flatbuffers-${FLATBUFFERS_VERSION}/build && \
    CC=/usr/bin/gcc-12 CXX=/usr/bin/g++-12 cmake -G "Unix Makefiles" .. && \
    make && make install && \
    cd ../.. && rm -rf flatbuffers-${FLATBUFFERS_VERSION} v${FLATBUFFERS_VERSION}.tar.gz
