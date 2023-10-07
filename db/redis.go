package db

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/bairrya/sj-rss/config"
	"github.com/redis/go-redis/v9"
)

func RedisConnect() (*redis.Client, error) {
	config, err := config.ENV()
	if err != nil {
		log.Printf("Error loading config: %s", err)
		return nil, err
	}
	opt, err := redis.ParseURL(config.REDIS_URI)
	if err != nil {
		log.Printf("Error parsing redis url: %s", config.REDIS_URI)
		return nil, err
	}

	client := redis.NewClient(opt)

	return client, nil
}

func SetSeries(series Series) error {
	ctx := context.Background()
	redis, err := RedisConnect()
	if err != nil {
		return err
	}

	key := fmt.Sprintf("series:%s", series.Handle)
	_, err = redis.HSet(ctx, key, "name", series.Name, "url", series.URL, "last_update", series.LastUpdate, "handle", series.Handle).Result()
	if err != nil {
		return err
	}
	log.Println("Saved series:", series.Handle)
	return nil
}

func GetSeries(handle string) (*Series, error) {
	return nil, nil
}

func GetAllSeries() (series []Series, err error) {
	ctx := context.Background()

	redis, err := RedisConnect()
	if err != nil {
		log.Println("Error connecting to redis:", err)
		return nil, err
	}

	keys, _, err := redis.Scan(ctx, 0, "series:*", 100).Result()
	if err != nil {
		log.Println("Error getting keys:", err)
		return nil, err
	}
	for _, key := range keys {
		data, err := redis.HGetAll(ctx, key).Result()
		if err != nil {
			log.Println("Error getting series:", err)
			return nil, err
		}

		s, err := strconv.ParseInt(data["last_update"], 10, 64)
		tm := time.Unix(s, 0)
		if err != nil {
			log.Println("Error parsing last_update:", err)
			return nil, err
		}

		series = append(series, Series{
			Name:       data["name"],
			Handle:     data["handle"],
			URL:        data["url"],
			LastUpdate: tm.Unix(),
		})
	}

	return series, nil
}
