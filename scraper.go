package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/maxhawkins/eventscrape/graphapi"
	"github.com/maxhawkins/eventscrape/scraper"
	"github.com/tebeka/selenium"
)

type Scraper struct {
	db            *sql.DB
	webdriverAddr string
	apiClient     *http.Client
	FBUsername    string
	FBPassword    string
}

type ScrapeRequest struct {
	Location string `json:"location"`
	Tomorrow bool   `json:"tomorrow"`
}

func (s *Scraper) Scrape(req *ScrapeRequest) error {
	wd, err := newWebDriver(s.webdriverAddr)
	if err != nil {
		return err
	}
	defer wd.Quit()

	placeID, err := graphapi.GetPlaceID(s.apiClient, req.Location)
	if err != nil {
		return err
	}
	day := "today"
	if req.Tomorrow {
		day = "tomorrow"
	}
	todayURL := fmt.Sprintf("https://www.facebook.com/search/%s/events-near/%s/date/events/intersect", placeID, day)

	ids := make(chan string)
	events := make(chan graphapi.Event)

	go func() {
		if err := scraper.GetAllEvents(wd, todayURL, s.FBUsername, s.FBPassword, ids); err != nil {
			log.Println(err)
		}
		close(ids)
	}()

	getEvents := func(ids []string) {
		result, err := graphapi.GetEvents(s.apiClient, ids)
		if err != nil {
			fmt.Fprintln(os.Stderr, "GetEvents:", err)
			return
		}
		for _, r := range result {
			events <- r
		}
	}

	go func() {
		var batch []string
		for id := range ids {
			batch = append(batch, id)

			if len(batch) > 20 {
				getEvents(batch)
				batch = nil
			}
		}
		if batch != nil {
			getEvents(batch)
			batch = nil
		}
		close(events)
	}()

	for event := range events {
		var e struct {
			ID        string `json:"id"`
			StartTime FBTime `json:"start_time"`
			EndTime   FBTime `json:"end_time"`
			Place     struct {
				Location struct {
					Latitude  float64 `json:"latitude"`
					Longitude float64 `json:"longitude"`
				} `json:"location"`
			} `json:"place"`
		}
		if err := json.Unmarshal([]byte(*event), &e); err != nil {
			return err
		}

		_, err := s.db.Exec(`INSERT OR REPLACE INTO events
			(id, data, start_time, end_time, latitude, longitude)
			VALUES
			(?, ?, ?, ?, ?, ?)`,
			e.ID,
			string(*event),
			time.Time(e.StartTime).UTC(),
			time.Time(e.EndTime).UTC(),
			e.Place.Location.Latitude,
			e.Place.Location.Longitude,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func newWebDriver(addr string) (selenium.WebDriver, error) {
	for i := 0; i < 50; i++ {
		conn, err := net.Dial("tcp", addr)
		if err == nil {
			conn.Close()
			break
		}

		fmt.Printf("webdriver connect: %v\n", err)
		time.Sleep(500 * time.Millisecond)
	}

	caps := selenium.Capabilities{
		"browserName": "firefox",
	}
	wd, err := selenium.NewRemote(caps, "")
	if err != nil {
		return nil, err
	}

	if err := wd.SetImplicitWaitTimeout(1 * time.Second); err != nil {
		return nil, err
	}

	return wd, nil
}

type FBTime time.Time

func (f *FBTime) UnmarshalJSON(data []byte) error {
	t, err := time.Parse(`"2006-01-02T15:04:05-0700"`, string(data))
	if err != nil {
		return err
	}
	*f = FBTime(t)
	return nil
}
