package controllers

import (
	"net/http"

	"sama/sama-backend-2025/src/middlewares"
	"sama/sama-backend-2025/src/services"

	"github.com/gin-gonic/gin"
)

// ImageController manages HTTP requests related to image handling.
type ImageController struct {
	imageService *services.ImageService
}

// NewImageController creates a new ImageController.
func NewImageController(imageService *services.ImageService) *ImageController {
	return &ImageController{
		imageService: imageService,
	}
}

// UploadRequest represents the request body for an image upload.
type UploadRequest struct {
	FileExtension string `json:"file_extension" binding:"required,oneof=jpg jpeg png gif webp" example:"png"`
}

// UploadResponse represents the response for a successful upload request.
type UploadResponse struct {
	URL    string            `json:"url" example:"https://your-s3-bucket.s3.amazonaws.com"`
	Fields map[string]string `json:"fields"`
}

// DownloadRequest represents the request body for an image download.
type DownloadRequest struct {
	ObjectKey string `json:"object_key" binding:"required" example:"user_id/e3c4e512-421e-45a2-921d-a9f3c7e0c4f8.png"`
}

// DownloadResponse represents the response for a successful download request.
type DownloadResponse struct {
	URL string `json:"url" example:"https://your-s3-bucket.s3.amazonaws.com/user_id/image.png?X-Amz-..."`
}

// RequestUploadPresignedURL handles the request for an image upload presigned URL.
// @Summary Get presigned URL for image upload
// @Description Generates a presigned URL and form fields for a direct, secure image upload to S3.
// @Tags Image
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param upload body UploadRequest true "File extension of the image to be uploaded"
// @Success 200 {object} UploadResponse "Presigned URL and form data for upload"
// @Failure 400 {object} ErrorResponse "Invalid request payload or validation error"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /images/upload-url [post]
func (h *ImageController) RequestUploadPresignedURL(c *gin.Context) {
	claims, ok := middlewares.GetUserClaimsFromContext(c)
	if !ok {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "User claims not found in context"})
		return
	}

	userID := claims.UserID

	var req UploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid request payload: " + err.Error()})
		return
	}

	presignedPostRequest, err := h.imageService.RequestUploadPresignedURL(c.Request.Context(), userID, req.FileExtension)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Failed to get presigned URL: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, UploadResponse{
		URL:    presignedPostRequest.URL,
		Fields: presignedPostRequest.Values,
	})
}

// RequestDownloadPresignedURL handles the request for an image download presigned URL.
// @Summary Get presigned URL for image download
// @Description Generates a temporary URL for a direct, secure download of an image from S3.
// @Tags Image
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param download body DownloadRequest true "Object key of the image to be downloaded"
// @Success 200 {object} DownloadResponse "Presigned URL for download"
// @Failure 400 {object} ErrorResponse "Invalid request payload or validation error"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /images/download-url [post]
func (h *ImageController) RequestDownloadPresignedURL(c *gin.Context) {
	var req DownloadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid request payload: " + err.Error()})
		return
	}

	presignedHTTPRequest, err := h.imageService.RequestDownloadPresignedURL(c.Request.Context(), req.ObjectKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Failed to get presigned download URL: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, DownloadResponse{
		URL: presignedHTTPRequest.URL,
	})
}
