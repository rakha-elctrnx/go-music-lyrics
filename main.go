package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

func main() {
	song := "Billie_Eilish-Party_Favor"
	f, err := os.Open(song + ".mp3")
	if err != nil {
		log.Fatalf("Failed to open MP3 file: %v", err)
	}
	defer f.Close()

	streamer, format, err := mp3.Decode(f)
	if err != nil {
		log.Fatalf("Failed to decode MP3 file: %v", err)
	}
	defer streamer.Close()

	// Initialize the speaker
	err = speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
	if err != nil {
		log.Fatalf("Failed to initialize speaker: %v", err)
	}

	// Start audio playback in a goroutine
	done := make(chan struct{})
	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
		close(done)
	})))

	// Display lyrics from the LRC file
	err = displayLyrics(song + ".lrc")
	if err != nil {
		log.Fatalf("Failed to display lyrics: %v", err)
	}

	// Wait for playback to finish
	<-done
}

func displayLyrics(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open lyrics file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var re = regexp.MustCompile(`\[(\d+):(\d+\.\d+)\]`)
	var startTime time.Time
	var songOffset time.Duration

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "[") {
			matches := re.FindStringSubmatch(line)
			if len(matches) < 3 {
				continue
			}

			minutesStr := matches[1]
			secondsStr := matches[2]
			lyric := line[len(matches[0]):]

			minutes, err := strconv.Atoi(minutesStr)
			if err != nil {
				return fmt.Errorf("failed to parse minutes: %v", err)
			}

			seconds, err := strconv.ParseFloat(secondsStr, 64)
			if err != nil {
				return fmt.Errorf("failed to parse seconds: %v", err)
			}

			timestamp := float64(minutes*60) + seconds

			if startTime.IsZero() {
				startTime = time.Now()
			}

			currentTime := time.Since(startTime)
			waitTime := time.Duration(timestamp*1000)*time.Millisecond - currentTime + songOffset

			time.Sleep(waitTime)

			fmt.Println(lyric)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scanner error: %v", err)
	}

	return nil
}
