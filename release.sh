#!/bin/bash

DIR=$(cd $(dirname ${BASH_SOURCE[0]}) && pwd)
cd $DIR

set -e

# Ensure that the GITHUB_TOKEN is exposed in the environment.
if ! env | grep GITHUB_TOKEN= > /dev/null; then
  echo "GITHUB_TOKEN must be exported in the environment to perform a release." 2>&1
  exit 1
fi

export GO111MODULE=on

version=$(go run github.com/influxdata/changelog nextver)
git tag -s -m "Release $version" $version
go run github.com/goreleaser/goreleaser release --rm-dist --release-notes <(go run github.com/influxdata/changelog generate --version $version --commit-url https://github.com/influxdata/flux/commit)
