package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/maxhawkins/eventscrape/graphapi"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	db *sql.DB
}

type GetEventsRequest struct {
	Bounds []string
	Start  time.Time
	End    time.Time
}

func (d *DB) Close() error {
	return d.db.Close()
}

func (d *DB) SaveEvent(event graphapi.Event) error {
	var parsedEvent struct {
		ID        string `json:"id"`
		StartTime FBTime `json:"start_time"`
		EndTime   FBTime `json:"end_time"`
		Place     struct {
			Location struct {
				Latitude  float64 `json:"latitude"`
				Longitude float64 `json:"longitude"`
			} `json:"location"`
		} `json:"place"`
	}
	if err := json.Unmarshal(*event, &parsedEvent); err != nil {
		return err
	}

	_, err := d.db.Exec(`
			INSERT OR REPLACE INTO events
			(id, data, start_time, end_time, latitude, longitude)
			VALUES
			(?, ?, ?, ?, ?, ?)`,
		parsedEvent.ID,
		string(*event),
		time.Time(parsedEvent.StartTime).UTC(),
		time.Time(parsedEvent.EndTime).UTC(),
		parsedEvent.Place.Location.Latitude,
		parsedEvent.Place.Location.Longitude,
	)
	if err != nil {
		return err
	}

	return nil
}

func (d *DB) GetEvents(req GetEventsRequest) ([][]byte, error) {
	query := squirrel.Select("data").From("events")

	if len(req.Bounds) == 4 {
		query = query.Where(squirrel.GtOrEq{"latitude": req.Bounds[0]})
		query = query.Where(squirrel.GtOrEq{"longitude": req.Bounds[1]})
		query = query.Where(squirrel.LtOrEq{"latitude": req.Bounds[2]})
		query = query.Where(squirrel.LtOrEq{"longitude": req.Bounds[3]})
	}

	if !req.Start.IsZero() {
		query = query.Where(squirrel.GtOrEq{"start_time": req.Start.UTC()})
	}
	if !req.End.IsZero() {
		query = query.Where(squirrel.Lt{"start_time": req.End.UTC()})
	}

	rows, err := query.RunWith(d.db).Query()
	if err != nil {
		return nil, err
	}

	var events [][]byte

	for rows.Next() {
		var data []byte
		if err := rows.Scan(&data); err != nil {
			return nil, err
		}
		events = append(events, data)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return events, nil
}

func NewDB(path string) (*DB, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		log.Fatal(err)
	}

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

	return &DB{db}, nil
}
