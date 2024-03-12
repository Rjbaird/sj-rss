package main

import (
	"html/template"
	"log"
	"os"
	"time"

	"github.com/bairrya/sj-rss/internal/models"
	"github.com/bairrya/sj-rss/internal/web"
	"github.com/gorilla/feeds"
	"github.com/redis/go-redis/v9"
)

func (s *server) updateFeeds() error {
	const baseURL = "https://www.viz.com"
	const recentURL = baseURL + "/read/shonenjump/section/free-chapters"

	now := time.Now()

	// main feed
	feed := &feeds.Feed{
		Title:       "Weekly Shonen Jump",
		Link:        &feeds.Link{Href: recentURL},
		Description: "The world's most popular manga!",
		Author:      &feeds.Author{Name: "Shonen Jump | VIZ"},
		Created:     now,
	}

	results, err := web.GetRecentChapters()
	if err != nil {
		return err
	}

	feed.Items = results.Feed

	atom, err := feed.ToAtom()
	if err != nil {
		log.Println("Error converting series feed to atom:", err)
		return err
	}

	// create main feed
	err = createXML(s.config.rssPath, "main", atom)
	if err != nil {
		log.Println("Error creating main feed:", err)
		return err
	}

	// Get the redis url from the environment
	redisURL := os.Getenv("REDIS_URL")
	options, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Println("Error parsing redis url", err)
		return err
	}

	// Create a new redis client connection
	client := redis.NewClient(options)
	defer client.Close()

	model := &models.SeriesModel{DB: client}

	// use handles to scrape series
	for _, manga := range results.Series {
		mangaFeed, err := web.GetSeriesData(manga.Handle)
		if err != nil {
			log.Println("Error updating series feed:", err)
		}

		mAtom, err := mangaFeed.ToAtom()
		if err != nil {
			log.Println("Error converting series feed to atom:", err)
		}

		// create series feed
		log.Println("Creating feed for", manga.Handle)
		createXML(s.config.rssPath, manga.Handle, mAtom)

		err = model.SetSeries(manga)
		if err != nil {
			log.Println("Error setting series:", err)
			return err
		}
		time.Sleep(3 * time.Second)
	}
	return nil
}

func (s *server) generateIndex() error {
	s.logger.Info("Generating index.html...")

	// Get all the series from the database
	series, err := s.series.GetAllSeries()
	if err != nil {
		s.logger.Error("Error getting all series", err)
		return err
	}

	// Parse the template files
	tmpl, err := template.ParseFiles("views/layout/base.html")
	if err != nil {
		s.logger.Error("Error parsing index.html", err)
		return err
	}

	// Create a new index.html file
	f, err := os.Create("views/index.html")
	if err != nil {
		s.logger.Error("Error creating index.html", err)
		return err
	}
	defer f.Close()

	// Write the html to the file
	err = tmpl.ExecuteTemplate(f, "base", series)
	if err != nil {
		s.logger.Error("Error writing to index.html", err)
		return err
	}

	s.logger.Info("index.html generated")
	return nil
}

func createXML(path string, name string, atom string) error {
	file, err := os.Create(path + name + ".xml")
	if err != nil {
		log.Println("Error creating file:", err)
		return err
	}
	defer file.Close()

	_, err = file.WriteString(atom)
	if err != nil {
		log.Println("Error writing to xml feed:", err)
		return err
	}
	return nil
}
