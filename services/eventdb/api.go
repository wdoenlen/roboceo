package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// API provides a REST API for events
type API struct {
	db *DB
}

type Event struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Latitude    float64   `json:"latitude,omitempty"`
	Longitude   float64   `json:"longitude,omitempty"`
	StartTime   time.Time `json:"start_time,omitempty"`
	EndTime     time.Time `json:"end_time,omitempty"`
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

// HACK(maxhawkins): this method has a lot of temporary hacks to deal
// with filtering events, something we weren't doing serverside before
// but is necessary to reduce data consumption.
func (a *API) getMessages(req GetEventsRequest) ([]Event, error) {

	events, err := a.db.GetEvents(req)
	if err != nil {
		return nil, err
	}

	filteredEvents := make([]Event, 0)
	for _, event := range events {
		// TODO(maxhawkins): don't actually want to modify this here
		if event.EndTime.IsZero() {
			// if there's no end time listed assume it's 1hr long
			event.EndTime = time.Time(event.StartTime).Add(1 * time.Hour)
		}
		if event.EndTime.Sub(event.StartTime) > 24*time.Hour {
			// really long events are usually bar happy hours and promotions
			continue
		}

		filteredEvents = append(filteredEvents, event)
	}

	return filteredEvents, nil
}

type EventsReply struct {
	TotalResults int     `json:"total_results"`
	Events       []Event `json:"events"`
}

func (a *API) HandleAdd(w http.ResponseWriter, r *http.Request) {
	var fbEvents []FBEvent
	if err := json.NewDecoder(r.Body).Decode(&fbEvents); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	events := make([]Event, len(fbEvents))
	for i := range fbEvents {
		events[i] = ConvertEvent(fbEvents[i])
	}

	if err := a.db.SaveEvents(events); err != nil {
		fmt.Fprintln(os.Stderr, "[error]", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, "got it")
}

func (a *API) HandleList(w http.ResponseWriter, r *http.Request) {
	var req GetEventsRequest

	start, _ := time.Parse(time.RFC3339, r.FormValue("start"))
	req.Start = start

	end, _ := time.Parse(time.RFC3339, r.FormValue("end"))
	req.End = end

	limit, _ := strconv.Atoi(r.FormValue("limit"))
	if limit <= 0 {
		limit = 500
	}
	req.Limit = limit

	bb := strings.Split(r.FormValue("bb"), ",")
	for _, coordStr := range bb {
		coord, err := strconv.ParseFloat(coordStr, 64)
		if err != nil {
			continue
		}
		req.Bounds = append(req.Bounds, coord)
	}

	events, err := a.getMessages(req)
	if err != nil {
		fmt.Fprintln(os.Stderr, "[error]", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Randomly drop events until we match the requested result limit
	limited := make([]Event, 0)
	perm := rand.Perm(len(events))
	if limit > len(events) {
		limit = len(events)
	}
	for _, i := range perm[:limit] {
		limited = append(limited, events[i])
	}

	for i := range limited {
		limited[i].Name = truncate(limited[i].Name, 50)
		limited[i].Description = truncate(limited[i].Description, 50)
	}

	reply := EventsReply{
		TotalResults: len(events),
		Events:       limited,
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(reply)
}
