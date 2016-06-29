package main

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/cznic/ql"
	"github.com/maxhawkins/eventscrape/graphapi"
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

	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`DELETE FROM events WHERE id == $1`, parsedEvent.ID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
			INSERT INTO events
			(id, data, start_time, end_time, latitude, longitude)
			VALUES
			($1, $2, $3, $4, $5, $6)`,
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

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (d *DB) GetEvents(req GetEventsRequest) ([][]byte, error) {
	query := squirrel.
		Select("data").
		From("events").
		PlaceholderFormat(squirrel.Dollar)

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

	queryStr, queryArgs, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := d.db.Query(queryStr, queryArgs...)
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
	ql.RegisterDriver()

	db, err := sql.Open("ql", path)
	if err != nil {
		return nil, err
	}

	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`CREATE TABLE IF NOT EXISTS events (
		id string NOT NULL,
		data string NOT NULL,
		start_time time,
		end_time time,
		latitude float64,
		longitude float64)`)
	if err != nil {
		return nil, err
	}

	tx.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS ON events (id)`)
	tx.Exec(`CREATE INDEX IF NOT EXISTS ON events (start_time, end_time, latitude, longitude)`)

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &DB{db}, nil
}
