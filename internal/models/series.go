package models

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
)

type Meta struct {
	Name       string `json:"name"`
	Handle     string `json:"handle"`
	Icon       string `json:"icon"`
	LastUpdate int64  `json:"last_update"`
	RecentURL  string `json:"recent_url"`
}

type Series struct {
	Meta
	Description string   `json:"description"`
	CreatedBy   string   `json:"created_by"`
	Hero        string   `json:"hero"`
	WebOnly     bool     `json:"web_only"`
	OneShot     bool     `json:"one_shot"`
	Volumes     []Volume `json:"volumes"`
}

type Volume struct {
	Number   int       `json:"number"`
	URL      string    `json:"url"`
	Image    string    `json:"image"`
	Chapters []float64 `json:"chapters"`
}

type SeriesModel struct {
	Logger *slog.Logger
}

// GetAllSeries reads all the series json files from the data folder and returns a slice of Series
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

// SetSeries writes a Series struct to a json file in the data folder
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

// GetSeries reads a json file from the data folder and returns a Series struct
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
