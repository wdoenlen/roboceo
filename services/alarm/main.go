package main

import (
	"encoding/gob"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/sideshow/apns2"
	"github.com/sideshow/apns2/certificate"
	"github.com/sideshow/apns2/payload"
)

var client *apns2.Client

var sendersMu sync.Mutex
var senders = make(map[string]*Sender)

type Sender struct {
	Token string
	stop  chan struct{}
}

func (s *Sender) Close() {
	s.stop <- struct{}{}
}

func getTask(isWork bool) (string, error) {
	u := "http://backend.machineexecutive.com/scheduler/task?context=anywhere&work="
	if isWork {
		u += "true"
	} else {
		u += "false"
	}

	resp, err := http.Get(u)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	txt, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(txt), nil
}

func (s *Sender) SendLoop() {
	for {
		duration := time.Duration(rand.Int63())%(75*time.Minute) + (15 * time.Minute)
		duration = duration - (duration % time.Minute) // round
		isWork := rand.Float64() > 0.5

		endTime := time.Now().Add(duration)
		zone, err := time.LoadLocation("Asia/Tokyo")
		if err != nil {
			log.Fatal(err)
		}
		endTimeStr := endTime.In(zone).Format("15:04")

		task, err := getTask(isWork)
		if err != nil {
			log.Println("Error:", err)
			continue
		}

		alertMsg := fmt.Sprintf("%s until %s", task, endTimeStr)

		fmt.Println("send pushes...", s.Token)

		pload := payload.NewPayload().
			Alert(alertMsg).
			Sound("voicemail.caf")

		notification := &apns2.Notification{
			DeviceToken: s.Token,
			Topic:       "com.executivemachine.Alarm",
			Payload:     pload,
		}

		resp, err := client.Push(notification)
		if err != nil {
			log.Println("Error:", err)
			continue
		}
		if resp.StatusCode != http.StatusOK {
			log.Println("Error:", resp.Reason)
			continue
		}

		time.Sleep(duration)
	}
}

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

	var handler http.Handler
	handler = r
	handler = handlers.LoggingHandler(os.Stderr, handler)

	addr := fmt.Sprint(":", *port)
	fmt.Println("listening at", addr)
	http.ListenAndServe(addr, handler)
}
