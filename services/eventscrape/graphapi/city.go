package graphapi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

type Place struct {
	ID string `json:"id"`
}

func GetPlaceID(client *http.Client, query string) (string, error) {
	u := "https://graph.facebook.com/search?type=place&q=" + url.QueryEscape(query)
	resp, err := client.Get(u)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf("bad response: %s", string(body))
	}

	var placesResp struct {
		Data []Place `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&placesResp); err != nil {
		return "", err
	}

	if len(placesResp.Data) == 0 {
		return "", fmt.Errorf("no results")
	}

	place := placesResp.Data[0]
	return place.ID, nil
}
