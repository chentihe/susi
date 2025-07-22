package handlers

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/tihe/susi/backend/models"
	"github.com/tihe/susi/backend/services"
)

type ApartmentHandler struct {
	Service services.ApartmentService
}

func NewApartmentHandler(service services.ApartmentService) *ApartmentHandler {
	return &ApartmentHandler{Service: service}
}

func RegisterApartmentRoutes(rg *gin.RouterGroup, handler *ApartmentHandler) {
    rg.POST("/apartments", handler.Create)
    rg.GET("/apartments/:id", handler.Get)
    rg.PUT("/apartments/:id", handler.Update)
    rg.DELETE("/apartments/:id", handler.Delete)
}

func (h *ApartmentHandler) Create(c *gin.Context) {
	var req models.Apartment
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	if err := h.Service.CreateApartment(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, req)
} 