package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/sideshow/apns2"
	"github.com/sideshow/apns2/payload"
)

type Sender struct {
	Token string
	Todo  []Task
	stop  chan struct{}
}

func (s *Sender) Close() {
	s.stop <- struct{}{}
}

// func getTask(isWork bool) (string, error) {
// 	u := "http://backend.machineexecutive.com/scheduler/task?context=anywhere&work="
// 	if isWork {
// 		u += "true"
// 	} else {
// 		u += "false"
// 	}

// 	resp, err := http.Get(u)
// 	if err != nil {
// 		return "", err
// 	}
// 	defer resp.Body.Close()

// 	txt, err := ioutil.ReadAll(resp.Body)
// 	if err != nil {
// 		return "", err
// 	}

// 	return string(txt), nil
// }

func Now() time.Duration {
	zone, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		log.Fatal(err)
	}
	now := time.Now().In(zone)

	return time.Duration(now.Hour())*time.Hour +
		time.Duration(now.Minute())*time.Minute
}

func (s *Sender) Send(message string) error {
	fmt.Println("sending...", message, s.Token)

	pload := payload.NewPayload().
		Alert(message).
		Sound("voicemail.caf")

	notification := &apns2.Notification{
		DeviceToken: s.Token,
		Topic:       "com.executivemachine.Alarm",
		Payload:     pload,
	}

	resp, err := client.Push(notification)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("push: %s", resp.Reason)
	}

	fmt.Println("sent.")

	return nil
}

func (s *Sender) SendNext() error {
	now := Now()

	var todo []Task
	var todoLater []Task

	for _, task := range s.Todo {
		if task.Time > now {
			todoLater = append(todoLater, task)
		} else {
			todo = append(todo, task)
		}
	}

	if len(todo) == 0 {
		return nil
	}

	todoNow := todo[len(todo)-1]

	if err := s.Send(todoNow.Description); err != nil {
		return err
	}

	s.Todo = todoLater

	return nil
}

func (s *Sender) SendLoop() {
	for {
		if err := s.SendNext(); err != nil {
			log.Println("Error:", err)
		}

		time.Sleep(time.Minute)
	}
}
