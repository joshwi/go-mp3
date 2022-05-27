package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"regexp"

	"github.com/joshwi/go-pkg/logger"
	"github.com/joshwi/go-pkg/utils"
	"github.com/joshwi/go-svc/db"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

var (
	// Pull in env variables: username, password, uri
	directory = os.Getenv("DIRECTORY")
	username  = os.Getenv("NEO4J_USERNAME")
	password  = os.Getenv("NEO4J_PASSWORD")
	host      = os.Getenv("NEO4J_SERVICE_HOST")
	port      = os.Getenv("NEO4J_SERVICE_PORT")

	// Init flag values
	query    string
	name     string
	filename string
	logfile  string
)

func init() {

	// Define flag arguments for the application
	flag.StringVar(&query, `q`, ``, `Run query to DB for input parameters. Default: <empty>`)
	flag.StringVar(&logfile, `l`, `./run.log`, `Location of script logfile. Default: ./run.log`)
	flag.Parse()

	// Initialize logfile at user given path. Default: ./collection.log
	logger.InitLog(logfile)

	logger.Logger.Info().Str("filename", filename).Str("config", name).Str("query", query).Str("status", "start").Msg("LYRIC AUDIT")
}

func main() {

	var a0 = regexp.MustCompile(`[^a-zA-Z\d\/]+`)
	var a1 = regexp.MustCompile(`\_{2,}`)

	// Create application session with Neo4j
	uri := "bolt://" + host + ":" + port
	driver := db.Connect(uri, username, password)
	sessionConfig := neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite}
	session := driver.NewSession(sessionConfig)

	songs, _ := db.RunCypher(session, query)

	for _, entry := range songs {

		var label string
		var filepath string
		var artist string
		var album string
		var track string
		var title string
		for _, item := range entry {
			if item.Name == "label" {
				label = item.Value
			}
			if item.Name == "filepath" {
				filepath = item.Value
			}
			if item.Name == "artist" {
				artist = item.Value
			}
			if item.Name == "album" {
				album = item.Value
			}
			if item.Name == "track" {
				track = item.Value
			}
			if item.Name == "title" {
				title = item.Value
			}
		}

		expected_filepath := fmt.Sprintf("/%v/%v/%v_%v", artist, album, track, title)
		expected_filepath = a0.ReplaceAllString(expected_filepath, "_")
		expected_filepath = a1.ReplaceAllString(expected_filepath, "_")
		expected_filepath += ".mp3"

		_, err := os.Stat(expected_filepath)
		if os.IsNotExist(err) {
			err = os.MkdirAll(path.Dir(expected_filepath), 0755)
			os.Rename(directory+filepath, directory+expected_filepath)
			db.PutNode(session, "music", label, []utils.Tag{{Name: "filepath", Value: expected_filepath}})
		}

	}

}
