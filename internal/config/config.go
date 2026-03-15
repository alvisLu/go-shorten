package config

import (
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
	godotenv.Load()

	cfg := &Config{
		HOST:        os.Getenv("HOST"),
		PORT:        os.Getenv("PORT"),
		GIN_MODE:    os.Getenv("GIN_MODE"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
	}

	if cfg.DatabaseURL == "" {
		panic("DATABASE_URL is required")
	}

	return cfg
}
