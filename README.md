# go-mp3: Library for managing and auditng music files

## Table of contents
* [Requirements](#requirements)
* [Installation](#installation)
* [Build](#setup)
* [Run](#run)
* [Debug](#debug)

## Requirements

1. Path to directory containing music library (example:`/home/go-mp3/Music`)

- Only mp3 and m4a files will be considered

2. Access to a Neo4j DB instance

3. ffmpeg installed on your machine for vm

## Installation

Clone the go-mp3 Repository:

```
git clone https://github.com/joshwi/go-mp3.git
```

## Build

Run the build script:

```
bash ./scripts/build.sh
```

## Run

1. Create & open .env in editor: 
```
nano .env
```
2. Add your env variables:
```
DIRECTORY=/home/go-mp3/Music
NEO4J_USERNAME=neo4j
NEO4J_PASSWORD=guest
NEO4J_SERVICE_HOST=example.com
NEO4J_SERVICE_PORT=7687
```
3. Source .env file
```
source .env
```
4. Run the formatter script to fix file paths and convert m4a files to mp3:
```
bash ./scripts/format
```
5. Run audit script to add lyrics to files:
```
bash ./scripts/audit "MF DOOM"
```

## Debug

If some songs are not being fetched properly from genius, you can try to manually insert values:
```
./app/builds/lyrics -q "MATCH (n:music) WHERE n.artist='MF DOOM' AND n.album='Madvillainy' AND n.lyrics='' RETURN n.label as label, 'Madvillain' as artist, n.title as title" -f "./config/genius.json" -n "genius_song_lyrics"
./app/builds/write -q "MATCH (n:music) WHERE n.artist='MF DOOM' AND n.lyrics<>'' RETURN n.filepath as filepath, n.lyrics as lyrics"
```
