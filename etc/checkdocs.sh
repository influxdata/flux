#!/bin/bash

set -e

stdlib=${FLUX_STDLIB_DIR-./stdlib}
fluxc=${FLUXC-fluxc}
fluxdoc=${FLUXDOC-fluxdoc}

dir=$(mktemp -d)
stdlib_compiled="${dir}/stdlib"

# Compile stdlib to a temporary dir
mkdir -p $stdlib_compiled
$fluxc stdlib --srcdir "${stdlib}" --outdir "${stdlib_compiled}"

# Lint the docs to ensure they are valid
$fluxdoc lint --dir ${stdlib} --stdlib-dir "${stdlib_compiled}"

rm -rf "$dir"
