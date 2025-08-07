package pkg

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"sama/sama-backend-2025/src/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3Client encapsulates the S3 presigning functionality.
type S3Client struct {
	presignClient *s3.PresignClient
	bucketName    string
	lifetime      time.Duration
}

// NewS3Client creates a new S3Client instance with a default lifetime for presigned URLs.
func NewS3Client(config config.Config) *S3Client {

	cfg, err := awsConfig.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	return &S3Client{
		presignClient: s3.NewPresignClient(s3.NewFromConfig(cfg)),
		bucketName:    config.S3.Bucket,
		lifetime:      time.Duration(config.S3.PreSignedLifeTimeMinutes) * time.Minute,
	}
}

// GetPresignedDownloadURL generates a presigned request for downloading an object.
func (c *S3Client) GetPresignedDownloadURL(ctx context.Context, objectKey string) (*v4.PresignedHTTPRequest, error) {
	request, err := c.presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.bucketName),
		Key:    aws.String(objectKey),
	}, s3.WithPresignExpires(c.lifetime))

	if err != nil {
		log.Printf("failed to generate a presigned download request: %v\n", err)
	}
	return request, err
}

// GetPresignedUploadURL generates a presigned request for uploading an object.
func (c *S3Client) GetPresignedUploadURL(ctx context.Context, objectKey string) (*v4.PresignedHTTPRequest, error) {
	request, err := c.presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(c.bucketName),
		Key:    aws.String(objectKey),
	}, s3.WithPresignExpires(c.lifetime))

	if err != nil {
		log.Printf("failed to generate a presigned upload request: %v\n", err)
	}
	return request, err
}

// GetPresignedDeleteURL generates a presigned request for deleting an object.
func (c *S3Client) GetPresignedDeleteURL(ctx context.Context, objectKey string) (*v4.PresignedHTTPRequest, error) {
	request, err := c.presignClient.PresignDeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(c.bucketName),
		Key:    aws.String(objectKey),
	}, s3.WithPresignExpires(c.lifetime))

	if err != nil {
		log.Printf("failed to generate a presigned delete request: %v\n", err)
		return nil, err
	}

	return request, err
}

func (c *S3Client) PresignPostObject(ctx context.Context, objectKey string) (*s3.PresignedPostRequest, error) {
	// policy := `[["starts-with", "$Content-Type", "image/"]]`
	policy := `[]`
	var policyJson []interface{}
	err := json.Unmarshal([]byte(policy), &policyJson)
	if err != nil {
		return nil, err
	}

	request, err := c.presignClient.PresignPostObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(c.bucketName),
		Key:    aws.String(objectKey),
	}, func(options *s3.PresignPostOptions) {
		options.Expires = c.lifetime
		options.Conditions = policyJson
	})
	if err != nil {
		log.Printf("failed to generate a presigned post request: %v\n", err)
		return nil, err
	}
	return request, nil
}
