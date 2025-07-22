package handlers

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/tihe/susi/backend/models"
	"github.com/tihe/susi/backend/services"
)

type AdminHandler struct {
	Service services.AdminService
}

func NewAdminHandler(service services.AdminService) *AdminHandler {
	return &AdminHandler{Service: service}
}

func RegisterAdminRoutes(rg *gin.RouterGroup, handler *AdminHandler) {
	rg.POST("/admins", handler.Create)
	rg.GET("/admins/:id", handler.Get)
	rg.PUT("/admins/:id", handler.Update)
	rg.DELETE("/admins/:id", handler.Delete)
}

func (h *AdminHandler) Create(c *gin.Context) {
	var req models.Admin
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	if err := h.Service.CreateAdmin(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, req)
}