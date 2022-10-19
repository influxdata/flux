#!/bin/bash
FLATBUFFERS_VERSION=22.9.29
FLATBUFFERS_CHECKSUM=372df01795c670f6538055a7932fc7eb3e81b3653be4a216c081e9c3c26b1b6d
curl -LS https://github.com/google/flatbuffers/archive/v${FLATBUFFERS_VERSION}.tar.gz -O && \
    echo "${FLATBUFFERS_CHECKSUM} v${FLATBUFFERS_VERSION}.tar.gz" | sha256sum --check -- && \
    tar xvzf v${FLATBUFFERS_VERSION}.tar.gz && \
    mkdir flatbuffers-${FLATBUFFERS_VERSION}/build && \
    cd flatbuffers-${FLATBUFFERS_VERSION}/build && \
    cmake -G "Unix Makefiles" .. && \
    make && make install && \
    cd ../.. && rm -rf flatbuffers-${FLATBUFFERS_VERSION} v${FLATBUFFERS_VERSION}.tar.gz
