package main

import (
	"flag"
	"os"

	"github.com/joshwi/go-mp3/app/audit"
	"github.com/joshwi/go-pkg/logger"
	"github.com/joshwi/go-pkg/utils"
	"github.com/joshwi/go-svc/db"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

var (
	// Pull in env variables: username, password, uri
	username = os.Getenv("NEO4J_USERNAME")
	password = os.Getenv("NEO4J_PASSWORD")
	host     = os.Getenv("NEO4J_SERVICE_HOST")
	port     = os.Getenv("NEO4J_SERVICE_PORT")

	// Init flag values
	query    string
	name     string
	filename string
	logfile  string
)

func init() {

	// Define flag arguments for the application
	flag.StringVar(&query, `q`, ``, `Run query to DB for input parameters. Default: <empty>`)
	flag.StringVar(&name, `n`, ``, `Specify field name for audit. Default: <empty>`)
	flag.StringVar(&logfile, `l`, `./run.log`, `Location of script logfile. Default: ./run.log`)
	flag.Parse()

	// Initialize logfile at user given path. Default: ./collection.log
	logger.InitLog(logfile)

	logger.Logger.Info().Str("filename", filename).Str("config", name).Str("query", query).Str("status", "start").Msg("LYRIC AUDIT")
}

func main() {

	config := audit.CONFIG[name]
	commands := audit.Compile(config)

	// Create application session with Neo4j
	uri := "bolt://" + host + ":" + port
	driver := db.Connect(uri, username, password)
	sessionConfig := neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite}
	session := driver.NewSession(sessionConfig)

	nodes := map[string]string{}

	songs, _ := db.RunCypher(session, query)

	for _, entry := range songs {
		var label, field string
		for _, item := range entry {
			if item.Name == "label" {
				label = item.Value
			}
			if item.Name == name {
				field = item.Value
			}
		}
		if len(label) > 0 && len(field) > 0 {
			nodes[label] = field
		}
	}

	results := map[string]string{}

	for k, v := range nodes {
		v = audit.Run(v, commands)
		results[k] = v
	}

	for k, v := range results {
		db.PutNode(session, "music", k, []utils.Tag{{Name: name, Value: v}})
	}

	logger.Logger.Info().Str("filename", filename).Str("config", name).Str("query", query).Str("status", "end").Msg("LYRIC AUDIT")
}
