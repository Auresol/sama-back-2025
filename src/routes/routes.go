package routes

import (
	"sama/sama-backend-2025/src/config"
	"sama/sama-backend-2025/src/controllers"
	"sama/sama-backend-2025/src/middlewares"
	"sama/sama-backend-2025/src/pkg"
	"sama/sama-backend-2025/src/services"
	"sama/sama-backend-2025/src/utils"

	_ "sama/sama-backend-2025/docs"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupRoutes configures all the routes for the application
func SetupRoutes(cfg *config.Config) *gin.Engine {
	router := gin.Default()

	validate := utils.Validate
	s3Client := pkg.NewS3Client(*cfg)

	// Initialize services
	authService := services.NewAuthService(
		cfg.JWT.Secret,
		cfg.JWT.Expiry,
		cfg.RefreshJWT.Secret,
		cfg.RefreshJWT.Expiry,
		validate,
	)
	userService := services.NewUserService(validate)
	schoolService := services.NewSchoolService(validate)
	activityService := services.NewActivityService(validate)
	recordService := services.NewRecordService(validate)
	imageService := services.NewImageService(s3Client)

	// Initialize handlers
	authController := controllers.NewAuthController(authService, validate)
	userController := controllers.NewUserController(userService, activityService, recordService, validate)
	schoolController := controllers.NewSchoolController(schoolService, userService, validate)
	activityController := controllers.NewActivityController(activityService, validate)
	recordController := controllers.NewRecordController(recordService)
	imageController := controllers.NewImageController(imageService)

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
		publicRoutes.POST("/refresh-token", authController.RefreshToken)
		publicRoutes.POST("/forgot-password/request", authController.RequestOtp)
		publicRoutes.POST("/forgot-password/validate", authController.ValidateOtp)
		publicRoutes.POST("/school", schoolController.CreateSchool)
		publicRoutes.GET("/school", schoolController.GetAllSchools)
	}

	// Authenticated routes (protected by JWT middlewares)
	authRoutes := router.Group("/api/v1")
	authRoutes.Use(middlewares.Authmiddlewares(cfg.JWT.Secret))
	{
		authRoutes.GET("/user/me", userController.GetMyProfile)
		authRoutes.GET("/user/:id", userController.GetUserByID)
		authRoutes.PUT("/user/:id", userController.UpdateUserProfile)
		authRoutes.DELETE("/user/:id", userController.DeleteUser)
		authRoutes.GET("/user/:id/activity", userController.GetAssignedActivities)
		// authRoutes.POST("/user/presigned-url", userController.RequestProfilePresignedURL)

		authRoutes.GET("/school/:id", schoolController.GetSchoolByID)
		authRoutes.PUT("/school/:id", schoolController.UpdateSchool)
		authRoutes.DELETE("/school/:id", schoolController.DeleteSchool)
		authRoutes.POST("/school/advance-semester", schoolController.AdvanceSemester)
		authRoutes.POST("/school/revert-semester", schoolController.RevertSemester)
		authRoutes.GET("/school/:id/user", schoolController.GetUsersBySchoolID)
		authRoutes.GET("/school/:id/statistic", schoolController.GetStatistic)

		authRoutes.POST("/activity", activityController.CreateActivity)
		authRoutes.GET("/activity", activityController.GetAllActivities)
		authRoutes.GET("/activity/:id", activityController.GetActivityByID)
		authRoutes.PUT("/activity/:id", activityController.UpdateActivity)
		authRoutes.DELETE("/activity/:id", activityController.DeleteActivity)

		authRoutes.GET("/record", recordController.GetAllRecords)
		authRoutes.GET("/record/:id", recordController.GetRecordByID)
		authRoutes.POST("/record", recordController.CreateRecord)
		authRoutes.PUT("/record/:id", recordController.UpdateRecord)
		authRoutes.DELETE("/record/:id", recordController.DeleteRecord)
		authRoutes.PATCH("/record/:id/send", recordController.SendRecord)
		authRoutes.PATCH("/record/:id/unsend", recordController.UnsendRecord)
		authRoutes.PATCH("/record/:id/approve", recordController.ApproveRecord)
		authRoutes.PATCH("/record/:id/reject", recordController.RejectRecord)

		authRoutes.POST("/images/download-url", imageController.RequestDownloadPresignedURL)
		authRoutes.POST("/images/upload-url", imageController.RequestUploadPresignedURL)
	}

	return router
}
