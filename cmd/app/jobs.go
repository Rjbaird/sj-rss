package main

import (
	"html/template"
	"log"
	"os"
	"time"

	"github.com/gorilla/feeds"
	"github.com/rjbaird/sj-rss/internal/web"
)

func (app *application) generateSeriesFeeds() error {
	app.logger.Info("Generating series feeds...")

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
		app.logger.Error("Error getting recent chapters:", err)
		return err
	}

	feed.Items = results.Feed

	atom, err := feed.ToAtom()
	if err != nil {
		app.logger.Error("Error converting feed to atom:", err)
		return err
	}

	// create main feed
	err = createXML(app.config.rssPath, "main", atom)
	if err != nil {
		app.logger.Error("Error creating main feed:", err)
		return err
	}

	// use handles to scrape series
	for _, manga := range results.Series {
		mangaFeed, err := web.GetSeriesData(manga.Handle)
		if err != nil {
			app.logger.Error("Error updating series feed:", err)
			return err
		}

		mAtom, err := mangaFeed.ToAtom()
		if err != nil {
			app.logger.Error("Error converting series feed to atom:", err)
			return err
		}

		// create series feed

		app.logger.Info("Creating feed for " + manga.Handle)
		err = createXML(app.config.rssPath, manga.Handle, mAtom)
		if err != nil {
			app.logger.Error("Error creating series feed:", err)
			return err
		}

		err = app.series.SetSeries(manga)
		if err != nil {
			log.Println("Error setting series:", err)
			return err
		}
		time.Sleep(3 * time.Second)
		msg := "Series feed created for " + manga.Handle
		app.logger.Info(msg)
	}
	app.logger.Info("Series feeds generated")
	return nil
}

func (app *application) generateIndex() error {
	app.logger.Info("Generating index.html...")

	// Get all the series from the database
	series, err := app.series.GetAllSeries()
	if err != nil {
		app.logger.Error("Error getting all series", err)
		return err
	}

	// Parse the template files
	tmpl, err := template.ParseFiles("views/layouts/base.html")
	if err != nil {
		app.logger.Error("Error parsing index.html", err)
		return err
	}

	// Create a new index.html file
	f, err := os.Create("views/index.html")
	if err != nil {
		app.logger.Error("Error creating index.html", err)
		return err
	}
	defer f.Close()

	// Write the html to the file
	err = tmpl.ExecuteTemplate(f, "base", series)
	if err != nil {
		app.logger.Error("Error writing to index.html", err)
		return err
	}

	app.logger.Info("index.html generated")
	return nil
}

// NOTE: Pass in *slog.Logger to log errors? Or just return them?
func createXML(path string, name string, atom string) error {
	file, err := os.Create(path + name + ".xml")
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(atom)
	if err != nil {
		return err
	}
	return nil
}
