package router

import (
	"github.com/alvisLu/go-shorten/api/handler"
	"github.com/alvisLu/go-shorten/internal/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func Start(db *gorm.DB, gin *gin.Engine) {
	publicRouter := gin.Group("")
	NewHealthRoute(publicRouter)
	NewUrlRoute(db, publicRouter)
}

func NewHealthRoute(gin *gin.RouterGroup) {
	svc := service.NewService()
	h := handler.NewHandler(svc)

	gin.GET("/", h.Health)
}
