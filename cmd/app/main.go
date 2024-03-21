package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"github.com/rjbaird/sj-rss/internal/models"
	"github.com/robfig/cron/v3"
)

type config struct {
	port       string
	staticPath string
	rssPath    string
}

type application struct {
	config   *config
	logger   *slog.Logger
	series   *models.SeriesModel
	router   *chi.Mux
	schedule *cron.Cron
}

func main() {
	// Create a new logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// Run the server
	if err := run(); err != nil {
		logger.Error("Error running server", err)
		os.Exit(1)
	}
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

	// Initialize a new server
	port := os.Getenv("PORT")
	if port == "" {
		logger.Info("PORT not set, defaulting to 3000")
		port = "3000"
	}

	config := &config{
		port:       ":" + port,
		staticPath: "./views/static/",
		rssPath:    "./views/rss/",
	}

	router := chi.NewRouter()

	application := &application{logger: logger, config: config, series: &models.SeriesModel{DB: client}, router: router, schedule: cron.New(cron.WithLocation(time.UTC))}

	// Create a new router with middleware
	application.router.Use(application.logRequest)
	application.router.Use(middleware.Recoverer)

	// Set up the heartbeat route
	application.router.Use(middleware.Heartbeat("/ping"))

	// Handle 404 errors
	application.router.NotFound(application.notFound404)

	// Handle static assets
	staticFileServer := http.FileServer(http.Dir(application.config.staticPath))
	application.router.Handle("/static/*", http.StripPrefix("/static", staticFileServer))

	// Handle rss files
	rssFileServer := http.FileServer(http.Dir(application.config.rssPath))
	application.router.Handle("/rss/*", http.StripPrefix("/rss", rssFileServer))

	// Create series feeds
	err = application.generateSeriesFeeds()
	if err != nil {
		logger.Error("Error generating series feeds", err)
		return err
	}

	// Create index.html
	err = application.generateIndex()
	if err != nil {
		logger.Error("Error generating index.html", err)
		return err
	}

	// Define application routes
	application.router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "views/index.html")
	})

	// Start the cron jobs
	application.schedule.AddFunc("0 10,12,14 * * *", func() {
		application.generateSeriesFeeds()
		application.generateIndex()
	})
	application.schedule.Start()

	// Start the server
	logger.Info("Starting server on " + application.config.port)
	err = http.ListenAndServe(application.config.port, application.router)
	return err
}
