package services

import (
	"context"
	"errors"
	"fmt"
	"sama/sama-backend-2025/src/pkg"

	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

// ImageService handles business logic for image uploads.
type ImageService struct {
	s3Client *pkg.S3Client
}

// NewImageService creates a new instance of ImageService.
func NewImageService(
	s3Client *pkg.S3Client,
) *ImageService {
	return &ImageService{
		s3Client: s3Client,
	}
}

// RequestDownloadPresignedURL generates a presigned URL for downloading an object.
// The URL is valid for the duration configured in the S3 client.
func (s *ImageService) RequestDownloadPresignedURL(ctx context.Context, objectKey string) (*v4.PresignedHTTPRequest, error) {
	if objectKey == "" {
		return nil, errors.New("objectKey cannot be empty")
	}

	// Call the S3 client to get the presigned download URL
	request, err := s.s3Client.GetPresignedDownloadURL(ctx, objectKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get presigned download URL from S3 client: %w", err)
	}

	return request, nil
}

// RequestUploadPresignedURL generates a presigned POST URL for a user to upload an image.
// The object key will be formatted as "user_id/uuid.extension".
func (s *ImageService) RequestUploadPresignedURL(ctx context.Context, userID uint, fileExtension string) (*s3.PresignedPostRequest, error) {
	if userID == 0 {
		return nil, errors.New("userID cannot be empty")
	}
	if fileExtension == "" {
		return nil, errors.New("fileExtension cannot be empty")
	}

	// Generate a unique filename using userID and a random UUID
	filename := fmt.Sprintf("%d/%s.%s", userID, uuid.New().String(), fileExtension)

	// Call the S3 client to get the presigned POST URL with a policy for images
	request, err := s.s3Client.PresignPostObject(ctx, filename)
	if err != nil {
		return nil, fmt.Errorf("failed to get presigned URL from S3 client: %w", err)
	}

	return request, nil
}
