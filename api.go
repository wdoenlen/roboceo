package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/vmihailenco/msgpack.v2"
)

type API struct {
	scraper *Scraper
	db      *DB
}

func (a *API) HandleScrape(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	req := &ScrapeRequest{
		Location: r.FormValue("location"),
		Tomorrow: r.FormValue("tomorrow") == "true",
	}
	go func() {
		if err := a.scraper.Scrape(req); err != nil {
			log.Printf("[error]: %v", err)
		}
	}()
	w.WriteHeader(http.StatusAccepted)
	fmt.Fprintln(w, "scrape started")
}

type EventMessage struct {
	ID          string    `msgpack:"id" json:"id"`
	Name        string    `msgpack:"name" json:"name"`
	Description string    `msgpack:"description" json:"description"`
	Latitude    float64   `msgpack:"latitude" json:"latitude"`
	Longitude   float64   `msgpack:"longitude" json:"longitude"`
	StartTime   time.Time `msgpack:"start_time" json:"start_time"`
	EndTime     time.Time `msgpack:"end_time" json:"end_time"`
}

func truncate(str string, n int) string {
	var runeCount int
	for i := range str {
		runeCount++
		if runeCount > n {
			return str[:i] + "…"
		}
	}
	return str
}

func (a *API) getMessages(req GetEventsRequest) ([]EventMessage, error) {
	events, err := a.db.GetEvents(req)
	if err != nil {
		return nil, err
	}

	var eventMsgs []EventMessage
	for _, event := range events {
		var e struct {
			ID          string `json:"id"`
			Name        string `json:"name"`
			Description string `json:"description"`
			StartTime   FBTime `json:"start_time"`
			EndTime     FBTime `json:"end_time"`
			Place       struct {
				Location struct {
					Latitude  float64 `json:"latitude"`
					Longitude float64 `json:"longitude"`
				} `json:"location"`
			} `json:"place"`
		}
		if err := json.Unmarshal(event, &e); err != nil {
			return nil, err
		}

		eventMsg := EventMessage{
			ID:          e.ID,
			Name:        truncate(e.Name, 50),
			Description: truncate(e.Description, 100),
			Latitude:    e.Place.Location.Latitude,
			Longitude:   e.Place.Location.Longitude,
			StartTime:   time.Time(e.StartTime),
			EndTime:     time.Time(e.EndTime),
		}
		eventMsgs = append(eventMsgs, eventMsg)
	}

	return eventMsgs, nil
}

// slimmed down version
func (a *API) HandleEventsMobile(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	var req GetEventsRequest

	start, _ := time.Parse(time.RFC3339, r.FormValue("start"))
	req.Start = start

	end, _ := time.Parse(time.RFC3339, r.FormValue("end"))
	req.End = end

	bb := strings.Split(r.FormValue("bb"), ",")
	for _, coordStr := range bb {
		coord, err := strconv.ParseFloat(coordStr, 64)
		if err != nil {
			continue
		}
		req.Bounds = append(req.Bounds, coord)
	}

	fmt.Println("eventReq", req)

	eventMsgs, err := a.getMessages(req)
	if err != nil {
		fmt.Fprintln(os.Stderr, "[error]", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/x-msgpack")
	buf, err := msgpack.Marshal(eventMsgs)
	if err != nil {
		fmt.Fprintln(os.Stderr, "[error]", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(buf)
}

func (a *API) HandleEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	var req GetEventsRequest

	start, _ := time.Parse(time.RFC3339, r.FormValue("start"))
	req.Start = start

	end, _ := time.Parse(time.RFC3339, r.FormValue("end"))
	req.End = end

	bb := strings.Split(r.FormValue("bb"), ",")
	for _, coordStr := range bb {
		coord, err := strconv.ParseFloat(coordStr, 64)
		if err != nil {
			continue
		}
		req.Bounds = append(req.Bounds, coord)
	}

	eventMsgs, err := a.getMessages(req)
	if err != nil {
		fmt.Fprintln(os.Stderr, "[error]", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(eventMsgs)
}
