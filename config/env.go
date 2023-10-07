package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	PORT      string
	REDIS_URI string
}

func ENV() (*Config, error) {
	godotenv.Load(".env")

	PORT := os.Getenv("PORT")
	if PORT == "" {
		log.Println("no PORT environment variable provided")
		log.Println("Setting PORT to 3000")
		PORT = "3000"
	}

	REDIS_URI := os.Getenv("REDIS_URI")
	if REDIS_URI == "" {
		log.Fatal("You must set your 'REDIS_URI' environment variable. See\n\t https://www.mongodb.com/docs/drivers/go/current/usage-examples/#environment-variable")
	}

	config := Config{PORT: PORT, REDIS_URI: REDIS_URI}

	return &config, nil
}
