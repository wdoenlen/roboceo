package main

import (
	"database/sql"
	"time"

	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

// DB provides a SQL backend Event querying/persistence
type DB struct {
	db *sql.DB
}

type GetEventsRequest struct {
	Bounds []float64
	Start  time.Time
	End    time.Time
	Limit  int
}

func (d *DB) Close() error {
	return d.db.Close()
}

func (d *DB) SaveEvents(events []Event) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, event := range events {
		startTime := event.StartTime.UTC()
		endTime := event.EndTime.UTC()
		if endTime.IsZero() {
			endTime = startTime
		}

		validCoord := !(event.Latitude == 0 && event.Longitude == 0)
		lat := sql.NullFloat64{Float64: event.Latitude, Valid: validCoord}
		lng := sql.NullFloat64{Float64: event.Longitude, Valid: validCoord}

		_, err := tx.Exec(`
			INSERT INTO events
			(id, name, description, start_time, end_time, latitude, longitude)
			VALUES
			($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT (id) DO UPDATE
			SET name=$2, description=$3, start_time=$4, end_time=$5, latitude=$6, longitude=$7`,
			event.ID,
			event.Name,
			event.Description,
			startTime,
			endTime,
			lat,
			lng)
		if err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (d *DB) GetEvents(req GetEventsRequest) ([]Event, error) {
	var args []interface{}
	query := `SELECT
		id, name, description, start_time, end_time, latitude, longitude
		FROM events
		WHERE tstzrange($1, $2) && tstzrange(start_time, COALESCE(end_time, start_time))`

	args = append(args,
		pq.NullTime{Time: req.Start.UTC(), Valid: !req.Start.IsZero()},
		pq.NullTime{Time: req.End.UTC(), Valid: !req.End.IsZero()})

	if len(req.Bounds) == 4 {
		query += ` AND box(point($3, $4), point($5, $6)) @> point(latitude, longitude)`
		args = append(args, req.Bounds[0], req.Bounds[1], req.Bounds[2], req.Bounds[3])
	}

	rows, err := d.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []Event

	for rows.Next() {
		var start, end pq.NullTime
		var lat, lng sql.NullFloat64

		var event Event
		err = rows.Scan(
			&event.ID,
			&event.Name,
			&event.Description,
			&start,
			&end,
			&lat,
			&lng,
		)
		if err != nil {
			return nil, err
		}

		event.StartTime = start.Time
		event.EndTime = end.Time
		event.Latitude = lat.Float64
		event.Longitude = lng.Float64

		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return events, nil
}

func NewDB(addr string) (*DB, error) {
	db, err := sql.Open("postgres", addr)
	if err != nil {
		return nil, err
	}

	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	tx.Exec(`CREATE TABLE IF NOT EXISTS events (
		id varchar(40) NOT NULL,
		name text NOT NULL,
		description text NOT NULL,
		start_time timestamptz NOT NULL,
		end_time timestamptz,
		latitude float,
		longitude float)`)

	tx.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS id_idx ON events (id)`)
	tx.Exec(`CREATE INDEX IF NOT EXISTS time_idx ON events (start_time, end_time)`)
	tx.Exec(`CREATE INDEX IF NOT EXISTS location_idx ON events USING gist (point(latitude, longitude))`)

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &DB{db}, nil
}
