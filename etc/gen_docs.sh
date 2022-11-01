#!/bin/bash

set -e

stdlib=${FLUX_STDLIB_SRC-./stdlib}
fluxdoc=${FLUXDOC-fluxdoc}
gendir=${DOCS_GEN_DIR-./generated_docs}
full_docs="${gendir}/flux-docs-full.json"
short_docs="${gendir}/flux-docs-short.json"

# Generate docs JSON files
mkdir -p "${gendir}"
$fluxdoc dump --dir ${stdlib} --output ${full_docs} --nested
$fluxdoc dump --dir ${stdlib} --output ${short_docs} --short

