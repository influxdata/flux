# This file describes an image that is capable of building Flux.

FROM golang:1.12 AS dev

# Install common packages
RUN apt-get update && \
    apt-get install --no-install-recommends -y \
    ruby \
    ca-certificates curl file \
    build-essential \
    autoconf automake autotools-dev libtinfo5 libtool xutils-dev && \
    rm -rf /var/lib/apt/lists/*

# Install ragel
ENV RAGEL_VERSION=6.10
RUN curl http://www.colm.net/files/ragel/ragel-${RAGEL_VERSION}.tar.gz -O && \
    tar -xzf ragel-${RAGEL_VERSION}.tar.gz && \
    cd ragel-${RAGEL_VERSION}/ && \
    ./configure --prefix=/usr/local && \
    make && \
    make install && \
    cd .. && rm -rf ragel-${RAGEL_VERSION}*
ENV PATH="/usr/local/bin:${PATH}"

# Install and configure openssl - needed for proper Rust install
ENV SSL_VERSION=1.0.2q

RUN curl https://www.openssl.org/source/openssl-$SSL_VERSION.tar.gz -O && \
    tar -xzf openssl-$SSL_VERSION.tar.gz && \
    cd openssl-$SSL_VERSION && ./config && make depend && make install && \
    cd .. && rm -rf openssl-$SSL_VERSION*

ENV OPENSSL_LIB_DIR=/usr/local/ssl/lib \
    OPENSSL_INCLUDE_DIR=/usr/local/ssl/include \
    OPENSSL_STATIC=1

# Install Clang
RUN curl -SL http://releases.llvm.org/8.0.0/clang+llvm-8.0.0-x86_64-linux-gnu-ubuntu-18.04.tar.xz \
    | tar -xJC . && \
    mv clang+llvm-8.0.0-x86_64-linux-gnu-ubuntu-18.04 clang_8.0.0

ENV PATH="/go/clang_8.0.0/bin:${PATH}" \
    LD_LIBRARY_PATH="/clang_8.0.0/lib:${LD_LIBRARY_PATH}" \
    CC=clang

# Add builder user
ENV UNAME=builder
ARG UID=1000
ARG GID=1000
RUN groupadd -g $GID -o $UNAME
RUN useradd -m -u $UID -g $UNAME -s /bin/bash $UNAME
USER $UNAME
ENV HOME=/home/$UNAME

# Install Rust
RUN curl https://sh.rustup.rs -sSf | \
    sh -s -- --default-toolchain stable -y
ENV PATH="$HOME/.cargo/bin:${PATH}"
RUN rustup component add rustfmt

# Install wasm-pack and sccache
RUN cargo install wasm-pack sccache
RUN rustup component add rust-std --target wasm32-unknown-unknown

# Use sccache rustc wrapper for friendly build caching
ENV RUSTC_WRAPPER=sccache

ENV GOPATH=/home/$UNAME/go
RUN mkdir -p ${GOPATH}/src/github.com/influxdata/flux
WORKDIR ${GOPATH}/src/github.com/influxdata/flux

FROM dev AS build

COPY --chown=$UNAME Makefile go.* ${GOPATH}/src/github.com/influxdata/flux/

RUN make deps depsfix

COPY --chown=$UNAME . ${GOPATH}/src/github.com/influxdata/flux

RUN make bin/linux/cmpgen
RUN make flux

FROM debian:buster AS release

COPY --from=build /home/builder/go/src/github.com/influxdata/flux/flux /bin/flux

ENTRYPOINT [ "/bin/flux" ]
CMD [ "repl" ]
