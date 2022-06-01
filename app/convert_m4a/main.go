package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/joshwi/go-pkg/logger"
	"github.com/joshwi/go-pkg/utils"
)

var (
	DIRECTORY = os.Getenv("DIRECTORY")
	LOGFILE   = os.Getenv("LOGFILE")
)

func init() {

	// Define flag arguments for the application

	flag.Parse()

	// Initialize logfile at user given path. Default: ./collection.log
	logger.InitLog(LOGFILE)
}

var a0 = regexp.MustCompile(`[^a-zA-Z\d\/]+`)
var a1 = regexp.MustCompile(`\_{2,}`)
var a2 = regexp.MustCompile(`\.\_.*?\.\w{3}`)

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

	logger.Logger.Info().Str("status", "start").Msg("CONVERT M4A TO MP3")

	m4a := []string{}
	commands := []string{}

	filetree, _ := utils.Scan(DIRECTORY)

	start := time.Now()

	for _, item := range filetree {
		if strings.ToLower(filepath.Ext(item)) == ".m4a" {
			m4a = append(m4a, item)
		}
	}

	for _, entry := range m4a {
		rel := filepath.Clean(entry)
		base := strings.TrimSuffix(rel, filepath.Base(entry))
		name := strings.TrimSuffix(filepath.Base(entry), filepath.Ext(entry))
		new_path := fmt.Sprintf("%v%v%v.mp3", DIRECTORY, base, name)
		output := fmt.Sprintf(`-v 5 -y -i %v%v -acodec libmp3lame -ac 2 -ab 192k %v`, DIRECTORY, rel, new_path)
		_, err := os.Stat(new_path)
		if os.IsNotExist(err) {
			commands = append(commands, output)
		}
	}

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

	close(ports)
	close(results)

	end := time.Now()
	elapsed := end.Sub(start)

	logger.Logger.Info().Str("status", "end").Msg("CONVERT M4A TO MP3")

	logger.Logger.Info().Msg(fmt.Sprintf("Time to proccess %v files: %v", len(commands), elapsed.Round(time.Second/1000)))

	if len(commands) > 0 {
		avg := (int(elapsed.Milliseconds()) / len(commands))
		logger.Logger.Info().Msg(fmt.Sprintf("%v milliseconds per file", avg))
	}

}
