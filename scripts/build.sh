DIR="./app/builds"
if [ ! -d "$DIR" ]; then
   echo "Creating directory: $DIR"
    mkdir ./app/builds
fi

go build -o ./app/builds/format ./app/format
go build -o ./app/builds/convert ./app/convert
go build -o ./app/builds/read ./app/read
go build -o ./app/builds/lyrics ./app/lyrics
go build -o ./app/builds/write ./app/write