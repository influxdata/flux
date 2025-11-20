#!/bin/bash
FLATBUFFERS_VERSION=25.9.23
FLATBUFFERS_CHECKSUM=9102253214dea6ae10c2ac966ea1ed2155d22202390b532d1dea64935c518ada
curl -LS https://github.com/google/flatbuffers/archive/v${FLATBUFFERS_VERSION}.tar.gz -O && \
    echo "${FLATBUFFERS_CHECKSUM} v${FLATBUFFERS_VERSION}.tar.gz" | sha256sum --check -- && \
    tar xvzf v${FLATBUFFERS_VERSION}.tar.gz && \
    mkdir flatbuffers-${FLATBUFFERS_VERSION}/build && \
    cd flatbuffers-${FLATBUFFERS_VERSION}/build && \
    CC=/usr/bin/gcc-12 CXX=/usr/bin/g++-12 cmake -G "Unix Makefiles" .. && \
    make && make install && \
    cd ../.. && rm -rf flatbuffers-${FLATBUFFERS_VERSION} v${FLATBUFFERS_VERSION}.tar.gz
