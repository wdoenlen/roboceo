package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/facebook"
)

func main() {
	var (
		clientID      = flag.String("client_id", "879528198763453", "")
		clientSecret  = flag.String("client_secret", "876f827f013b130184ec39cbd836069e", "")
		username      = flag.String("username", "throwaway.wh@gmail.com", "")
		password      = flag.String("password", "scrapescrape", "")
		dbPath        = flag.String("db", "app.db", "sqlite db location")
		port          = flag.Int("port", 8080, "http port")
		webdriverAddr = flag.String("webdriver_addr", "0.0.0.0:4444", "address of webdriver instance")
		redirectAddr  = flag.String("redirect_addr", ":3545", "")
		redirectURL   = flag.String("redirect_url", "http://127.0.0.1:3545/", "")
	)
	flag.Parse()

	rand.Seed(time.Now().UnixNano())

	db, err := NewDB(*dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	conf := &oauth2.Config{
		ClientID:     *clientID,
		ClientSecret: *clientSecret,
		Endpoint:     facebook.Endpoint,
	}

	provider, err := NewOAuthProvider(*conf, *redirectAddr, *redirectURL)
	if err != nil {
		log.Fatal(err)
	}

	scraper := &Scraper{
		db:            db,
		webdriverAddr: *webdriverAddr,
		oauthProvider: provider,
		FBUsername:    *username,
		FBPassword:    *password,
	}

	api := &API{
		scraper: scraper,
		db:      db,
	}

	http.Handle("/", http.FileServer(http.Dir("www")))
	http.HandleFunc("/scrape", api.HandleScrape)
	http.HandleFunc("/events", api.HandleEvents)

	addr := fmt.Sprint(":", *port)
	fmt.Fprintln(os.Stderr, "listening at", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
