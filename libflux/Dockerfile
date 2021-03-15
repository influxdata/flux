FROM ubuntu:18.04

# Install common packages
RUN apt-get update && \
    apt-get install --no-install-recommends -y \
    ca-certificates curl file \
    build-essential \
    gnupg \
    autoconf automake autotools-dev libtool xutils-dev && \
    rm -rf /var/lib/apt/lists/*

# Download, verify, and install ragel version 6.
ENV RAGEL_VERSION=6.10
RUN curl https://www.colm.net/files/thurston.asc | gpg --import - && \
    curl https://www.colm.net/files/ragel/ragel-${RAGEL_VERSION}.tar.gz -O && \
    curl https://www.colm.net/files/ragel/ragel-${RAGEL_VERSION}.tar.gz.asc -O && \
    gpg --verify ragel-${RAGEL_VERSION}.tar.gz.asc ragel-${RAGEL_VERSION}.tar.gz && \
    tar -xzf ragel-${RAGEL_VERSION}.tar.gz && \
    cd ragel-${RAGEL_VERSION}/ && \
    ./configure --prefix=/usr/local && \
    make && \
    make install && \
    cd .. && rm -rf ragel-${RAGEL_VERSION}*

# Download, verify, and install ragel version 7. This version depends on colm,
# so we get that as well. Alredy have the gpg key above.
ENV COLM_VERSION=0.14.2
RUN curl https://www.colm.net/files/colm/colm-${COLM_VERSION}.tar.gz -O && \
    curl https://www.colm.net/files/colm/colm-${COLM_VERSION}.tar.gz.asc -O && \
    gpg --verify colm-${COLM_VERSION}.tar.gz.asc colm-${COLM_VERSION}.tar.gz && \
    tar -xzf colm-${COLM_VERSION}.tar.gz && \
    cd colm-${COLM_VERSION}/ && \
    ./configure --prefix=/usr/local/ragel7 --disable-manual && \
    make && \
    make install && \
    cd .. && rm -rf colm-${COLM_VERSION}*

ENV RAGEL7_VERSION=7.0.1
RUN curl https://www.colm.net/files/ragel/ragel-${RAGEL7_VERSION}.tar.gz -O && \
    curl https://www.colm.net/files/ragel/ragel-${RAGEL7_VERSION}.tar.gz.asc -O && \
    gpg --verify ragel-${RAGEL7_VERSION}.tar.gz.asc ragel-${RAGEL7_VERSION}.tar.gz && \
    tar -xzf ragel-${RAGEL7_VERSION}.tar.gz && \
    cd ragel-${RAGEL7_VERSION}/ && \
    ./configure --prefix=/usr/local/ragel7 --with-colm=/usr/local/ragel7 --disable-manual && \
    make && \
    make install && \
    cd .. && rm -rf ragel-${RAGEL7_VERSION}*

# We installed ragel version 7 to a dedicated prefix. We put it on the path
# after the place where ragel version 6 is located. Version 7 also has the
# ragel binary, but we only need the ragel-rust binary from version 7.
ENV PATH="/usr/local/bin:/usr/local/ragel7/bin:${PATH}"


# Download, verify, install, and configure openssl (needed for proper Rust install)
ENV SSL_VERSION=1.0.2q
RUN curl https://www.openssl.org/source/openssl-$SSL_VERSION.tar.gz -O && \
    curl https://www.openssl.org/source/openssl-$SSL_VERSION.tar.gz.sha256 -O && \
    echo "$(cat openssl-$SSL_VERSION.tar.gz.sha256) openssl-$SSL_VERSION.tar.gz" | sha256sum --check && \
    tar -xzf openssl-$SSL_VERSION.tar.gz && \
    cd openssl-$SSL_VERSION && ./config && make depend && make install && \
    cd .. && rm -rf openssl-$SSL_VERSION*
ENV OPENSSL_LIB_DIR=/usr/local/ssl/lib \
    OPENSSL_INCLUDE_DIR=/usr/local/ssl/include \
    OPENSSL_STATIC=1

# Download, verify, and install Clang
# (See https://releases.llvm.org/download.html#8.0.0 for the latest available binaries)
RUN curl https://releases.llvm.org/8.0.0/hans-gpg-key.asc | gpg --import - && \
    curl https://releases.llvm.org/8.0.0/clang+llvm-8.0.0-x86_64-linux-gnu-ubuntu-18.04.tar.xz -O && \
    curl https://releases.llvm.org/8.0.0/clang+llvm-8.0.0-x86_64-linux-gnu-ubuntu-18.04.tar.xz.sig -O && \
    gpg --verify clang+llvm-8.0.0-x86_64-linux-gnu-ubuntu-18.04.tar.xz.sig \
                 clang+llvm-8.0.0-x86_64-linux-gnu-ubuntu-18.04.tar.xz && \
    tar -xvf clang+llvm-8.0.0-x86_64-linux-gnu-ubuntu-18.04.tar.xz && \
    mv clang+llvm-8.0.0-x86_64-linux-gnu-ubuntu-18.04 clang_8.0.0 && \
    rm -rf clang+llvm-8.0.0-x86_64-linux-gnu-ubuntu-18.04*


ENV PATH="/clang_8.0.0/bin:${PATH}" \
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

# The latest stable rust version
# TODO: Update manually.
ENV RUST_LATEST_VERSION=1.50.0

# Install Rust
RUN curl https://sh.rustup.rs -sSf | \
    sh -s -- --default-toolchain none
RUN rustup toolchain install $RUST_LATEST_VERSION && \
    rustup default $RUST_LATEST_VERSION
ENV PATH="$HOME/.cargo/bin:${PATH}"
RUN rustup component add rustfmt clippy --toolchain $RUST_LATEST_VERSION

# Install wasm-pack
RUN cargo install wasm-pack
RUN rustup component add rust-std --target wasm32-unknown-unknown

VOLUME /src
VOLUME $HOME/.cache
WORKDIR /src

# This is a workaround to chown the $HOME/.cache dir on startup as the builder user.
# This way the build cache can be reused between runs.
USER root
ADD entry.sh /entry.sh
ENTRYPOINT ["/entry.sh"]
