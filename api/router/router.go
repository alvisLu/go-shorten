package router

import (
	"github.com/alvisLu/go-short/api/handler"
	"github.com/alvisLu/go-short/internal/service"
	"github.com/gin-gonic/gin"
)

func Start(gin *gin.Engine) {
	publicRouter := gin.Group("")
	NewHealthRoute(publicRouter)
}

func NewHealthRoute(gin *gin.RouterGroup) {
	svc := service.NewService()
	h := handler.NewHandler(svc)

	gin.GET("/", h.Health)
}
