#!/bin/bash

if [ $# -eq 0 ]; then
    ./app/builds/read_tags
elif [ $# -eq 1 ]; then
    query="MATCH (n:music) WHERE n.artist=~'$query' RETURN n.filepath as filepath" 
    ./app/builds/read_tags -q="$query"
else
  echo 1>&2 "$0: too many arguments"
  exit 2
fi