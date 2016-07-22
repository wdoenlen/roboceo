package main

import "time"

func ConvertEvent(fbEvent FBEvent) Event {
	return Event{
		ID:          fbEvent.ID,
		Name:        fbEvent.Name,
		Description: fbEvent.Description,
		Latitude:    fbEvent.Place.Location.Latitude,
		Longitude:   fbEvent.Place.Location.Longitude,
		StartTime:   time.Time(fbEvent.StartTime),
		EndTime:     time.Time(fbEvent.EndTime),
	}
}

type FBEvent struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	StartTime   FBTime `json:"start_time"`
	EndTime     FBTime `json:"end_time"`
	Place       struct {
		Location struct {
			Latitude  float64 `json:"latitude"`
			Longitude float64 `json:"longitude"`
		} `json:"location"`
	} `json:"place"`
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
