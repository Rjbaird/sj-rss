package web

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/bairrya/sj-rss/db"
	"github.com/gocolly/colly"
	"github.com/gorilla/feeds"
)

type RecentChapter struct {
	Feed   []*feeds.Item
	Series []db.Series
}

func GetRecentChapters() (RecentChapter, error) {
	const baseURL = "https://www.viz.com"
	const recentURL = baseURL + "/read/shonenjump/section/free-chapters"

	now := time.Now()
	year := now.Year()
	cutoff := now.AddDate(0, 0, -21)

	feed := []*feeds.Item{}
	series := []db.Series{}

	const seriesURL = "/shonenjump/chapters/"

	// Get recent chapters
	c := colly.NewCollector()

	c.OnRequest(func(r *colly.Request) {
		log.Println("Visiting", r.URL)
	})

	c.OnError(func(_ *colly.Response, err error) {
		log.Panic("Something went wrong getting recent chapters:", err)
	})

	c.OnHTML(".o_sortable", func(e *colly.HTMLElement) {
		name := e.ChildText("div.type-center")
		recent := strings.Split(e.ChildText("span"), "\n")[0]
		chapterNumber := strings.Replace(recent, "Latest: ", "", -1)
		chapterLink := e.ChildAttr("a.o_inner-link", "href")
		handle := strings.Replace(e.ChildAttr("a.o_chapters-link", "href"), seriesURL, "", -1)
		title := fmt.Sprintf("%s - %s", name, chapterNumber)
		description := fmt.Sprintf("Read %s - %s", name, chapterNumber)
		release := e.ChildText("span.type-bs--sm")
		date := fmt.Sprintf("%s, %d", release, year)
		image := e.ChildAttr("img", "src")

		if release != "" {
			pubDate, err := time.Parse("January 2, 2006", date)
			if err != nil {
				log.Println("Error parsing date:", err)
			}

			manga := &feeds.Item{
				Title:       title,
				Description: description,
				Link:        &feeds.Link{Href: baseURL + chapterLink},
				Created:     pubDate,
			}
			// filter to most recent 2 weeks
			if pubDate.After(cutoff) {
				feed = append(feed, manga)
				series = append(series, db.Series{
					Name:       name,
					Handle:     handle,
					URL:        baseURL + chapterLink,
					LastUpdate: now.Unix(),
					Image:      image,
				})
			}
		}
	})

	c.Visit(recentURL)

	return RecentChapter{
		Series: series,
		Feed:   feed,
	}, nil
}

func GetSeriesData(handle string) (*feeds.Feed, error) {
	now := time.Now()

	const baseURL = "https://www.viz.com"
	seriesURL := baseURL + "/shonenjump/chapters/" + handle

	feed := &feeds.Feed{
		Title:       "Weekly Shonen Jump",
		Link:        &feeds.Link{Href: seriesURL},
		Description: "The world's most popular manga!",
		Author:      &feeds.Author{Name: "Shonen Jump | VIZ"},
		Created:     now,
	}

	results := []*feeds.Item{}

	c := colly.NewCollector()

	c.OnRequest(func(r *colly.Request) {
		log.Println("Visiting", handle)
	})

	c.OnError(func(_ *colly.Response, err error) {
		log.Panic("Something went wrong getting recent chapters:", err)
	})

	c.OnHTML("body", func(e *colly.HTMLElement) {
		title := e.ChildText("h2.type-lg")
		description := e.ChildText("div.line-solid.type-md")

		feed.Title = title
		feed.Description = description

		author := e.ChildText("span.disp-bl--bm")

		// chapters
		e.ForEach("div.o_sortable", func(i int, el *colly.HTMLElement) {
			if i > 3 {
				return
			}

			chapter_string := el.ChildText("td.ch-num-list-spacing")
			chapter_link := el.ChildAttr("a.o_chapter-container", "href")
			chapter_date := el.ChildText("div.style-italic")

			if chapter_string != "" && !strings.Contains(chapter_link, "join to read") {
				pubDate, err := time.Parse("January 2, 2006", chapter_date)
				if err != nil {
					log.Println("Error parsing date:", err)
				}

				chapter := &feeds.Item{
					Title:       fmt.Sprintf("%s - %s", title, chapter_string),
					Description: fmt.Sprintf("Read %s - %s", title, chapter_string),
					Link:        &feeds.Link{Href: baseURL + chapter_link},
					Author:      &feeds.Author{Name: author},
					Created:     pubDate,
				}

				results = append(results, chapter)
			}
		})
	})

	c.Visit(seriesURL)

	feed.Items = results

	return feed, nil
}
