package main

import (
	"github.com/alvisLu/go-shorten/api/router"
	"github.com/alvisLu/go-shorten/internal/config"
	"github.com/alvisLu/go-shorten/internal/db"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.LoadConfig()

	database, err := db.NewDB(cfg.DatabaseURL)
	if err != nil {
		panic("failed to connect database")
	}

	gin.SetMode(cfg.GIN_MODE)
	gin := gin.New()
	gin.SetTrustedProxies(nil)

	router.Start(database, gin)

	gin.Run(cfg.HOST + ":" + cfg.PORT)
}
