#!/bin/bash

set -e

stdlib=${FLUX_STDLIB_SRC-./stdlib}
fluxc=${FLUXC-fluxc}
fluxdoc=${FLUXDOC-fluxdoc}
gendir=${DOCS_GEN_DIR-./generated_docs}
full_docs="${gendir}/flux-docs-full.json"
short_docs="${gendir}/flux-docs-short.json"

dir=$(mktemp -d)
stdlib_compiled="${dir}/stdlib"

# Compile stdlib to a temporary dir
mkdir -p $stdlib_compiled
$fluxc stdlib --srcdir "${stdlib}" --outdir "${stdlib_compiled}"

# Generate docs JSON files
mkdir -p "${gendir}"
$fluxdoc dump --dir ${stdlib} --stdlib-dir "${stdlib_compiled}" --allow-exceptions --output ${full_docs} --nested
$fluxdoc dump --dir ${stdlib} --stdlib-dir "${stdlib_compiled}" --allow-exceptions --output ${short_docs} --short

rm -rf "${dir}"

