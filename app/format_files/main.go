package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/joshwi/go-utils/logger"
	"github.com/joshwi/go-utils/utils"
)

var (
	directory = os.Getenv("DIRECTORY")
	logfile   string
)

func init() {

	// Define flag arguments for the application
	flag.StringVar(&logfile, `l`, `./run.log`, `Location of script logfile. Default: ./run.log`)
	flag.Parse()

	// Initialize logfile at user given path. Default: ./collection.log
	logger.InitLog(logfile)
}

func FormatPath(base string, location string) {

	new_path := strings.ReplaceAll(location, " ", "_")

	ext := path.Ext(new_path)

	new_path = strings.ReplaceAll(new_path, ext, "")

	new_path = a0.ReplaceAllString(new_path, "_")

	new_path += ext

	new_path = a1.ReplaceAllString(new_path, "_")

	abs_path := base + new_path

	_, err := os.Stat(abs_path)
	if os.IsNotExist(err) {
		err = os.MkdirAll(path.Dir(abs_path), 0755)
	}

	err = os.Rename(base+location, abs_path)
	if err != nil {
		log.Fatal(err)
	}
}

var a0 = regexp.MustCompile(`[^a-zA-Z\d\/]+`)
var a1 = regexp.MustCompile(`\_{2,}`)
var a2 = regexp.MustCompile(`\.\_.*?\.\w{3}`)

func main() {

	logger.Logger.Info().Str("status", "start").Msg("AUDITING FILENAMES")

	filetree, _ := utils.Scan(directory)

	directories := []string{}
	files := []string{}

	start := time.Now()

	for _, item := range filetree {
		name := path.Base(item)
		match := a2.FindString(name)
		if len(match) > 0 {
			info, err := os.Stat(directory + item)
			if os.IsNotExist(err) {
				log.Println(item)
				log.Fatal("File does not exist.")
			}
			if info.IsDir() {
				directories = append(directories, item)
			} else {
				files = append(files, item)
			}
		}
	}

	for _, entry := range files {
		err := os.RemoveAll(directory + entry)
		if err != nil {
			log.Fatal(err)
		}
	}

	end := time.Now()
	elapsed := end.Sub(start)

	logger.Logger.Info().Str("status", "end").Msg("AUDITING FILENAMES")

	logger.Logger.Info().Msg(fmt.Sprintf("Auditing filenames completed in: %v", elapsed.Round(time.Second/1000)))

	/*

		Reformat files and directories to no spaces

	*/

	logger.Logger.Info().Str("status", "start").Msg("FORMAT DIR")

	filetree, _ = utils.Scan(directory)

	directories = []string{}
	files = []string{}

	start = time.Now()

	for _, item := range filetree {
		if strings.Contains(item, " ") {
			info, err := os.Stat(directory + item)
			if os.IsNotExist(err) {
				log.Println(item)
				log.Fatal("File does not exist.")
			}
			if info.IsDir() {
				directories = append(directories, item)
			} else {
				files = append(files, item)
			}
		}
	}

	for _, entry := range files {
		FormatPath(directory, entry)
	}

	for _, entry := range directories {
		err := os.RemoveAll(directory + entry)
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
	}

	end = time.Now()
	elapsed = end.Sub(start)

	logger.Logger.Info().Str("status", "end").Msg("FORMAT DIR")

	logger.Logger.Info().Msg(fmt.Sprintf("Formatting directories completed in: %v", elapsed.Round(time.Second/1000)))

}
