package controllers

// ErrorResponse represents a generic error response.
type ErrorResponse struct {
	Message string `json:"message" example:"Error description"`
}

type SuccessfulResponse struct {
	Message string `json:"message" example:"operation success"`
}
