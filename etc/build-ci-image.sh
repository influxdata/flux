#!/bin/bash

set -e

IMAGE_NAME="quay.io/influxdb/flux-build"

docker pull "$IMAGE_NAME" || true
docker build \
  -t "$IMAGE_NAME" \
  --cache-from "$IMAGE_NAME" \
  --build-arg BUILDKIT_INLINE_CACHE=1 \
  -f Dockerfile_build .
