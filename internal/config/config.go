package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	HOST        string
	PORT        string
	GIN_MODE    string
	DatabaseURL string
}

func LoadConfig() *Config {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	return &Config{
		HOST:        os.Getenv("HOST"),
		PORT:        os.Getenv("PORT"),
		GIN_MODE:    os.Getenv("GIN_MODE"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
	}
}
