package handlers

import (
	"net/http"
	"time"
	"github.com/gin-gonic/gin"
	"github.com/tihe/susi/backend/app"
	"github.com/tihe/susi/backend/events"
)

func Register(ctx *app.AppContext, c *gin.Context) {
	// Example: publish an ApartmentCreated event (replace with real logic)
	event := events.Event{
		Type:      events.EventApartmentCreated,
		Payload:   map[string]interface{}{ "example": "apartment created" },
		Timestamp: time.Now(),
	}
	if ctx.KafkaProducer != nil {
		err := ctx.KafkaProducer.Publish(event)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to publish event"})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"message": "Register endpoint (event published)"})
}

func Login(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Login endpoint (mock)"})
} 