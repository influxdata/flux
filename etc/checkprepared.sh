#!/bin/bash

# This script's purpose is to check that any needed changes to the source
# have been made prerelease.
#
# For example we check that all instances of 'introduced: NEXT' in the docs
# have been replaced with a version instead.

# Make sure we are at the repo root
DIR=$(cd $(dirname ${BASH_SOURCE[0]}) && pwd)/..
cd $DIR

EXIT=0
while read f
do
    # Check for any 'introduced: NEXT' or 'deprecated: NEXT' comments that still exist
    grep '^//[[:space:]]*\(introduced\|deprecated\):[[:space:]]\+NEXT[[:space:]]*$' $f > /dev/null
    ret=$?
    if [ $ret -eq 0 ]
    then
        EXIT=1
        echo "$f contains 'introduced: NEXT' or 'deprecated: NEXT'"
    fi
done < <(find ./stdlib -name '*.flux')

if [ $EXIT -ne 0 ]
then
    echo "Flux not prepared for release. Run ./prep_release.sh to start the release preparation process."
fi
exit $EXIT
