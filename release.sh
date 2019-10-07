#!/bin/bash

DIR=$(cd $(dirname ${BASH_SOURCE[0]}) && pwd)
cd $DIR

set -e

version=$(go run github.com/influxdata/changelog nextver)
git tag -s -m "Release $version" $version
git push origin "$version"
