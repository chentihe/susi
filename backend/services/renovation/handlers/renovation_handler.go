package handlers

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/tihe/susi/backend/models"
	"github.com/tihe/susi/backend/services"
)

type RenovationHandler struct {
	Service services.RenovationService
}

func NewRenovationHandler(service services.RenovationService) *RenovationHandler {
	return &RenovationHandler{Service: service}
}

func RegisterRenovationRoutes(rg *gin.RouterGroup, handler *RenovationHandler) {
	rg.POST("/renovations", handler.Create)
	rg.GET("/renovations/:id", handler.Get)
	rg.PUT("/renovations/:id", handler.Update)
	rg.DELETE("/renovations/:id", handler.Delete)
}

func (h *RenovationHandler) Create(c *gin.Context) {
	var req models.Renovation
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	if err := h.Service.CreateRenovation(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, req)
} 