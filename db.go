package main

import (
	"database/sql"
	"time"

	"github.com/Masterminds/squirrel"
)

type DB struct {
	db *sql.DB
}

type GetEventsRequest struct {
	Bounds []string
	Start  time.Time
	End    time.Time
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
