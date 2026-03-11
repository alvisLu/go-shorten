package main

import (
	"github.com/alvisLu/go-short/api/handler"
	"github.com/alvisLu/go-short/internal/config"
	"github.com/alvisLu/go-short/internal/service"
	"github.com/gin-gonic/gin"
)

func main() {
	config := config.LoadConfig()
	service := service.NewService()
	handler := handler.NewHandler(service)

	gin := gin.Default()
	gin.GET("/", handler.Health)
	gin.Run(config.HOST + ":" + config.PORT)
}
