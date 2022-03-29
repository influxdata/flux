#!/bin/bash

HAS_FMT_ERR=0

function check_go_fmt () {
    echo "Checking Go source files..."
    # For every Go file in the project, excluding vendor...
    for file in $(go list -f '{{$dir := .Dir}}{{range .GoFiles}}{{printf "%s/%s\n" $dir .}}{{end}}' ./...); do
      # ... if file does not contain standard generated code comment (https://golang.org/s/generatedcode)...
      if ! grep -Exq '^// Code generated .* DO NOT EDIT\.$' $file; then
        FMT_OUT="$(gofmt -l -d -e "$file")" # gofmt exits 0 regardless of whether it's formatted.
        # ... and if gofmt had any output...
        if [[ -n "$FMT_OUT" ]]; then
          if [ "$HAS_FMT_ERR" -eq "0" ]; then
            # Only print this once.
            HAS_FMT_ERR=1
            echo 'Commit includes files that are not gofmt-ed' && \
            echo ''
          fi
          echo "$FMT_OUT" # Print output and continue, so developers don't fix one file at a time
        fi
       fi
    done
}

function check_rust_fmt() {
    echo "Checking Rust source files..."
    cd libflux || exit 1
    cargo fmt --all -- --check
    ret=$?
    cd ..
    if [[ $ret -ne 0 ]]; then
        echo 'Commit includes files that are not rustfmt-ed' && \
        echo ''
        HAS_FMT_ERR=1
    fi
}

function check_flux_fmt() {
    echo "Checking Flux source files..."
    env GO111MODULE=on go run -tags '' ./cmd/flux fmt -c stdlib
    ret=$?
    if [[ $ret -ne 0 ]]; then
        echo 'Commit includes flux files that are not fluxfmt-ed' && \
        echo ''
        HAS_FMT_ERR=1
    fi

}

check_go_fmt
check_rust_fmt
check_flux_fmt

## print at the end too... sometimes it is nice to see what to do at the end.
if [ "$HAS_FMT_ERR" -eq "1" ]; then
    echo 'Commit includes files that are not formatted' && \
    echo 'run "make fmt"' && \
    echo ''
fi
exit "$HAS_FMT_ERR"
