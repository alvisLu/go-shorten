package gateway

import (
	"github.com/alvisLu/go-shorten/internal/config"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Server struct {
	router *gin.Engine
	cfg    *config.Config
}

func NewHttpServer(cfg *config.Config, db *gorm.DB) *Server {
	gin.SetMode(cfg.GIN_MODE)
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	r.SetTrustedProxies(nil)

	registerRoutes(cfg, db, r)

	return &Server{router: r, cfg: cfg}
}

func (s *Server) ListenAndServe() error {
	return s.router.Run(s.cfg.HOST + ":" + s.cfg.PORT)
}
