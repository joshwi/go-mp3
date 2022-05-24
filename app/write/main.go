package main

import (
	"flag"
	"log"
	"os"

	"github.com/joshwi/go-mp3/app/tags"
	"github.com/joshwi/go-plugins/graphdb"
	"github.com/joshwi/go-utils/logger"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

var (
	directory = os.Getenv("DIRECTORY")
	username  = os.Getenv("NEO4J_USERNAME")
	password  = os.Getenv("NEO4J_PASSWORD")
	host      = os.Getenv("NEO4J_SERVICE_HOST")
	port      = os.Getenv("NEO4J_SERVICE_PORT")
	types     = map[string]string{
		"title":    "TIT2",
		"album":    "TALB",
		"artist":   "TPE1",
		"genre":    "TCON",
		"producer": "TCOM",
		"track":    "TRCK",
		"year":     "TYER",
		"comments": "COMM",
		"lyrics":   "USLT",
	}

	// Init flag values
	query   string
	logfile string
)

func init() {

	// Define flag arguments for the application
	flag.StringVar(&query, `q`, ``, `Run query to DB for input parameters. Default: <empty>`)
	flag.StringVar(&logfile, `l`, `./script.log`, `Location of script logfile. Default: ./script.log`)
	flag.Parse()

	// Initialize logfile at user given path. Default: ./collection.log
	logger.InitLog(logfile)

	logger.Logger.Info().Str("status", "start").Msg("WRITE TAGS")
}

func main() {

	// Create application session with Neo4j
	uri := "bolt://" + host + ":" + port
	driver := graphdb.Connect(uri, username, password)
	sessionConfig := neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite}
	session := driver.NewSession(sessionConfig)

	songs, _ := graphdb.RunCypher(session, query)

	for _, song := range songs {
		filepath := ""
		for _, item := range song {
			if item.Name == "filepath" {
				filepath = item.Value
			}
		}
		_, err := os.Stat(directory + filepath)
		if os.IsNotExist(err) {
			log.Println(song)
			log.Fatal(err)
		}
		err = tags.WriteTags(directory, filepath, song)
		if err != nil {
			log.Fatal(err)
		}
	}

	logger.Logger.Info().Str("status", "end").Msg("WRITE TAGS")
}
