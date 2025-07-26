package routes

import (
	"log"
	"os"
	"sama/sama-backend-2025/src/controllers"
	"sama/sama-backend-2025/src/middlewares"
	"sama/sama-backend-2025/src/services"
	"sama/sama-backend-2025/src/utils"

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

	validate := utils.Validate

	// Initialize services
	userService := services.NewUserService(jwtSecret, jwtExpirationMinutes, validate)
	schoolService := services.NewSchoolService(validate)
	activityService := services.NewActivityService(validate)

	// Initialize handlers
	userController := controllers.NewUserController(userService, validate)
	schoolController := controllers.NewSchoolController(schoolService, validate)
	activityController := controllers.NewActivityController(activityService, validate)

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
		publicRoutes.POST("/register", userController.RegisterUser)
		publicRoutes.POST("/login", userController.Login)

		publicRoutes.POST("/schools", schoolController.CreateSchool)
		publicRoutes.GET("/schools", schoolController.GetAllSchools)
		publicRoutes.PUT("/schools/:id", schoolController.UpdateSchool)
		publicRoutes.DELETE("/schools/:id", schoolController.DeleteSchool)
	}

	// Authenticated routes (protected by JWT middlewares)
	authRoutes := router.Group("/api/v1")
	authRoutes.Use(middlewares.Authmiddlewares(jwtSecret))
	{
		authRoutes.GET("/me", userController.GetMyProfile)
		authRoutes.GET("/users/:id", userController.GetUserByID)
		authRoutes.PUT("/users/:id", userController.UpdateUserProfile)
		authRoutes.PUT("/users/:id/password", userController.UpdateUserPassword) // New endpoint for password change
		authRoutes.DELETE("/users/:id", userController.DeleteUser)

		// User-related routes with school context
		// authRoutes.GET("/schools/:school_id/users", userController.GetUsersBySchoolID)
		// authRoutes.PUT("/schools/:school_id/classrooms", userController.UpdateClassroomForSchool)
		// authRoutes.PUT("/users/:student_id/classroom", userController.UpdateClassroomForStudent) // Uncommented
		authRoutes.POST("/check-student-email", userController.CheckStudentEmailForRegistration)

		// Activity Routes (newly added)
		authRoutes.POST("/activities", activityController.CreateActivity)
		authRoutes.GET("/activities", activityController.GetAllActivities)
		authRoutes.GET("/activities/:id", activityController.GetActivityByID)
		authRoutes.PUT("/activities/:id", activityController.UpdateActivity)
		authRoutes.DELETE("/activities/:id", activityController.DeleteActivity)

		authRoutes.GET("/schools/:id", schoolController.GetSchoolByID)
	}

	return router
}
