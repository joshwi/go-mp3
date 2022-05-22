package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/joshwi/go-utils/utils"
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

func worker(ports chan string, results chan int) {
	for item := range ports {
		args := strings.Split(item, " ")
		cmd := exec.Command("ffmpeg", args...)
		err := cmd.Run()
		if err != nil {
			log.Println(cmd, err)
			results <- 0
			continue
		}
		results <- 1
	}
}

func main() {

	/*

		Reformat all m4a files to mp3

	*/

	directory := "/Users/josh/Desktop/m4a/Music"

	m4a := []string{}
	commands := []string{}

	filetree, _ := utils.Scan(directory)

	start := time.Now()

	log.Printf("START: Formatting .m4a to .mp3")

	for _, item := range filetree {
		if strings.ToLower(filepath.Ext(item)) == ".m4a" {
			m4a = append(m4a, item)
		}
	}

	num_files := len(m4a)

	for _, entry := range m4a {
		rel := filepath.Clean(entry)
		base := strings.TrimSuffix(rel, filepath.Base(entry))
		name := strings.TrimSuffix(filepath.Base(entry), filepath.Ext(entry))
		output := fmt.Sprintf(`-v 5 -y -i %v%v -acodec libmp3lame -ac 2 -ab 192k %v%v%v.mp3`, directory, rel, directory, base, name)
		commands = append(commands, output)
	}

	// var wg sync.WaitGroup

	// for _, command := range commands {
	// 	wg.Add(1)
	// 	go func(command string) {
	// 		defer wg.Done()
	// 		args := strings.Split(command, " ")
	// 		cmd := exec.Command("ffmpeg", args...)
	// 		err := cmd.Run()
	// 		if err != nil {
	// 			log.Println(err)
	// 		}
	// 	}(command)
	// }

	// wg.Wait()

	ports := make(chan string, 100)
	results := make(chan int)

	for i := 0; i < cap(ports); i++ {
		go worker(ports, results)
	}

	go func() {
		for _, command := range commands {
			ports <- command
		}
	}()

	pass := []int{}

	for i := range commands {
		success := <-results
		if success == 1 {
			pass = append(pass, i)
		}
	}

	log.Println(pass)

	close(ports)
	close(results)

	end := time.Now()
	elapsed := end.Sub(start)

	log.Printf("END: Formatting .m4a to .mp3")

	log.Printf("Time to proccess %v files: %v", num_files, elapsed.Round(time.Second/1000))

	avg := (int(elapsed.Milliseconds()) / num_files)

	log.Printf("%v milliseconds per file", avg)

}
