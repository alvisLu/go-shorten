package router

import (
	"github.com/alvisLu/go-shorten/api/handler"
	"github.com/alvisLu/go-shorten/internal/repository"
	"github.com/alvisLu/go-shorten/internal/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func NewUrlRoute(db *gorm.DB, gin *gin.RouterGroup) {

	ur := repository.NewURLRepository(db)
	svc := service.NewURLService(ur)
	h := handler.NewURLHandler(svc)

	gin.POST("/shorten", h.CreateShortUrl)
	gin.GET("/:code", h.GetOriginalURL)
}
