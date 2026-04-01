package main

import (
	"log"

	"github.com/alvisLu/go-shorten/internal/config"
	"github.com/alvisLu/go-shorten/internal/db"
	"github.com/alvisLu/go-shorten/internal/gateway"
	"github.com/alvisLu/go-shorten/internal/stt"
	"github.com/alvisLu/go-shorten/internal/stt/whisper"
)

func main() {
	cfg := config.LoadConfig()

	database, err := db.NewDB(cfg.DatabaseURL)
	if err != nil {
		panic("failed to connect database: " + err.Error())
	}

	var pipeline *stt.Pipeline
	if cfg.WhisperModelPath != "" {
		w, err := whisper.LoadModel(cfg.WhisperModelPath)
		if err != nil {
			log.Fatalf("failed to load whisper model: %v", err)
		}
		defer w.Close()
		pipeline = stt.NewPipeline(w, nil)
	} else {
		log.Println("WHISPER_MODEL_PATH not set, STT disabled")
		pipeline = stt.NewPipeline(nil, nil)
	}

	server := gateway.NewHttpServer(cfg, database, pipeline)
	if err := server.ListenAndServe(); err != nil {
		panic("failed to start server: " + err.Error())
	}
}
