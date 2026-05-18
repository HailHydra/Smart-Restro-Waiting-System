package httpapi

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"smartrestro-backend/internal/queue"
)

func RegisterMiddleware(router *gin.Engine) {
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})
}

func RegisterRoutes(router *gin.Engine, service *queue.Service) {
	handler := &Handler{service: service}

	router.GET("/health", handler.health)

	api := router.Group("/api/v1")
	{
		api.GET("/restaurants", handler.listRestaurants)
		api.GET("/restaurants/nearby", handler.listNearbyRestaurants)
		api.GET("/restaurants/:id", handler.getRestaurant)
		api.GET("/restaurants/:id/queue", handler.getQueue)
		api.GET("/restaurants/:id/queue-summary", handler.getQueueSummary)
		api.POST("/restaurants/:id/queue/join", handler.joinQueue)
		api.GET("/queue/:id", handler.getQueueEntry)
		api.PATCH("/queue/:id/status", handler.transitionQueueEntry)
	}
}
