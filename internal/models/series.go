package models

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
)

type Series struct {
	Name        string `json:"name"`
	Handle      string `json:"handle"`
	Description string `json:"description"`
	CreatedBy   string `json:"created_by"`
	Icon        string `json:"icon"`
	Hero        string `json:"hero"`
	RecentURL   string `json:"recent_url"`
	WebOnly     bool   `json:"web_only"`
	OneShot     bool   `json:"one_shot"`
	LastUpdate  int64  `json:"last_update"`
}

type SeriesModel struct {
	Logger *slog.Logger
}

func (s *SeriesModel) GetAllSeries() ([]*Series, error) {
	s.Logger.Info("Getting all series")
	// Create a slice of Series to hold the data
	var series []*Series

	// Get all the json files from the data folder
	files, err := os.ReadDir("./data")
	if err != nil {
		s.Logger.Error("Error reading data folder", err)
		return nil, err
	}

	// Loop through the files
	for _, file := range files {
		// Skip any files that are not json
		if file.IsDir() || filepath.Ext(file.Name()) != ".json" {
			continue
		}

		// Open the file
		f, err := os.Open("./data/" + file.Name())
		if err != nil {
			s.Logger.Error("Error opening file", err)
			return nil, err
		}
		defer f.Close()

		// Create a new Series to hold the data
		var ser Series

		// Decode the json file into the Series struct
		err = json.NewDecoder(f).Decode(&ser)
		if err != nil {
			s.Logger.Error("Error decoding json", err)
			return nil, err
		}

		// Append the Series to the slice
		series = append(series, &ser)

	}

	// Return the slice of Series
	return series, nil
}

func (s *SeriesModel) SetSeries(series Series) error {
	s.Logger.Info("Setting series:" + series.Handle)

	data, err := json.Marshal(series)
	if err != nil {
		s.Logger.Error("Error marshalling series", err)
		return err
	}
	err = os.WriteFile("./data/"+series.Handle+".json", data, 0644)
	if err != nil {
		s.Logger.Error("Error writing series file", err)
		return err
	}

	return nil
}

func (s *SeriesModel) GetSeries(handle string) (*Series, error) {
	s.Logger.Info("Getting series")

	// Open the file
	f, err := os.Open(handle + ".json")
	if err != nil {
		s.Logger.Error("Error opening file", err)
		return nil, err
	}
	defer f.Close()

	// Convert the file to a Series struct
	var series Series
	err = json.NewDecoder(f).Decode(&series)
	if err != nil {
		s.Logger.Error("Error decoding json", err)
		return nil, err
	}

	return &series, nil
}
