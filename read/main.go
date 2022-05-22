package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/bogem/id3v2"
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
	flag.StringVar(&logfile, `l`, `../script.log`, `Location of script logfile. Default: ../script.log`)
	flag.Parse()

	// Initialize logfile at user given path. Default: ./collection.log
	logger.InitLog(logfile)

	logger.Logger.Info().Str("status", "start").Msg("READ TAGS")
}

var a0 = regexp.MustCompile(`\s+`)
var a1 = regexp.MustCompile(`[^a-zA-Z\d]+`)
var a2 = regexp.MustCompile(`\_{2,}`)

func WriteTags(dir string, filename string, properties []utils.Tag) error {
	tag, err := id3v2.Open(dir+filename, id3v2.Options{Parse: true})
	if err != nil {
		logger.Logger.Error().Str("filename", filename).Str("status", "failed").Err(err).Msg("WriteTags")
	}
	defer tag.Close()

	var track, track_total string

	for _, entry := range properties {
		switch entry.Name {
		case "filepath":
			break
		case "comments":
			comment := id3v2.CommentFrame{
				Encoding:    id3v2.EncodingUTF8,
				Language:    "eng",
				Description: "comment",
				Text:        entry.Value,
			}
			tag.AddCommentFrame(comment)
		case "lyrics":
			lyrics := id3v2.UnsynchronisedLyricsFrame{
				Encoding: id3v2.EncodingUTF8,
				Language: "eng",
				Lyrics:   entry.Value,
			}
			tag.AddUnsynchronisedLyricsFrame(lyrics)
		case "track":
			track = entry.Value
		case "track_total":
			track_total = entry.Value
			break
		default:
			tag.AddTextFrame(tag.CommonID(types[entry.Name]), tag.DefaultEncoding(), entry.Value)
		}

	}

	if len(track) > 0 && len(track_total) > 0 {
		tag.AddTextFrame(tag.CommonID("TRCK"), tag.DefaultEncoding(), fmt.Sprintf("%v/%v", track, track_total))
	}

	err = tag.Save()
	if err != nil {
		logger.Logger.Error().Str("filename", filename).Str("status", "failed").Err(err).Msg("WriteTags")
		log.Fatal(err)
	}

	logger.Logger.Info().Str("filename", filename).Str("status", "success").Msg("WriteTags")

	return nil
}

func ReadTags(dir string, filename string) ([]utils.Tag, string, error) {
	// Open tags from file
	tag, err := id3v2.Open(dir+filename, id3v2.Options{Parse: true})
	if err != nil {
		logger.Logger.Error().Str("filename", filename).Str("status", "failed").Err(err).Msg("ReadTags")
		return []utils.Tag{}, "", err
	}
	defer tag.Close()

	// Parse comment frame
	commFrames := tag.GetLastFrame(tag.CommonID("COMM"))
	comment, _ := commFrames.(id3v2.CommentFrame)

	// Parse lyrics frame
	lyrics := tag.GetLastFrame(tag.CommonID("USLT"))
	uslf, _ := lyrics.(id3v2.UnsynchronisedLyricsFrame)

	// Parse track tag for track # and total tracks in album
	tracks := strings.Split(tag.GetTextFrame("TRCK").Text, "/")
	if len(tracks) < 2 {
		tracks = append(tracks, "")
	}

	// Format m4a tags into utils tag structure
	tags := []utils.Tag{
		{Name: "title", Value: tag.GetTextFrame("TIT2").Text},
		{Name: "artist", Value: tag.GetTextFrame("TPE1").Text},
		{Name: "album", Value: tag.GetTextFrame("TALB").Text},
		{Name: "genre", Value: tag.GetTextFrame("TCON").Text},
		{Name: "producer", Value: tag.GetTextFrame("TCOM").Text},
		{Name: "year", Value: tag.GetTextFrame("TYER").Text},
		{Name: "track", Value: tracks[0]},
		{Name: "track_total", Value: tracks[1]},
		{Name: "comments", Value: comment.Text},
		{Name: "lyrics", Value: uslf.Lyrics},
		{Name: "filepath", Value: filename},
	}

	// Build unique label for DB entry
	label := tag.Artist() + "_" + tag.Album() + "_" + tracks[0]
	label = a0.ReplaceAllString(label, "_")
	label = a1.ReplaceAllString(label, "_")
	label = a2.ReplaceAllString(label, "_")

	logger.Logger.Info().Str("filename", filename).Str("status", "success").Msg("ReadTags")

	return tags, label, nil
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
		tags, label, _ := ReadTags(directory, entry)
		graphdb.PostNode(session, "music", label, tags)
		results <- 1
	}
	session.Close()
}
