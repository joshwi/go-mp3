package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/joshwi/go-pkg/logger"
	"github.com/joshwi/go-pkg/utils"
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

func main() {

	/*

		Reformat all m4a files to mp3

	*/

	logger.Logger.Info().Str("status", "start").Msg("MOVE M4A FILES")

	m4a := []string{}

	filetree, _ := utils.Scan(directory)

	for _, item := range filetree {
		if strings.ToLower(filepath.Ext(item)) == ".m4a" {
			m4a = append(m4a, item)
		}
	}

	for _, entry := range m4a {
		new_file := fmt.Sprintf("%v/m4a%v", directory, entry)
		_, err := os.Stat(new_file)
		if os.IsNotExist(err) {
			err = os.MkdirAll(path.Dir(new_file), 0755)
		}
		os.Rename(directory+entry, new_file)
	}

	logger.Logger.Info().Str("status", "end").Msg("MOVE M4A FILES")

}
