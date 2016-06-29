package main

import (
	"errors"
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
	db            *DB
	webdriverAddr string
	oauthProvider *OAuthProvider
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

	authURL, clientChan := s.oauthProvider.RequestClient()

	if err := scraper.EnsureLoggedIn(wd, s.FBUsername, s.FBPassword); err != nil {
		return err
	}

	if err := wd.Get(authURL); err != nil {
		return err
	}

	var apiClient *http.Client
	select {
	case apiClient = <-clientChan:
		if apiClient == nil {
			return errors.New("failed to get api client")
		}
		break
	case <-time.After(20 * time.Second):
		return errors.New("timed out waiting for oauth token")
	}

	placeID, err := graphapi.GetPlaceID(apiClient, req.Location)
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
		if err := scraper.GetAllEvents(wd, todayURL, ids); err != nil {
			log.Println(err)
		}
		close(ids)
	}()

	getEvents := func(ids []string) {
		result, err := graphapi.GetEvents(apiClient, ids)
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
		if err := s.db.SaveEvent(event); err != nil {
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
