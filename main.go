package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/tebeka/selenium"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/facebook"

	"github.com/maxhawkins/eventscrape/graphapi"
	"github.com/maxhawkins/eventscrape/scraper"
)

func main() {
	var (
		clientID     = flag.String("client_id", "879528198763453", "")
		clientSecret = flag.String("client_secret", "876f827f013b130184ec39cbd836069e", "")
		username     = flag.String("username", "throwaway.wh@gmail.com", "")
		password     = flag.String("password", "scrapescrape", "")
		location     = flag.String("location", "Berlin, Germany", "")
		outdir       = flag.String("out", ".", "")
		tomorrow     = flag.Bool("tomorrow", false, "today or tomorrow")
	)
	flag.Parse()

	rand.Seed(time.Now().UnixNano())

	conf := &oauth2.Config{
		ClientID:     *clientID,
		ClientSecret: *clientSecret,
		Endpoint:     facebook.Endpoint,
	}

	client := newOAuthClient(oauth2.NoContext, conf)

	wd, err := newWebDriver("0.0.0.0:4444")
	if err != nil {
		log.Fatal("[error]", err)
	}
	defer wd.Quit()

	placeID, err := graphapi.GetPlaceID(client, *location)
	if err != nil {
		log.Fatal(err)
	}
	day := "today"
	if *tomorrow {
		day = "tomorrow"
	}
	todayURL := fmt.Sprintf("https://www.facebook.com/search/%s/events-near/%s/date/events/intersect", placeID, day)

	fmt.Printf("scraping %q\n", todayURL)

	currentTime := time.Now().Local() // TODO(maxhawkins): make local in location
	if *tomorrow {
		currentTime = currentTime.Add(24 * time.Hour)
	}
	filename := placeID + "_" + currentTime.Format("02-Jan-06") + ".json"
	outPath := filepath.Join(*outdir, filename)

	fmt.Printf("saving to %q\n", filename)

	ids := make(chan string)
	events := make(chan graphapi.Event)

	go func() {
		if err := scraper.GetAllEvents(wd, todayURL, *username, *password, ids); err != nil {
			log.Println(err)
		}
		close(ids)
	}()

	getEvents := func(ids []string) {
		result, err := graphapi.GetEvents(client, ids)
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

	var allEvents []graphapi.Event
	for event := range events {
		allEvents = append(allEvents, event)

		js, err := json.Marshal(allEvents)
		if err != nil {
			log.Fatal(err)
		}
		if _, err := WriteFileSafe(outPath, js); err != nil {
			log.Fatal(err)
		}
	}
}

func WriteFileSafe(filename string, data []byte) (int, error) {
	temp, err := ioutil.TempFile("", "")
	if err != nil {
		return -1, err
	}
	defer os.Remove(temp.Name())

	n, err := temp.Write(data)
	if err != nil {
		return -1, err
	}
	if err := temp.Close(); err != nil {
		return -1, err
	}

	if err := os.Rename(temp.Name(), filename); err != nil {
		return -1, err
	}

	return n, nil
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
