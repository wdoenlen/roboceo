package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/oauth2/clientcredentials"
	"golang.org/x/oauth2/facebook"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	var (
		clientID      = flag.String("client_id", "879528198763453", "")
		clientSecret  = flag.String("client_secret", "876f827f013b130184ec39cbd836069e", "")
		username      = flag.String("username", "throwaway.wh@gmail.com", "")
		password      = flag.String("password", "scrapescrape", "")
		dbAddr        = flag.String("db", "app.db", "sqlite db location")
		port          = flag.Int("port", 8080, "http port")
		webdriverAddr = flag.String("webdriver_addr", "0.0.0.0:4444", "address of webdriver instance")
	)
	flag.Parse()

	rand.Seed(time.Now().UnixNano())

	db, err := sql.Open("sqlite3", *dbAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS events (
		id TEXT PRIMARY KEY,
		data TEXT NOT NULL,
		start_time TIMESTAMP,
		end_time TIMESTAMP,
		latitude DOUBLE,
		longitude DOUBLE)`)
	if err != nil {
		log.Fatal(err)
	}

	conf := clientcredentials.Config{
		ClientID:     *clientID,
		ClientSecret: *clientSecret,
		TokenURL:     facebook.Endpoint.TokenURL,
	}
	client := conf.Client(context.Background())

	scraper := &Scraper{
		db:            db,
		webdriverAddr: *webdriverAddr,
		apiClient:     client,
		FBUsername:    *username,
		FBPassword:    *password,
	}

	dbWrap := &DB{db}

	api := &API{
		scraper: scraper,
		db:      dbWrap,
	}

	http.Handle("/", http.FileServer(http.Dir("www")))
	http.HandleFunc("/scrape", api.HandleScrape)
	http.HandleFunc("/events", api.HandleEvents)

	addr := fmt.Sprint(":", *port)
	fmt.Fprintln(os.Stderr, "listening at", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
