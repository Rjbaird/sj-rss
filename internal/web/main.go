package web

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/gocolly/colly"
	"github.com/gorilla/feeds"
	"github.com/rjbaird/sj-rss/internal/models"
)

type RecentChapter struct {
	Feed   []*feeds.Item
	Series []models.Series
}

// GetRecentChapters gets the recent chapters from the VIZ website. It returns a feed of the chapters and a slice of series data
func GetRecentChapters() (RecentChapter, error) {
	// Create a new logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	const baseURL = "https://www.viz.com"
	const recentURL = baseURL + "/read/shonenjump/section/free-chapters"
	now := time.Now()
	year := now.Year()
	cutoff := now.AddDate(0, 0, -28)

	feed := []*feeds.Item{}
	series := []models.Series{}

	const seriesURL = "/shonenjump/chapters/"

	// Get recent chapters
	c := colly.NewCollector()

	c.OnRequest(func(r *colly.Request) {
		logger.Info("Visiting: " + r.URL.String())
	})

	c.OnError(func(_ *colly.Response, err error) {
		logger.Error("Error getting recent chapters", err)
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

		if release != "" {
			pubDate, err := time.Parse("January 2, 2006", date)
			if err != nil {
				logger.Error("Error parsing date:", err)
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
				meta := models.Meta{
					Name:       name,
					Handle:     handle,
					LastUpdate: now.Unix(),
					RecentURL:  baseURL + chapterLink,
				}
				series = append(series, models.Series{
					Meta: meta,
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

// GetSeriesData gets the series data from the VIZ website for a given handle. It returns a feed of the chapters
func GetSeriesData(handle string) (*feeds.Feed, error) {
	now := time.Now()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	const baseURL = "https://www.viz.com"
	seriesURL := baseURL + "/shonenjump/chapters/" + handle

	// Create a new feed
	feed := &feeds.Feed{
		Title:       "Weekly Shonen Jump",
		Link:        &feeds.Link{Href: seriesURL},
		Description: "The world's most popular manga!",
		Author:      &feeds.Author{Name: "Shonen Jump | VIZ"},
		Created:     now,
	}

	// Create a new slice of feed items
	results := []*feeds.Item{}

	// Create a new colly collector
	c := colly.NewCollector()

	// Log when visiting a page with the handle
	c.OnRequest(func(r *colly.Request) {
		logger.Info("Visiting: " + r.URL.String())
	})

	// Log any errors
	c.OnError(func(_ *colly.Response, err error) {
		logger.Error("Error getting series data", err)
	})

	// Get the title, description, author, and the first 3 chapters
	c.OnHTML("body", func(e *colly.HTMLElement) {
		title := e.ChildText("h2.type-lg")
		description := e.ChildText("div.line-solid.type-md")
		// hero := e.ChildAttr("img.o_hero-media", "src")

		feed.Title = title
		feed.Description = description

		author := e.ChildText("span.disp-bl--bm")

		// Get the first 3 chapters
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
					logger.Error("Error parsing date:", err)
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

	// Visit the series URL
	c.Visit(seriesURL)

	// Set the feed items
	feed.Items = results

	// Return the completed feed
	return feed, nil
}
