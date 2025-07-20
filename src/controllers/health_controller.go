package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type HealthController struct{}

func NewHealthController() *HealthController {
	return &HealthController{}
}

// HealthCheck return if database is healty
// @Summary Check health
// @Description return if database is healty
// @Tags Health
// @Produce json
// @Success 200 {string} string "ok"
// @Router /health [get]
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
