package models

import (
	"database/sql"
)

type Series struct {
	Name       string `json:"name"`
	Handle     string `json:"handle"`
	URL        string `json:"url"`
	LastUpdate int64  `json:"last_update"`
}

type SeriesRow struct {
	ID int
	Series
}

type SeriesModel struct {
	DB *sql.DB
}

func (s *SeriesModel) SelectAllSeries() ([]*SeriesRow, error) {
	rows, err := s.DB.Query("SELECT * FROM series")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Create a slice of Series to hold the data
	var seriesRow []*SeriesRow

	// Iterate over the rows
	for rows.Next() {
		// Create a new Series
		var s SeriesRow

		// Scan the rows into the Series
		err := rows.Scan(&s.ID, &s.Name, &s.Handle, &s.URL, &s.LastUpdate)
		if err != nil {
			return nil, err
		}

		seriesRow = append(seriesRow, &s)
	}

	return seriesRow, nil
}

func (s *SeriesModel) UpsertSeries(series Series) error {
	// Create a statement to upcert the series
	_, err := s.DB.Exec("INSERT INTO series (name, handle, url, last_update) VALUES (?, ?, ?, ?) ON CONFLICT (handle) DO UPDATE SET name = ?, url = ?, last_update = ?", series.Name, series.Handle, series.URL, series.LastUpdate, series.Name, series.URL, series.LastUpdate)
	if err != nil {
		return err
	}
	return nil
}
