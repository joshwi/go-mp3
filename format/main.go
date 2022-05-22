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
	directory = os.Getenv("DIRECTORY")
)

func FormatPath(base string, location string) {
	new_path := strings.ReplaceAll(base+location, " ", "_")

	ext := path.Ext(new_path)

	new_path = strings.ReplaceAll(new_path, ext, "")

	new_path = a0.ReplaceAllString(new_path, "_")

	new_path += ext

	new_path = a1.ReplaceAllString(new_path, "_")

	_, err := os.Stat(new_path)
	if os.IsNotExist(err) {
		err = os.MkdirAll(path.Dir(new_path), 0755)
	}

	err = os.Rename(base+location, new_path)
	if err != nil {
		log.Fatal(err)
	}
}

var a0 = regexp.MustCompile(`[^a-zA-Z\d\/]+`)
var a1 = regexp.MustCompile(`\_{2,}`)
var a2 = regexp.MustCompile(`\.\_.*?\.\w{3}`)

func main() {

	filetree, _ := utils.Scan(directory)

	directories := []string{}
	files := []string{}

	start := time.Now()

	log.Printf("START: Removing invalid filenames")

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

	log.Printf("END: Removing invalid filenames")

	log.Printf("Time remove invalid filenames: %v", elapsed.Round(time.Second/1000))

	/*

		Reformat files and directories to no spaces

	*/

	filetree, _ = utils.Scan(directory)

	directories = []string{}
	files = []string{}

	start = time.Now()

	log.Printf("START: Formatting directories")

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
		}
	}

	end = time.Now()
	elapsed = end.Sub(start)

	log.Printf("END: Formatting directories")

	log.Printf("Time to move files & format directories: %v", elapsed.Round(time.Second/1000))

}
