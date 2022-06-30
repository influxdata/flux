#!/bin/bash

set -e

if [ "$CIRCLE_BRANCH" != "master" ]; then
  echo "Skipping publish step on non-master branch."
  exit 0
fi

docker login "-u=${QUAY_CD_USER}" --password-stdin quay.io > /dev/null 2>&1 <<< "$QUAY_CD_PASSWORD"
docker push quay.io/influxdb/flux-build
