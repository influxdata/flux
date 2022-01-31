#!/bin/bash

set -e

check_library_hash() {
  md5sum libflux/target/*/libflux.a
}

# Ensure the build tree is clean (it should be) and then build
# the libflux library.
make clean
make libflux

check_library_hash > md5sum.txt
trap "rm -f md5sum.txt" EXIT

make clean
make libflux

if ! diff -u md5sum.txt <(check_library_hash); then
  echo "Build produced two different hashes when built twice" 2>&1
  exit 1
fi
