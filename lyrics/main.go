package main

import (
	"flag"
	"fmt"
	"html"
	"log"
	"os"
	"regexp"

	"github.com/joshwi/go-plugins/graphdb"
	"github.com/joshwi/go-utils/logger"
	"github.com/joshwi/go-utils/parser"
	"github.com/joshwi/go-utils/utils"
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
	flag.StringVar(&name, `c`, `pfr_team_season`, `Specify config. Default: pfr_team_season`)
	flag.StringVar(&filename, `f`, ``, `Location of parsing config file. Default: <empty>`)
	flag.StringVar(&logfile, `l`, `../script.log`, `Location of script logfile. Default: ../script.log`)
	flag.Parse()

	// Initialize logfile at user given path. Default: ./collection.log
	logger.InitLog(logfile)

	logger.Logger.Info().Str("config", name).Str("query", query).Str("status", "start").Msg("LYRIC AUDIT")
}

func main() {

	a0 := regexp.MustCompile(`[^a-zA-Z0-9\-\s]+`)
	a1 := regexp.MustCompile(`[\.]+`)
	a2 := regexp.MustCompile(`\s+`)

	b0 := regexp.MustCompile(`<h2.*h2>`)
	b1 := regexp.MustCompile(`<i>|<\/i>`)
	b2 := regexp.MustCompile(`<[^>].*?>`)
	b3 := regexp.MustCompile(`\n+`)
	b4 := regexp.MustCompile(`\[`)
	b5 := regexp.MustCompile(`\]`)
	b6 := regexp.MustCompile(`\n$|^\n`)

	base_url := "https://www.genius.com/"

	config, err := parser.Init(name, filename)
	if err != nil {
		log.Fatal(err)
	}

	// Create application session with Neo4j
	uri := "bolt://" + host + ":" + port
	driver := graphdb.Connect(uri, username, password)
	sessionConfig := neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite}
	session := driver.NewSession(sessionConfig)

	audit_list := map[string]string{}

	songs, _ := graphdb.RunCypher(session, query)

	for _, entry := range songs {
		var label string
		var artist string
		var title string
		for _, item := range entry {
			if item.Name == "label" {
				label = item.Value
			}
			if item.Name == "artist" {
				artist = item.Value
			}
			if item.Name == "title" {
				title = item.Value
			}
		}
		param := fmt.Sprintf("%v-%v-lyrics", artist, title)
		param = a0.ReplaceAllString(param, "")
		param = a1.ReplaceAllString(param, "")
		param = a2.ReplaceAllString(param, "-")
		audit_list[label] = base_url + param
	}

	for k, v := range audit_list {
		response, _ := utils.Get(v, map[string]string{})
		if response.Status == 200 {
			output := parser.Collect(response.Data, config.Parser)
			for _, entry := range output.Tags {
				if entry.Name == "lyrics" {
					lyrics := b0.ReplaceAllString(entry.Value, "")
					lyrics = b1.ReplaceAllString(lyrics, "")
					lyrics = b2.ReplaceAllString(lyrics, "\n")
					lyrics = b3.ReplaceAllString(lyrics, "\n")
					lyrics = b4.ReplaceAllString(lyrics, "\n[")
					lyrics = b5.ReplaceAllString(lyrics, "]\n")
					lyrics = b6.ReplaceAllString(lyrics, "")
					lyrics = html.UnescapeString(lyrics)
					err := graphdb.PutNode(session, "music", k, []utils.Tag{{Name: "lyrics", Value: lyrics}})
					if err != nil {
						log.Println(err)
					}
				}
			}
		}
	}

	logger.Logger.Info().Str("config", name).Str("query", query).Str("status", "end").Msg("LYRIC AUDIT")
}

// package main

// import (
// 	"fmt"
// 	"io/ioutil"
// 	"log"
// 	"net/http"
// 	"regexp"
// )

// func Search(url string, items []regexp.Regexp) []string {
// 	output := []string{}
// 	client := &http.Client{}
// 	resp, _ := client.Get(url)
// 	if resp.StatusCode == 200 {
// 		body, _ := ioutil.ReadAll(resp.Body)
// 		log.Println(url, resp.StatusCode)
// 		for _, entry := range items {
// 			results := entry.FindAllString(string(string(body)), -1)
// 			output = append(output, results...)
// 		}
// 	}
// 	return output
// }
// func main() {

// 	r0 := regexp.MustCompile(`^href=\"`)
// 	r1 := regexp.MustCompile(`\"$`)

// 	base_url := "https://genius.com/artists"

// 	search_urls := []string{"https://genius.com/artists", "https://genius.com/albums", "https://genius.com/artists"}

// 	patterns := []regexp.Regexp{}

// 	for _, entry := range search_urls {
// 		patterns = append(patterns, *regexp.MustCompile(fmt.Sprintf("href=\"%v.*?\"", entry)))
// 	}

// 	output := []string{}

// 	results := Search(base_url, patterns)

// 	for _, entry := range results {
// 		temp := r0.ReplaceAllString(entry, "")
// 		temp = r1.ReplaceAllString(temp, "")
// 		output = append(output, temp)
// 	}

// 	log.Println(output)

// }
