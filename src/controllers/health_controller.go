package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type HealthController struct{}

func NewHealthController() *HealthController {
	return &HealthController{}
}

// HealthCheck handles health check requests
func (c *HealthController) HealthCheck(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": "Service is running",
		"service": "sama-backend-2025",
	})
}

// ReadyCheck handles readiness check requests
func (c *HealthController) ReadyCheck(ctx *gin.Context) {
	// You can add database connectivity checks here
	ctx.JSON(http.StatusOK, gin.H{
		"status":  "ready",
		"message": "Service is ready to handle requests",
	})
}
