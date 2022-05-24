#!/bin/bash

if ./app/builds/format_files; then
    ./app/builds/convert_m4a
    # if ./builds/convert_m4a; then
    #     mkdir m4a
    #     find ./Music -name '*.m4a' -exec mv {}  ./m4a  \;
else
    echo "Failed to run binary: ./app/builds/format_files"
fi