package handlers

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/tihe/susi/backend/models"
	"github.com/tihe/susi/backend/services"
)

type RoomHandler struct {
	Service services.RoomService
}

func NewRoomHandler(service services.RoomService) *RoomHandler {
	return &RoomHandler{Service: service}
}

func RegisterRoomRoutes(rg *gin.RouterGroup, handler *RoomHandler) {
	rg.POST("/rooms", handler.Create)
	rg.GET("/rooms/:id", handler.Get)
	rg.PUT("/rooms/:id", handler.Update)
	rg.DELETE("/rooms/:id", handler.Delete)
}

func (h *RoomHandler) Create(c *gin.Context) {
	var req models.Room
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	if err := h.Service.CreateRoom(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, req)
} 