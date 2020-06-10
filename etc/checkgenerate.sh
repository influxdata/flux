#!/bin/bash

set -e

make cleangenerate
make generate

status=$(git status --porcelain)
if [ -n "$status" ]; then
  >&2 echo "generated code is not accurate, please run make generate"
  >&2 echo "ragel version is: $(ragel --version)"
  >&2 echo -e "Files changed:\n$status"
  >&2 git diff
  exit 1
fi
