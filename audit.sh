#!/bin/bash

# n1=$(wc -l < go.mod)
# echo "$n1"

if [ $# -lt 1 ]; then
  echo 1>&2 "$0: not enough arguments"
  exit 2
elif [ $# -gt 1 ]; then
  echo 1>&2 "$0: too many arguments"
  exit 2
fi

query=$1

