package graphapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

type BatchResponse struct {
	Code int    `json:"code"`
	Body string `json:"body"`
}

type Event *json.RawMessage

var fields = "place,cover,attending_count,declined_count,description,end_time,id,maybe_count,name,noreply_count,owner,start_time,ticket_uri"

func GetEvents(client *http.Client, ids []string) ([]Event, error) {
	reqs := make([]map[string]string, len(ids))
	for i, id := range ids {
		reqs[i] = map[string]string{
			"method":       "GET",
			"relative_url": fmt.Sprintf("v2.5/%s?fields=%s", id, fields),
		}
	}
	req := map[string]interface{}{"batch": reqs}

	batchBody := bytes.NewBuffer(nil)
	if err := json.NewEncoder(batchBody).Encode(req); err != nil {
		return nil, err
	}

	resp, err := client.Post("https://graph.facebook.com", "application/json", batchBody)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("bad response: %s", string(body))
	}

	var responses []BatchResponse
	if err := json.NewDecoder(resp.Body).Decode(&responses); err != nil {
		return nil, err
	}

	var parsed []Event
	for i, r := range responses {
		if r.Code != 200 {
			fmt.Fprintf(os.Stderr, "bad response for %s: %s\n", ids[i], r.Body)
			continue
		}
		var msg *json.RawMessage
		if err := json.Unmarshal([]byte(r.Body), &msg); err != nil {
			return nil, err
		}
		parsed = append(parsed, Event(msg))
	}

	return parsed, nil
}
