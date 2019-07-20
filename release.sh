#!/bin/bash

DIR=$(cd $(dirname ${BASH_SOURCE[0]}) && pwd)
cd $DIR

set -e

# Ensure that the GITHUB_TOKEN is exposed in the environment.
if ! env | grep GITHUB_TOKEN= > /dev/null; then
  echo "GITHUB_TOKEN must be exported in the environment to perform a release." 2>&1
  exit 1
fi

# Run git fetch to ensure that the origin is updated.
git fetch

# Ensure that the maint branch is included in this release.
#if ! git merge-base --is-ancestor origin/maint HEAD; then
#  echo "maint branch has not been merged into $(git rev-parse --abbrev-ref HEAD)." 2>&1
#  exit 1
#fi

export GO111MODULE=on

version=$(go run ./internal/cmd/changelog nextver)
git tag -s -m "Release $version" $version
#git push origin "$version" "HEAD:maint"
go run github.com/goreleaser/goreleaser release --rm-dist --release-notes <(go run ./internal/cmd/changelog generate --version $version --commit-url https://github.com/influxdata/flux/commit)
