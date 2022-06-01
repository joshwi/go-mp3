package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html"
	"log"
	"os"
	"regexp"

	"github.com/joshwi/go-pkg/logger"
	"github.com/joshwi/go-pkg/parser"
	"github.com/joshwi/go-pkg/utils"
	"github.com/joshwi/go-svc/db"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

var (
	// Pull in env variables: username, password, uri
	USERNAME   = os.Getenv("NEO4J_USERNAME")
	PASSWORD   = os.Getenv("NEO4J_PASSWORD")
	HOST       = os.Getenv("NEO4J_SERVICE_HOST")
	PORT       = os.Getenv("NEO4J_SERVICE_PORT")
	LOGFILE    = os.Getenv("LOGFILE")
	config_dir = os.Getenv("CONFIG_DIR")

	// Init flag values
	query       string
	name        string
	NODE        string
	FIELD       string
	A1          string
	A2          string
	base_url    = "https://www.genius.com/"
	parser_file = fmt.Sprintf("%v/parser.json", config_dir)
	audit_file  = fmt.Sprintf("%v/audit.json", config_dir)
)

func init() {

	// Define flag arguments for the application
	flag.StringVar(&query, `query`, ``, `Run query to DB for input parameters. Default: <empty>`)
	flag.StringVar(&name, `config`, `genius_music_lyrics`, `Specify name of parser config. Default: genius_music_lyrics`)
	flag.StringVar(&NODE, `node`, `music`, `Specify DB node. Default: music`)
	flag.StringVar(&FIELD, `field`, `lyrics`, `Specify DB field to audit. Default: lyrics`)
	flag.StringVar(&A1, `a1`, `audit_lyrics_url`, `Specify name of url audit. Default: audit_lyrics_url`)
	flag.StringVar(&A2, `a2`, `audit_lyrics_html`, `Specify name of html audit. Default: audit_lyrics_html`)
	flag.Parse()

	// Initialize logfile at user given path. Default: ./collection.log
	logger.InitLog(LOGFILE)

	logger.Logger.Info().Str("config", name).Str("field", FIELD).Str("node", NODE).Str("query", query).Str("status", "start").Msg("LYRIC AUDIT")
}

func Init(file string, audit_one string, audit_two string) ([]utils.Match, []utils.Match, error) {

	// Open file with parsing configurations
	fileBytes, err := utils.Read(file)
	if err != nil {
		return []utils.Match{}, []utils.Match{}, err
	}

	// Unmarshall file into []Config struct
	var configurations map[string][]utils.Tag
	json.Unmarshal(fileBytes, &configurations)

	// Get config by name
	res_one := Compile(configurations[audit_one])
	res_two := Compile(configurations[audit_two])

	return res_one, res_two, nil
}

func Compile(input []utils.Tag) []utils.Match {
	tags := []utils.Match{}
	for _, n := range input {
		r := regexp.MustCompile(n.Value)
		exp := utils.Match{Name: n.Name, Value: *r}
		tags = append(tags, exp)
	}
	return tags
}

func Run(input string, commands []utils.Match) string {
	output := input
	for _, entry := range commands {
		output = entry.Value.ReplaceAllString(output, entry.Name)
	}
	return output
}

func req_worker(urls chan utils.Tag, songs chan utils.Tag, results chan error, config utils.Config, audit []utils.Match) {
	for item := range urls {
		response, _ := utils.Get(item.Value, map[string]string{})
		if response.Status == 200 {
			output := parser.Collect(response.Data, config.Parser)
			field_present := false
			for _, entry := range output.Tags {
				if entry.Name == FIELD {
					field_present = true
					result := Run(entry.Value, audit)
					result = html.UnescapeString(result)
					songs <- utils.Tag{Name: item.Name, Value: result}
					results <- nil
				}
			}
			if !field_present {
				songs <- utils.Tag{}
				results <- fmt.Errorf("Response missing field: %v", "lyrics")
			}
		} else {
			songs <- utils.Tag{}
			results <- fmt.Errorf("%v %v", response.Status, response.Error)
		}
	}
}

func db_worker(songs chan utils.Tag, results chan error, driver neo4j.Driver, field string) {
	for item := range songs {
		if item.Name != "" && item.Value != "" {
			sessionConfig := neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite}
			session := driver.NewSession(sessionConfig)
			err := db.PutNode(session, NODE, item.Name, []utils.Tag{{Name: field, Value: item.Value}})
			if err != nil {
				results <- err
			}
			results <- nil
		} else {
			results <- fmt.Errorf("Empty tag!")
		}
	}
}

func main() {

	config, err := parser.Init(name, parser_file)
	if err != nil {
		log.Fatal(err)
	}

	audit_in, audit_out, _ := Init(audit_file, A1, A2)

	// Create application session with Neo4j
	uri := "bolt://" + HOST + ":" + PORT
	driver := db.Connect(uri, USERNAME, PASSWORD)
	sessionConfig := neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite}
	session := driver.NewSession(sessionConfig)

	// Query DB to get audit search list
	nodes, _ := db.GetNode(session, NODE, query, 0, []string{"label", "artist", "title"})

	url_list := []utils.Tag{}
	for _, entry := range nodes {
		if entry["label"] != "" && entry["artist"] != "" && entry["title"] != "" {
			param := fmt.Sprintf("%v-%v-lyrics", entry["artist"], entry["title"])
			param = Run(param, audit_in)
			url_list = append(url_list, utils.Tag{Name: entry["label"], Value: base_url + param})
		}
	}

	// Create channels for data flow and error reporting
	urls := make(chan utils.Tag, 10)
	songs := make(chan utils.Tag, 10)
	req_err := make(chan error)
	db_err := make(chan error)

	// Input the url search list into channel
	go func() {
		for _, entry := range url_list {
			urls <- entry
		}
	}()

	// Run HTTP request worker to gather data
	for i := 0; i < cap(urls); i++ {
		go req_worker(urls, songs, req_err, config, audit_out)
	}

	songlist := make([]utils.Tag, len(url_list))

	for i := 0; i < cap(songlist); i++ {
		go db_worker(songs, db_err, driver, FIELD)
	}

	req_err_list := []error{}
	db_err_list := []error{}

	for range url_list {
		entry := <-req_err
		req_err_list = append(req_err_list, entry)
		item := <-db_err
		db_err_list = append(db_err_list, item)

	}

	close(urls)
	close(songs)
	close(req_err)
	close(db_err)

	logger.Logger.Info().Str("config", name).Str("field", FIELD).Str("node", NODE).Str("query", query).Str("status", "end").Msg("LYRIC AUDIT")
}
