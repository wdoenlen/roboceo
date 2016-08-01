package main

import (
	"encoding/gob"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/sideshow/apns2"
	"github.com/sideshow/apns2/certificate"
)

var client *apns2.Client

var sendersMu sync.Mutex
var senders = make(map[string]*Sender)

var currentSchedule string

func init() {
	sendersMu.Lock()
	defer sendersMu.Unlock()

	f, err := os.Open("db.gob")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()

	if err := gob.NewDecoder(f).Decode(&senders); err != nil {
		fmt.Println(err)
		return
	}

	for _, sender := range senders {
		go sender.SendLoop()
	}
}

func SaveSender(token string, sender *Sender) {
	senders[token] = sender

	f, err := os.Create("db.gob")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	if err := gob.NewEncoder(f).Encode(senders); err != nil {
		log.Fatal(err)
	}
}

func HandleSetSchedule(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	schedule := r.FormValue("schedule")

	currentSchedule = schedule

	tasks, err := ParseSchedule(schedule)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	now := Now()
	var futureTasks []Task
	for _, task := range tasks {
		if task.Time < now {
			continue
		}
		futureTasks = append(futureTasks, task)
	}

	sendersMu.Lock()
	for _, sender := range senders {
		sender.Todo = futureTasks
	}
	sendersMu.Unlock()

	fmt.Fprintf(w, "loaded %d tasks\n", len(futureTasks))
}

func HandleIndex(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, `<html>
		<head>
			<title></title>
			<meta charset="utf-8">
		</head>
		<body>
			<form action="" method="POST">
				<textarea rows=15 cols=50 name="schedule">%s</textarea>
				<br>
				<br>
				<input type="submit" value="update schedule">
			</form>
		</body>
	</html>`, currentSchedule)
}

func HandleRegister(w http.ResponseWriter, r *http.Request) {
	token := r.FormValue("token")
	if token == "" {
		http.Error(w, "no token", http.StatusBadRequest)
		return
	}

	sendersMu.Lock()
	defer sendersMu.Unlock()

	if _, ok := senders[token]; ok {
		fmt.Fprintln(w, "already registered")
		return
	}
	sender := &Sender{
		Token: token,
		stop:  make(chan struct{}),
	}
	SaveSender(token, sender)
	go sender.SendLoop()
}

func main() {
	var (
		port     = flag.Int("port", 8080, "")
		certPath = flag.String("cert", "cert.pem", "")
	)
	flag.Parse()

	cert, err := certificate.FromPemFile(*certPath, "")
	if err != nil {
		log.Fatal("Cert Error:", err)
	}

	client = apns2.NewClient(cert).Development()

	r := mux.NewRouter()
	r.HandleFunc("/register", HandleRegister)
	r.HandleFunc("/schedule", HandleIndex).Methods("GET")
	r.HandleFunc("/schedule", HandleSetSchedule).Methods("POST")

	var handler http.Handler
	handler = r
	handler = handlers.LoggingHandler(os.Stderr, handler)

	addr := fmt.Sprint(":", *port)
	fmt.Println("listening at", addr)
	http.ListenAndServe(addr, handler)
}
