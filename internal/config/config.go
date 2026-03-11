package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	HOST string
	PORT string
}

func LoadConfig() *Config {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	return &Config{
		HOST: os.Getenv("HOST"),
		PORT: os.Getenv("PORT"),
	}
}
