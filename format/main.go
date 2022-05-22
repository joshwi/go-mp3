package main

import (
	"log"
	"os"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/joshwi/go-utils/utils"
)

var (
	DIRECTORY = os.Getenv("DIRECTORY")
	USERNAME  = os.Getenv("NEO4J_USERNAME")
	PASSWORD  = os.Getenv("NEO4J_PASSWORD")
	HOST      = os.Getenv("NEO4J_SERVICE_HOST")
	PORT      = os.Getenv("NEO4J_SERVICE_PORT")
)

var regexp_1 = regexp.MustCompile(`[^a-zA-Z\d\/]+`)
var regexp_2 = regexp.MustCompile(`\_{2,}`)
var regexp_3 = regexp.MustCompile(`\.\_.*?\.\w{3}`)

func FormatPath(base string, location string) {
	new_path := strings.ReplaceAll(base+location, " ", "_")

	ext := path.Ext(new_path)

	new_path = strings.ReplaceAll(new_path, ext, "")

	new_path = regexp_1.ReplaceAllString(new_path, "_")

	new_path += ext

	new_path = regexp_2.ReplaceAllString(new_path, "_")

	_, err := os.Stat(new_path)
	if os.IsNotExist(err) {
		err = os.MkdirAll(path.Dir(new_path), 0755)
	}

	err = os.Rename(base+location, new_path)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {

	filetree, _ := utils.Scan(DIRECTORY)

	directories := []string{}
	files := []string{}

	start := time.Now()

	log.Printf("START: Removing invalid filenames")

	for _, item := range filetree {
		name := path.Base(item)
		match := regexp_3.FindString(name)
		if len(match) > 0 {
			info, err := os.Stat(DIRECTORY + item)
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
		err := os.RemoveAll(DIRECTORY + entry)
		if err != nil {
			log.Fatal(err)
		}
	}

	end := time.Now()
	elapsed := end.Sub(start)

	log.Printf("END: Removing invalid filenames")

	log.Printf("Time remove invalid filenames: %v", elapsed.Round(time.Second/1000))

	/*

		Reformat files and directories to no spaces

	*/

	filetree, _ = utils.Scan(DIRECTORY)

	directories = []string{}
	files = []string{}

	start = time.Now()

	log.Printf("START: Formatting directories")

	for _, item := range filetree {
		if strings.Contains(item, " ") {
			info, err := os.Stat(DIRECTORY + item)
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
		FormatPath(DIRECTORY, entry)
	}

	for _, entry := range directories {
		err := os.RemoveAll(DIRECTORY + entry)
		if err != nil {
			log.Fatal(err)
		}
	}

	end = time.Now()
	elapsed = end.Sub(start)

	log.Printf("END: Formatting directories")

	log.Printf("Time to move files & format directories: %v", elapsed.Round(time.Second/1000))

}
