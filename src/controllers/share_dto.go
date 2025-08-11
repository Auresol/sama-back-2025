package controllers

import "sama/sama-backend-2025/src/models"

// ErrorResponse represents a generic error response.
type ErrorResponse struct {
	Message string `json:"message" example:"Error description"`
}

type SuccessfulResponse struct {
	Message string `json:"message" example:"operation success"`
}

// PaginateRecordsResponse represents the response body for retrieve records with paginate
type PaginateRecordsResponse struct {
	Records []models.Record `json:"data"`
	Offset  int             `json:"offset" example:"0"`
	Limit   int             `json:"limit" example:"10"`
	Total   int             `json:"total" example:"20"`
}

// PaginateRecordsResponse represents the response body for retrieve records with paginate
type PaginateUsersResponse struct {
	Users  []models.User `json:"data"`
	Offset int           `json:"offset" example:"0"`
	Limit  int           `json:"limit" example:"10"`
	Total  int           `json:"total" example:"20"`
}

// PaginateRecordsResponse represents the response body for retrieve records with paginate
type PaginateActivitiesResponse struct {
	Activities []models.Activity `json:"data"`
	Offset     int               `json:"offset" example:"0"`
	Limit      int               `json:"limit" example:"10"`
	Total      int               `json:"total" example:"20"`
}

// PaginateRecordsResponse represents the response body for retrieve records with paginate
type PaginateSchoolsResponse struct {
	Schools []models.School `json:"data"`
	Offset  int             `json:"offset" example:"0"`
	Limit   int             `json:"limit" example:"10"`
	Total   int             `json:"total" example:"20"`
}

// DownloadResponse represents the response for a successful presigned download request.
type DownloadResponse struct {
	URL string `json:"url" example:"https://your-s3-bucket.s3.amazonaws.com/user_id/image.png?X-Amz-..."`
}
