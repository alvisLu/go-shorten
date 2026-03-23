package gateway

import (
	"cmp"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	srv := &http.Server{
		Addr:    s.cfg.HOST + ":" + s.cfg.PORT,
		Handler: s.router,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("server error: %v", err)
		}
	}()

	<-quit
	log.Println("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return srv.Shutdown(ctx)
}
