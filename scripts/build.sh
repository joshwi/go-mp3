DIR="./app/builds"
if [ ! -d "$DIR" ]; then
   echo "Creating directory: $DIR"
    mkdir ./app/builds
fi

go build -o ./app/builds/format_files ./app/format_files
go build -o ./app/builds/convert_m4a ./app/convert_m4a
go build -o ./app/builds/read_tags ./app/read_tags
go build -o ./app/builds/audit_lyrics ./app/audit_lyrics
go build -o ./app/builds/write_tags ./app/write_tags