#!/bin/bash

if ./app/builds/format; then
    ./app/builds/convert
    # if ./builds/convert; then
    #     mkdir m4a
    #     find ./Music -name '*.m4a' -exec mv {}  ./m4a  \;
else
    echo "Failed to run binary: ./app/builds/format"
fi