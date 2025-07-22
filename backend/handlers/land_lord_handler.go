package handlers

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/tihe/susi/backend/models"
	"github.com/tihe/susi/backend/services"
)

type LandLordHandler struct {
	Service services.LandLordService
}

func NewLandLordHandler(service services.LandLordService) *LandLordHandler {
	return &LandLordHandler{Service: service}
}

func RegisterLandLordRoutes(rg *gin.RouterGroup, handler *LandLordHandler) {
	rg.POST("/landlords", handler.Create)
	rg.GET("/landlords/:id", handler.Get)
	rg.PUT("/landlords/:id", handler.Update)
	rg.DELETE("/landlords/:id", handler.Delete)
}

func (h *LandLordHandler) Create(c *gin.Context) {
	var req models.LandLord
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	if err := h.Service.CreateLandLord(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, req)
} 