package httpapi

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"smartrestro-backend/internal/queue"
)

type Handler struct {
	service *queue.Service
}

func (h *Handler) health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"store":  "memory",
	})
}

func (h *Handler) listRestaurants(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"data": h.service.ListRestaurants()})
}

func (h *Handler) listNearbyRestaurants(c *gin.Context) {
	lat, err := strconv.ParseFloat(c.Query("lat"), 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid lat"})
		return
	}

	lng, err := strconv.ParseFloat(c.Query("lng"), 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid lng"})
		return
	}

	radiusKm := 2.0
	if value := c.Query("radius"); value != "" {
		parsed, parseErr := strconv.ParseFloat(value, 64)
		if parseErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid radius"})
			return
		}
		radiusKm = parsed
	}

	c.JSON(http.StatusOK, gin.H{"data": h.service.ListNearbyRestaurants(lat, lng, radiusKm)})
}

func (h *Handler) getRestaurant(c *gin.Context) {
	restaurant, ok := h.service.GetRestaurant(c.Param("id"))
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "restaurant not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": restaurant})
}

func (h *Handler) getQueue(c *gin.Context) {
	restaurantID := c.Param("id")
	if _, ok := h.service.GetRestaurant(restaurantID); !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "restaurant not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": h.service.GetQueue(restaurantID)})
}

func (h *Handler) getQueueSummary(c *gin.Context) {
	summary, err := h.service.GetQueueSummary(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": summary})
}

func (h *Handler) joinQueue(c *gin.Context) {
	var req queue.JoinQueueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	entry, err := h.service.JoinQueue(c.Param("id"), req)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": entry})
}

func (h *Handler) getQueueEntry(c *gin.Context) {
	entry, events, ok := h.service.GetQueueEntry(c.Param("id"))
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "queue entry not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"entry":  entry,
			"events": events,
		},
	})
}

func (h *Handler) transitionQueueEntry(c *gin.Context) {
	var req queue.TransitionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	entry, err := h.service.TransitionQueueEntry(c.Param("id"), req.Status, "")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": entry})
}
