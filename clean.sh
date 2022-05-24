#!/bin/bash

if ./builds/format; then
    if ./builds/convert; then
        mkdir m4a
        find ./Music -name '*.m4a' -exec mv {}  ./m4a  \;
else
    echo "Failed to run binary: ./builds/format"
fi