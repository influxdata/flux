#!/bin/bash

# Run this script to prepare a Flux release
#
# This script is responsible for creating a commit that finalizes any changes
# to the source that need to be made before a release.
#
# The following optional dependencies are helpful if available.
#
# - `hub`, which will submit PRs for the update branches automatically if
#   available.

DIR=$(cd $(dirname ${BASH_SOURCE[0]}) && pwd)
cd $DIR

set -e

version=$(./gotool.sh github.com/influxdata/changelog nextver)

git checkout -b prep-release/$version

./etc/fixup_docs_version.sh $version
make generate

message="build(flux): prepare Flux release for $version"

git commit -am "$message"
git push

if ! command -v hub &> /dev/null
then
    echo "hub is not installed. Cannot open github PRs automatically."
    echo "Pull requests will have to be manually created."
    HAS_HUB=0
else
    HAS_HUB=1
fi

if [ $HAS_HUB -eq 1 ]
then
    hub pull-request -m "$message" -r influxdata/flux-team
fi
