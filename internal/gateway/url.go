package router

import (
	"github.com/alvisLu/go-shorten/internal/shorturl"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func NewUrlRoute(db *gorm.DB, gin *gin.RouterGroup) {
	repo := shorturl.NewRepository(db)
	svc := shorturl.NewService(repo)
	h := shorturl.NewHandler(svc)

	gin.POST("/shorten", h.CreateShortUrl)
	gin.GET("/:code", h.GetOriginalURL)
}
