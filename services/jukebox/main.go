package main

import (
	"compress/gzip"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
)

type Track struct {
	ID       string
	Duration time.Duration
}

var tracks []Track
var playlistStart = time.Unix(1467385200, 0)

func init() {
	resp, err := http.Get("https://storage.googleapis.com/machine-executive.appspot.com/random_tracks.txt.gz")
	if err != nil {
		log.Fatal(err)
	}
	gz, err := gzip.NewReader(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	reader := csv.NewReader(gz)
	reader.Comma = '\t'

	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		durationMS, err := strconv.Atoi(row[1])
		if err != nil {
			log.Fatal(err)
		}
		duration := time.Duration(durationMS) * time.Millisecond

		track := Track{
			ID:       string(row[0]),
			Duration: duration,
		}
		tracks = append(tracks, track)
	}
}

func GetTrackOffset(now time.Time) int {
	sinceStart := now.Sub(playlistStart)

	var soFar time.Duration
	for i, track := range tracks {
		soFar += track.Duration

		if soFar > sinceStart {
			return i
		}
	}

	return len(tracks) - 1
}

func HandleNext(w http.ResponseWriter, r *http.Request) {
	i := GetTrackOffset(time.Now())
	trackID := tracks[i].ID
	url := fmt.Sprintf("spotify:track:%s", trackID)

	http.Redirect(w, r, url, http.StatusFound)
}

func main() {
	var (
		port = flag.Int("port", 8080, "")
	)
	flag.Parse()

	http.HandleFunc("/next", HandleNext)

	addr := fmt.Sprint(":", *port)
	fmt.Println("listening at", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
