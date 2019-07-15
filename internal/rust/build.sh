#!/bin/bash

set -e

# Make sure our working dir is the dir of the script
DIR=$(cd $(dirname ${BASH_SOURCE[0]}) && pwd)
cd $DIR

# Home dir of the docker user
SRC_DIR=/src

# Host user ids
uid=$(id -u)
gid=$(id -g)

imagename=flux-rust-builder-img

# Build new docker image
docker build \
    -f Dockerfile \
    -t $imagename \
    --build-arg UID=$uid \
    --build-arg GID=$gid \
    $DIR

# Run docker container to perform build
docker run \
    --rm \
    --name $imagename \
    -v "$DIR:$SRC_DIR" \
    -v "$DIR/.cache:/home/builder/.cache" \
    $imagename wasm-pack build --scope @influxdata/parser --dev
