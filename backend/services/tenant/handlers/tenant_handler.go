package handlers

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/tihe/susi/backend/models"
	"github.com/tihe/susi/backend/services"
)

type TenantHandler struct {
	Service services.TenantService
}

func NewTenantHandler(service services.TenantService) *TenantHandler {
	return &TenantHandler{Service: service}
}

func RegisterTenantRoutes(rg *gin.RouterGroup, handler *TenantHandler) {
	rg.POST("/tenants", handler.Create)
	rg.GET("/tenants/:id", handler.Get)
	rg.PUT("/tenants/:id", handler.Update)
	rg.DELETE("/tenants/:id", handler.Delete)
}

func (h *TenantHandler) Create(c *gin.Context) {
	var req models.Tenant
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	if err := h.Service.CreateTenant(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, req)
} 