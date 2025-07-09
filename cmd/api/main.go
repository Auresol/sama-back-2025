package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"sama/sama-backend-2025/src/config"
	"sama/sama-backend-2025/src/pkg/logger"
	"sama/sama-backend-2025/src/repository"
	"sama/sama-backend-2025/src/routes"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// @title           Sama Backend API
// @version         1.0
// @description     A modern Go backend API for Sama application
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Initialize logger
	if err := logger.InitLogger(cfg.Logging.Level, cfg.Logging.File); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	logger.GetLogger().Info("Starting Sama Backend 2025",
		zap.String("version", "1.0.0"),
		zap.String("environment", cfg.Server.Mode),
	)

	// Set Gin mode
	gin.SetMode(cfg.Server.Mode)

	// Initialize database
	if err := repository.InitDatabase(cfg); err != nil {
		logger.GetLogger().Fatal("Failed to initialize database", zap.Error(err))
	}

	// Setup routes
	router := routes.SetupRoutes()

	// Add CORS middleware
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})

	// Add logging middleware
	router.Use(func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Log request details
		latency := time.Since(start)
		status := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method

		logger.GetLogger().Info("HTTP Request",
			zap.String("method", method),
			zap.String("path", path),
			zap.String("raw_query", raw),
			zap.String("client_ip", clientIP),
			zap.Int("status", status),
			zap.Duration("latency", latency),
			zap.String("user_agent", c.Request.UserAgent()),
		)
	})

	// Start server
	serverAddr := fmt.Sprintf(":%s", cfg.Server.Port)
	logger.GetLogger().Info("Server starting",
		zap.String("port", cfg.Server.Port),
		zap.String("address", serverAddr),
	)

	if err := router.Run(serverAddr); err != nil {
		logger.GetLogger().Fatal("Failed to start server", zap.Error(err))
	}
}
