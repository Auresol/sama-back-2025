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
	recordService := services.NewRecordService(validate)

	// Initialize handlers
	authController := controllers.NewAuthController(userService, validate)
	userController := controllers.NewUserController(userService, validate)
	schoolController := controllers.NewSchoolController(schoolService, userService, validate)
	activityController := controllers.NewActivityController(activityService, validate)
	recordController := controllers.NewRecordController(recordService)

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
		publicRoutes.POST("/register", authController.RegisterUser)
		publicRoutes.POST("/login", authController.Login)
		publicRoutes.POST("/forgot-password/request", authController.RequestOtp)
		publicRoutes.POST("/forgot-password/validate", authController.ValidateOtp)
	}

	// Authenticated routes (protected by JWT middlewares)
	authRoutes := router.Group("/api/v1")
	authRoutes.Use(middlewares.Authmiddlewares(jwtSecret))
	{
		authRoutes.GET("/me", userController.GetMyProfile)
		authRoutes.GET("/users/:id", userController.GetUserByID)
		authRoutes.PUT("/users/:id", userController.UpdateUserProfile)
		authRoutes.DELETE("/users/:id", userController.DeleteUser)
		authRoutes.GET("/users/activities", userController.GetAssignedActivities)
		authRoutes.GET("/users/records", userController.GetRelatedRecords) // require pagination

		authRoutes.GET("/schools", schoolController.GetAllSchools)
		authRoutes.GET("/school/:id", schoolController.GetSchoolByID)
		authRoutes.POST("/school", schoolController.CreateSchool)
		authRoutes.PUT("/school/:id", schoolController.UpdateSchool)
		authRoutes.DELETE("/school/:id", schoolController.DeleteSchool)
		authRoutes.POST("/school/advance-semester", schoolController.AdvanceSemester)
		authRoutes.POST("/school/revert-semester", schoolController.RevertSemester)
		authRoutes.GET("/school/:id/users", schoolController.GetUsersBySchoolID)

		authRoutes.POST("/activity", activityController.CreateActivity)
		authRoutes.GET("/activities", activityController.GetAllActivities)
		authRoutes.GET("/activity/:id", activityController.GetActivityByID)
		authRoutes.PUT("/activity/:id", activityController.UpdateActivity)
		authRoutes.DELETE("/activity/:id", activityController.DeleteActivity)

		authRoutes.GET("/records", recordController.GetAllRecords)
		authRoutes.GET("/record/:id", recordController.GetRecordByID)
		authRoutes.POST("/record", recordController.CreateRecord)
		authRoutes.PUT("/record/:id", recordController.UpdateRecord)
		authRoutes.DELETE("/record/:id", recordController.DeleteRecord)
		authRoutes.PATCH("/record/:id/send", recordController.SendRecord)
		authRoutes.PATCH("/record/:id/unsend", recordController.UnsendRecord)
		authRoutes.PATCH("/record/:id/approve", recordController.ApproveRecord)
		authRoutes.PATCH("/record/:id/reject", recordController.RejectRecord)
	}

	return router
}
