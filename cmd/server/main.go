package main

import (
	"github.com/alvisLu/go-short/api/router"
	"github.com/alvisLu/go-short/internal/config"
	"github.com/gin-gonic/gin"
)

func main() {
	config := config.LoadConfig()

	gin := gin.New()
	gin.SetTrustedProxies(nil)

	router.Start(gin)

	gin.Run(config.HOST + ":" + config.PORT)
}
