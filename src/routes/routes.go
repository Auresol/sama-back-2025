package routes

import (
	"log"
	"os"
	"sama/sama-backend-2025/src/controllers"
	"sama/sama-backend-2025/src/middlewares"
	"sama/sama-backend-2025/src/services"

	_ "sama/sama-backend-2025/docs"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupRoutes configures all the routes for the application
func SetupRoutes() *gin.Engine {
	router := gin.Default()

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable not set. Please provide a secret key.")
	}
	jwtExpirationMinutes := 60 // Example: Token expires in 60 minutes

	// Initialize services
	userService := services.NewUserService(jwtSecret, jwtExpirationMinutes)

	// Initialize handlers
	uesrController := controllers.NewUserController(userService)

	// Swagger documentation
	// docs.SwaggerInfo.BasePath = "/api/v1"
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Health check routes
	healthController := controllers.NewHealthController()
	router.GET("/health", healthController.HealthCheck)
	router.GET("/ready", healthController.ReadyCheck)

	// Public routes (no authentication required)
	publicRoutes := router.Group("/api/v1")
	{
		publicRoutes.POST("/register", uesrController.RegisterUser)
		publicRoutes.POST("/login", uesrController.Login)
	}

	// Authenticated routes (protected by JWT middleware)
	authRoutes := router.Group("/api/v1")
	authRoutes.Use(middlewares.AuthMiddleware(jwtSecret))
	{
		authRoutes.GET("/me", uesrController.GetMyProfile)
		authRoutes.GET("/users/:id", uesrController.GetUserByID)
		authRoutes.PUT("/users/:id", uesrController.UpdateUserProfile)
		authRoutes.PUT("/users/:id/password", uesrController.UpdateUserPassword) // New endpoint for password change
		authRoutes.DELETE("/users/:id", uesrController.DeleteUser)

		// New routes based on features
		authRoutes.GET("/schools/:school_id/users", uesrController.GetUsersBySchoolID)
		authRoutes.PUT("/schools/:school_id/classrooms", uesrController.UpdateClassroomForSchool)
		// authRoutes.PUT("/users/:student_id/classroom", uesrController.UpdateClassroomForStudent)
		authRoutes.POST("/check-student-email", uesrController.CheckStudentEmailForRegistration)
	}

	return router
}
