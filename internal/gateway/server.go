package gateway

import (
	"cmp"

	"github.com/alvisLu/go-shorten/internal/config"
	"github.com/alvisLu/go-shorten/internal/stt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Server struct {
	router *gin.Engine
	cfg    *config.Config
}

func NewHttpServer(cfg *config.Config, db *gorm.DB, pipeline *stt.Pipeline) *Server {
	mode := cmp.Or(cfg.GIN_MODE, gin.DebugMode)
	gin.SetMode(mode)
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	r.SetTrustedProxies(nil)

	registerRoutes(cfg, db, r, pipeline)

	return &Server{router: r, cfg: cfg}
}

func (s *Server) ListenAndServe() error {
	return s.router.Run(s.cfg.HOST + ":" + s.cfg.PORT)
}
