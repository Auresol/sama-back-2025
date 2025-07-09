package routes

import (
	"sama/sama-backend-2025/src/controllers"

	"github.com/gin-gonic/gin"
	_ "sama/sama-backend-2025/docs"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupRoutes configures all the routes for the application
func SetupRoutes() *gin.Engine {
	router := gin.Default()

	// Swagger documentation
	// docs.SwaggerInfo.BasePath = "/api/v1"
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Health check routes
	healthController := controllers.NewHealthController()
	router.GET("/health", healthController.HealthCheck)
	router.GET("/ready", healthController.ReadyCheck)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// User routes
		userController := controllers.NewUserController()
		users := v1.Group("/users")
		{
			users.POST("/register", userController.CreateUser)
			users.POST("/login", userController.Login)
			users.GET("/", userController.GetAllUsers)
			users.GET("/:id", userController.GetUser)
			users.PUT("/:id", userController.UpdateUser)
			users.DELETE("/:id", userController.DeleteUser)
		}
	}

	return router
}
