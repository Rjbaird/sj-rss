package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-co-op/gocron"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"github.com/rjbaird/sj-rss/internal/models"
)

type config struct {
	port       string
	staticPath string
	rssPath    string
}

type server struct {
	config *config
	logger *slog.Logger
	series *models.SeriesModel
	jobs   *gocron.Scheduler
	router *chi.Mux
}

func main() {
	// Run the server
	err := run()
	// Log any errors and exit
	log.Fatal(err)
}

func run() error {
	godotenv.Load(".env")

	// Create a new logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		// AddSource: true,
	}))

	// Get the redis url from the environment
	redisURL := os.Getenv("REDIS_URL")
	options, err := redis.ParseURL(redisURL)
	if err != nil {
		logger.Error("Error parsing redis url", err)
		return err
	}

	// Create a new redis client connection
	client := redis.NewClient(options)
	defer client.Close()

	// set up the cron jobs scheduler
	jobs := gocron.NewScheduler(time.UTC)

	// Initialize a new server
	port := os.Getenv("PORT")
	if port == "" {
		logger.Info("PORT not set, defaulting to 3000")
		port = "3000"
	}
	r := chi.NewRouter()

	server := &server{logger: logger, config: &config{
		port:       ":" + port,
		staticPath: "./views/static/",
		rssPath:    "./views/rss/",
	}, series: &models.SeriesModel{DB: client}, jobs: jobs, router: r}

	// Create a new router with middleware
	server.router.Use(server.logRequest)
	server.router.Use(middleware.Recoverer)

	// Set up the heartbeat route
	server.router.Use(middleware.Heartbeat("/ping"))

	// Handle 404 errors
	server.router.NotFound(server.notFound404)

	// Handle static assets
	staticFileServer := http.FileServer(http.Dir(server.config.staticPath))
	server.router.Handle("/assets/*", http.StripPrefix("/assets", staticFileServer))

	// Handle rss files
	rssFileServer := http.FileServer(http.Dir(server.config.rssPath))
	server.router.Handle("/rss/*", http.StripPrefix("/rss", rssFileServer))
	
	// Create series feeds
	err = server.generateSeriesFeeds()
	if err != nil {
		logger.Error("Error generating series feeds", err)
		return err
	}

	// Create index.html
	err = server.generateIndex()
	if err != nil {
		logger.Error("Error generating index.html", err)
		return err
	}

	// Define application routes
	server.router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "views/index.html")
	})

	// Start the cron jobs
	jobs.Day().At("10:00;12:00;14:00").Do(func() {
		server.generateSeriesFeeds()
		server.generateIndex()
	})
	server.jobs.StartAsync()

	// Start the server
	logger.Info("Starting server on " + server.config.port)
	err = http.ListenAndServe(server.config.port, server.router)
	return err
}
