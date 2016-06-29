package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
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
	req.Bounds = bb

	events, err := a.db.GetEvents(req)
	if err != nil {
		fmt.Fprintln(os.Stderr, "[error]", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte("["))
	w.Write(bytes.Join(events, []byte(",")))
	w.Write([]byte("]"))
}
