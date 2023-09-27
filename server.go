package main

import (
	"fmt"
	"log"
	"time"

	"github.com/bairrya/sj-rss/config"
	"github.com/bairrya/sj-rss/jobs"
	"github.com/go-co-op/gocron"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/favicon"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/gofiber/template/html/v2"
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
	s.Every(1).Day().At("10:00").At("12:00").At("14:00").Do(jobs.UpdateFeeds)

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
	go jobs.UpdateFeeds()
	// start server goroutine
	log.Fatal(server.Listen(fmt.Sprintf(":%s", config.PORT)))
}
