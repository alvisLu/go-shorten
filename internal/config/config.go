package config

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	HOST           string
	PORT           string
	GIN_MODE       string
	DatabaseURL    string
	AllowedOrigins []string
}

func LoadConfig() *Config {
	godotenv.Load()

	var allowedOrigins []string
	if raw := os.Getenv("ALLOWED_ORIGINS"); raw != "" {
		allowedOrigins = strings.Split(raw, ",")
	}

	cfg := &Config{
		HOST:           os.Getenv("HOST"),
		PORT:           os.Getenv("PORT"),
		GIN_MODE:       os.Getenv("GIN_MODE"),
		DatabaseURL:    os.Getenv("DATABASE_URL"),
		AllowedOrigins: allowedOrigins,
	}

	if cfg.DatabaseURL == "" {
		panic("DATABASE_URL is required")
	}

	return cfg
}
