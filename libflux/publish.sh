#!/bin/bash

set -e

# Make sure our working dir is the dir of the script
DIR=$(cd $(dirname ${BASH_SOURCE[0]}) && pwd)
cd $DIR

# Clean out any old build artifacts
rm -rf parser/pkg

# Build the WASM package
./build.sh

# Publish the package
cd parser/pkg
yarn publish --non-interactive --no-git-tag-version
