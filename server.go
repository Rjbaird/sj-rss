package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/bairrya/sj-rss/config"
	"github.com/go-co-op/gocron"
	"github.com/gocolly/colly"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/favicon"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/gofiber/template/html/v2"
	"github.com/gorilla/feeds"
)

type Series struct {
	Name string
}

func main() {
	// load config
	config, err := config.ENV()

	if err != nil {
		log.Fatal(err)
	}

	// set up cron jobs
	s := gocron.NewScheduler(time.UTC)
	s.Every(1).Day().At("10:00").At("12:00").At("14:00").Do(updateFeeds)

	// set up server
	engine := html.New("./views", ".html")
	server := fiber.New(fiber.Config{Views: engine, ViewsLayout: "layouts/main"})

	// set up no-sql database (redis w/ hashmaps + persistence) for api

	// set up middleware
	server.Use(logger.New())
	server.Use(helmet.New())
	server.Use(cors.New())
	server.Use(recover.New())
	server.Use(favicon.New(favicon.Config{
		File: "./assets/favicon.ico",
		URL:  "/favicon.ico",
	}))

	// set up routes
	server.Get("/", func(c *fiber.Ctx) error {
		return c.Render("index", fiber.Map{})
	})

	rss := server.Group("/rss")
	rss.Use(func(c *fiber.Ctx) error {
		c.Set("Content-Type", "application/xml")
		return c.Next()
	})
	rss.Static("/", "./views/rss")

	api := server.Group("/api")
	api.Use(requestid.New())

	api.Use(limiter.New(limiter.Config{
		Max:        20,
		Expiration: 30 * time.Second,
		// TODO: add 429 json in addition to status
		LimitReached: func(c *fiber.Ctx) error {
			return c.SendStatus(429)
		},
	}))

	api.Use(func(c *fiber.Ctx) error {
		c.Set("Content-Type", "application/json")
		return c.Next()
	})

	// GET /api - api documentation JSON
	// GET /api/manga - get recent updates
	// GET /api/manga/:series - get series JSON

	// start cron job goroutine
	s.StartAsync()
	go updateFeeds()
	// start server goroutine
	log.Fatal(server.Listen(fmt.Sprintf(":%s", config.PORT)))
}

const xmlPath = "views/rss/"

func updateFeeds() {
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
		updateSeriesFeed(handle)
		time.Sleep(3 * time.Second)
	}
}

func updateSeriesFeed(series string) {
	baseURL := "https://www.viz.com/shonenjump/chapters/" + series

	c := colly.NewCollector()

	now := time.Now()
	results := []*feeds.Item{}

	feed := &feeds.Feed{
		Title:       "Weekly Shonen Jump",
		Link:        &feeds.Link{Href: baseURL},
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

	c.Visit(baseURL)
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
