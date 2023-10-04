package jobs

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gocolly/colly"
	"github.com/gorilla/feeds"
)

const xmlPath = "views/rss/"

func UpdateFeeds() {
	const baseURL = "https://www.viz.com"
	const recentURL = baseURL + "/read/shonenjump/section/free-chapters"
	const seriesURL = "/shonenjump/chapters/"

	now := time.Now()
	year := now.Year()
	cutoff := now.AddDate(0, 0, -14)
	feed := &feeds.Feed{
		Title:       "Weekly Shonen Jump",
		Link:        &feeds.Link{Href: recentURL},
		Description: "The world's most popular manga!",
		Author:      &feeds.Author{Name: "Shonen Jump | VIZ"},
		Created:     now,
		// TODO: add main feed image
		// Image: &feeds.Image{},
	}

	results := []*feeds.Item{}
	handles := []string{}

	c := colly.NewCollector()

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
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

		if release != "" {
			pubDate, err := time.Parse("January 2, 2006", date)
			if err != nil {
				fmt.Println("Error parsing date:", err)
			}

			manga := feeds.Item{
				Title:       title,
				Description: description,
				Link:        &feeds.Link{Href: baseURL + chapterLink},
				Created:     pubDate,
			}
			// filter to most recent 2 weeks
			if pubDate.After(cutoff) {
				results = append(results, &manga)
				handles = append(handles, handle)
			}
		}
	})

	c.Visit(recentURL)

	feed.Items = results

	atom, err := feed.ToAtom()
	if err != nil {
		log.Fatal(err)
	}

	createXML(xmlPath, "main", atom)

	// use handles to scrape series
	for _, handle := range handles {
		UpdateSeriesFeed(handle)
		time.Sleep(3 * time.Second)
	}
}

func UpdateSeriesFeed(series string) {
	const baseURL = "https://www.viz.com"
	seriesURL := baseURL + "/shonenjump/chapters/" + series

	c := colly.NewCollector()

	now := time.Now()
	results := []*feeds.Item{}

	feed := &feeds.Feed{
		Title:       "Weekly Shonen Jump",
		Link:        &feeds.Link{Href: seriesURL},
		Description: "The world's most popular manga!",
		Author:      &feeds.Author{Name: "Shonen Jump | VIZ"},
		Created:     now,
		// TODO: add series image
		// Image: &feeds.Image{},
	}

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
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
					fmt.Println("Error parsing date:", err)
				}

				chapter := feeds.Item{
					Title:       fmt.Sprintf("%s - %s", title, chapter_string),
					Description: fmt.Sprintf("Read %s - %s", title, chapter_string),
					Link:        &feeds.Link{Href: baseURL + chapter_link},
					Author:      &feeds.Author{Name: author},
					Created:     pubDate,
				}

				results = append(results, &chapter)
			}
		})
	})

	c.Visit(seriesURL)
	// TODO: store in redis for api + tracking

	// update rss feed
	feed.Items = results

	atom, err := feed.ToAtom()
	if err != nil {
		log.Fatal(err)
	}
	createXML(xmlPath, series, atom)
}

func createXML(path string, name string, atom string) {
	file, err := os.Create(path + name + ".xml")
	if err != nil {
		log.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	_, err = file.WriteString(atom)
	if err != nil {
		log.Println("Error writing to xml feed:", err)
		return
	}
}

func downloadFile(filepath string, url string) (err error) {

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Writer the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}
