#!/bin/bash

if ./builds/format; then
    ./builds/convert
else
    echo "Failed to run binary: ./builds/format"
fi