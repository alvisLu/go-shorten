package main

import (
	"github.com/alvisLu/go-shorten/internal/config"
	"github.com/alvisLu/go-shorten/internal/db"
	"github.com/alvisLu/go-shorten/internal/gateway"
)

func main() {
	cfg := config.LoadConfig()

	database, err := db.NewDB(cfg.DatabaseURL)
	if err != nil {
		panic("failed to connect database: " + err.Error())
	}

	server := gateway.NewHttpServer(cfg, database)
	if err := server.ListenAndServe(); err != nil {
		panic("failed to start server: " + err.Error())
	}
}
