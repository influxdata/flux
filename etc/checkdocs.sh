#!/bin/bash

set -e

stdlib=${FLUX_STDLIB_DIR-./stdlib}
fluxdoc=${FLUXDOC-fluxdoc}

# Lint the docs to ensure they are valid
$fluxdoc lint --dir ${stdlib}
