#!/bin/bash

# Make sure we are at the repo root
DIR=$(cd $(dirname ${BASH_SOURCE[0]}) && pwd)
cd $DIR/..

# Version is the first and only arg, remove the 'v' prefix.
version=${1//v/}

while read f
do
    # Replace any 'introduced: NEXT' or 'deprecated: NEXT' comment with the actual version
    perl -pi -e "s,^//\s*(introduced|deprecated):\s+NEXT\s*$,// \1: $version\n,g" $f
done < <(find ./stdlib -name '*.flux')
