package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthHandler 處理健康檢查請求
type HealthHandler struct{}

// NewHealthHandler 建立新的健康檢查 handler
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// Health 健康檢查端點
func (h *HealthHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
	})
}

// Ready 就緒檢查端點
func (h *HealthHandler) Ready(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ready",
	})
}
