#!/bin/bash

set -e

function tmpl() {
    ../../../gotool.sh github.com/benbjohnson/tmpl "$@"
}

now=$(date +%s)
tmpl -data="\"${now}\"" -o hack.gen.go hack.gen.go.tmpl
