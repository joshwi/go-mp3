#!/bin/bash

if ./app/builds/format_files; then
    if ./app/builds/convert_m4a; then
        ./app/builds/move_m4a
    else
        echo "Failed to run binary: ./app/builds/convert_m4a"
    fi
else
    echo "Failed to run binary: ./app/builds/format_files"
fi