package gateway

import (
	"github.com/alvisLu/go-shorten/internal/shorturl"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func registerUrlRoutes(db *gorm.DB, r *gin.Engine) {
	repo := shorturl.NewRepository(db)
	svc := shorturl.NewService(repo)
	h := shorturl.NewHandler(svc)

	r.POST("/shorten", h.CreateShortUrl)
	r.GET("/:code", h.GetOriginalURL)
}
