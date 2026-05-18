package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"smartrestro-backend/internal/config"
	"smartrestro-backend/internal/httpapi"
	"smartrestro-backend/internal/predictor"
	"smartrestro-backend/internal/queue"
	"smartrestro-backend/internal/store"
)

func main() {
	cfg := config.Load()

	localPredictor := predictor.NewLocalPredictor()
	activePredictor := predictor.Predictor(localPredictor)
	if cfg.MLServiceURL != "" {
		activePredictor = predictor.NewHTTPPredictor(cfg.MLServiceURL, localPredictor)
	}

	var queueStore queue.Store
	if cfg.DatabaseURL != "" {
		postgresStore, err := store.NewPostgresStore(cfg.DatabaseURL)
		if err != nil {
			log.Printf("failed to connect to postgres, falling back to memory store: %v", err)
			queueStore = store.NewMemoryStore()
		} else {
			queueStore = postgresStore
			defer postgresStore.Close()
		}
	} else {
		queueStore = store.NewMemoryStore()
	}

	queueService := queue.NewService(queueStore, activePredictor)

	router := gin.Default()
	httpapi.RegisterMiddleware(router)
	httpapi.RegisterRoutes(router, queueService)

	log.Printf("SmartRestro API listening on :%s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
