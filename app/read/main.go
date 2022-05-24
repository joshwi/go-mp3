package main

import (
	"flag"
	"log"
	"os"
	"strings"
	"time"

	"github.com/joshwi/go-mp3/app/tags"
	"github.com/joshwi/go-plugins/graphdb"
	"github.com/joshwi/go-utils/logger"
	"github.com/joshwi/go-utils/utils"
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

	logger.Logger.Info().Str("status", "start").Msg("READ TAGS")
}

func main() {

	// Create application session with Neo4j
	uri := "bolt://" + host + ":" + port
	driver := graphdb.Connect(uri, username, password)

	filetree, _ := utils.Scan(directory)

	files := []string{}

	for _, item := range filetree {
		if strings.Contains(item, ".mp3") {
			info, err := os.Stat(directory + item)
			if os.IsNotExist(err) {
				log.Println(item)
				log.Fatal("File does not exist.")
			}
			if !info.IsDir() {
				files = append(files, item)
			}
		}
	}

	start := time.Now()

	queue := make(chan string, 100)
	results := make(chan int)

	for i := 0; i < cap(queue); i++ {
		go worker(driver, queue, results)
	}

	go func() {
		for _, entry := range files {
			queue <- entry
		}
	}()

	pass := []int{}

	for i := range files {
		success := <-results
		if success == 1 {
			pass = append(pass, i)
		}
	}

	close(queue)
	close(results)

	end := time.Now()
	elapsed := end.Sub(start)

	log.Printf("Time to proccess %v files: %v", len(pass), elapsed.Round(time.Second/1000))

	avg := (int(elapsed.Milliseconds()) / len(pass))

	log.Printf("%v milliseconds per file", avg)

	logger.Logger.Info().Str("status", "end").Msg("READ TAGS")
}

func worker(driver neo4j.Driver, queue chan string, results chan int) {
	sessionConfig := neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite}
	session := driver.NewSession(sessionConfig)
	for entry := range queue {
		tags, label, _ := tags.ReadTags(directory, entry)
		graphdb.PutNode(session, "music", label, tags)
		results <- 1
	}
	session.Close()
}
