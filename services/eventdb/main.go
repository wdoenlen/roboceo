package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func main() {
	var (
		dbPath = flag.String("db", "postgres://localhost/eventdb?sslmode=disable", "sqlite db location")
		port   = flag.Int("port", 8080, "http port")
		// yandexKey = flag.String("yandex_key", "trnsl.1.1.20160630T124034Z.39e1ba8746eb8752.86b1c259847b43cce1f510b3bce30942745aeab5", "")
	)
	flag.Parse()

	rand.Seed(time.Now().UnixNano())

	db, err := NewDB(*dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	api := &API{
		db: db,
	}

	m := mux.NewRouter()
	m.HandleFunc("/events", api.HandleList).Methods("GET")
	m.HandleFunc("/events", api.HandleAdd).Methods("POST")

	var handler http.Handler
	handler = handlers.CORS()(m)

	addr := fmt.Sprint(":", *port)
	fmt.Fprintln(os.Stderr, "listening at", addr)
	log.Fatal(http.ListenAndServe(addr, handler))
}
