package jobs

import (
	"log"
	"os"
	"time"

	"github.com/bairrya/sj-rss/db"
	"github.com/bairrya/sj-rss/web"
	"github.com/gorilla/feeds"
)

const xmlPath = "views/rss/"

func UpdateFeeds() {
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
		log.Fatal(err)
	}

	feed.Items = results.Feed

	atom, err := feed.ToAtom()
	if err != nil {
		log.Println("Error converting series feed to atom:", err)
	}

	// create main feed
	createXML(xmlPath, "main", atom)

	// use handles to scrape series
	for _, manga := range results.Series {

		mFeed, err := web.GetSeriesData(manga.Handle)
		if err != nil {
			log.Println("Error updating series feed:", err)
		}

		mAtom, err := mFeed.ToAtom()
		if err != nil {
			log.Println("Error converting series feed to atom:", err)
		}

		// create series feed
		log.Println("Creating feed for", manga.Handle)
		createXML(xmlPath, manga.Handle, mAtom)
		err = db.SetSeries(manga)
		if err != nil {
			log.Println("Error setting series:", err)
		}
		time.Sleep(3 * time.Second)
	}

	log.Println("done")
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
