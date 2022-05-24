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

./app/builds/lyrics -q "MATCH (n:music) WHERE n.artist=~'$query' AND n.lyrics='' RETURN n.label as label, n.artist as artist, n.title as title" -f "./config/genius.json" -n "genius_song_lyrics"
./app/builds/write -q "MATCH (n:music) WHERE n.artist=~'$query' AND n.lyrics<>'' RETURN n.filepath as filepath, n.lyrics as lyrics"