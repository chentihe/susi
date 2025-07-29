package handlers

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/tihe/susi/backend/models"
	"github.com/tihe/susi/backend/services"
)

type PropertyHandler struct {
	Service services.PropertyService
}

func NewPropertyHandler(service services.PropertyService) *PropertyHandler {
	return &PropertyHandler{Service: service}
}

func (h *PropertyHandler) Create(c *gin.Context) {
	var req models.Property
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	if err := h.Service.CreateProperty(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, req)
}

func (h *PropertyHandler) Get(c *gin.Context) {}
func (h *PropertyHandler) Update(c *gin.Context) {}
func (h *PropertyHandler) Delete(c *gin.Context) {}

func RegisterPropertyRoutes(rg *gin.RouterGroup, handler *PropertyHandler) {
	rg.POST("/properties", handler.Create)
	rg.GET("/properties/:id", handler.Get)
	rg.PUT("/properties/:id", handler.Update)
	rg.DELETE("/properties/:id", handler.Delete)
} 