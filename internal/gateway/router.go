package gateway

import (
	"github.com/alvisLu/go-shorten/internal/config"
	"github.com/alvisLu/go-shorten/internal/health"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func registerRoutes(cfg *config.Config, db *gorm.DB, r *gin.Engine) {
	svc := health.NewService()
	h := health.NewHandler(svc)
	r.GET("/", h.Health)

	registerUrlRoutes(db, r)

	// ws
	registerWsRoutes(cfg, r)
}
