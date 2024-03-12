package models

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type Series struct {
	Name       string `json:"name" redis:"name"`
	Handle     string `json:"handle" redis:"handle"`
	URL        string `json:"url" redis:"url"`
	LastUpdate int64  `json:"last_update" redis:"last_update"`
}

type SeriesModel struct {
	DB *redis.Client
}

func (s *SeriesModel) GetAllSeries() ([]*Series, error) {
	ctx := context.Background()

	// Get the last 100 keys that match the pattern "series:*"
	keys, _, err := s.DB.Scan(ctx, 0, "series:*", 100).Result()
	if err != nil {
		return nil, err
	}

	// Create a slice of Series to hold the data
	var series []*Series

	// TODO: Turn into a transaction to reduce round trips
	// Loop through the keys and get the data for each series
	for _, key := range keys {
		data, err := s.DB.HGetAll(ctx, key).Result()
		if err != nil {
			return nil, err
		}

		// Convert the last_update field to an int64
		s, err := strconv.ParseInt(data["last_update"], 10, 64)
		t := time.Unix(s, 0)
		if err != nil {
			return nil, err
		}

		series = append(series, &Series{
			Name:       data["name"],
			Handle:     data["handle"],
			URL:        data["url"],
			LastUpdate: t.Unix(),
		})
	}

	// Return the slice of Series
	return series, nil
}

func (s *SeriesModel) SetSeries(series Series) error {
	ctx := context.Background()

	key := fmt.Sprintf("series:%s", series.Handle)
	_, err := s.DB.HSet(ctx, key, "name", series.Name, "url", series.URL, "last_update", series.LastUpdate, "handle", series.Handle).Result()
	if err != nil {
		return err
	}
	return nil
}

func (s *SeriesModel) GetSeries(handle string) (*Series, error) {
	ctx := context.Background()

	key := fmt.Sprintf("series:%s", handle)
	data, err := s.DB.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	str, err := strconv.ParseInt(data["last_update"], 10, 64)
	t := time.Unix(str, 0)
	if err != nil {
		return nil, err
	}

	series := &Series{
		Name:       data["name"],
		Handle:     data["handle"],
		URL:        data["url"],
		LastUpdate: t.Unix(),
	}

	return series, nil
}
