#!/bin/bash

set -e

stdlib=${FLUX_STDLIB_DIR-./stdlib}
fluxc=${FLUXC-fluxc}
fluxdoc=${FLUXDOC-fluxdoc}

# Lint the docs to ensure they are valid
$fluxdoc lint --dir ${stdlib} --stdlib-dir "${stdlib}"
