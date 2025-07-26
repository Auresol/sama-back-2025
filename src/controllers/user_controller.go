package controllers

import (
	"fmt"
	"net/http"
	"strconv"

	"sama/sama-backend-2025/src/middlewares"
	"sama/sama-backend-2025/src/models"
	"sama/sama-backend-2025/src/services"

	// For JWT claims
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// UserController manages HTTP requests for user accounts.
type UserController struct {
	userService *services.UserService
	validate    *validator.Validate
}

// NewUserController creates a new UserController.
func NewUserController(userService *services.UserService, validate *validator.Validate) *UserController {
	return &UserController{
		userService: userService,
		validate:    validate,
	}
}

// GetMyProfile retrieves the profile of the authenticated user.
// @Summary Get authenticated user's profile
// @Description Retrieve the profile details of the currently authenticated user.
// @Tags User
// @Security BearerAuth
// @Produce json
// @Success 200 {object} models.User "User profile retrieved successfully"
// @Failure 401 {object} ErrorResponse "Unauthorized (missing or invalid token)"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /me [get]
func (h *UserController) GetMyProfile(c *gin.Context) {
	claims, ok := middlewares.GetUserClaimsFromContext(c)
	if !ok {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "User claims not found in context"})
		return
	}

	user, err := h.userService.GetUserByID(claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Failed to retrieve user profile: " + err.Error()})
		return
	}

	// Omit password from response
	user.Password = ""
	c.JSON(http.StatusOK, user)
}

// GetUserByID retrieves a user by ID (requires ADMIN/Sama Crew role).
// @Summary Get user by ID
// @Description Retrieve a user's profile by their ID. Requires ADMIN or Sama Crew role.
// @Tags User
// @Security BearerAuth
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} models.User "User profile retrieved successfully"
// @Failure 400 {object} ErrorResponse "Invalid user ID"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 403 {object} ErrorResponse "Forbidden (insufficient permissions)"
// @Failure 404 {object} ErrorResponse "User not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /user/{id} [get]
func (h *UserController) GetUserByID(c *gin.Context) {
	// Example of role-based access control (RBAC)
	claims, ok := middlewares.GetUserClaimsFromContext(c)
	if !ok {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "User claims not found in context"})
		return
	}
	// Simplified RBAC: ADMIN or SAMA_CREW can get any user.
	// You might add more granular checks here (e.g., ADMIN can only get users from their school).
	if claims.Role != "ADMIN" && claims.Role != "SAMA_CREW" {
		c.JSON(http.StatusForbidden, ErrorResponse{Message: "Forbidden: Insufficient permissions"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid user ID"})
		return
	}

	user, err := h.userService.GetUserByID(uint(id))
	if err != nil {
		if err.Error() == fmt.Sprintf("user with ID %d not found", id) { // Check for specific not found error
			c.JSON(http.StatusNotFound, ErrorResponse{Message: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Failed to retrieve user: " + err.Error()})
		return
	}

	user.Password = "" // Omit password
	c.JSON(http.StatusOK, user)
}

// UpdateUserProfileRequest represents the request body for updating a user's profile.
// Use a separate struct for update requests to control what fields can be updated.
type UpdateUserProfileRequest struct {
	Email             string  `json:"email,omitempty" binding:"omitempty,email" example:"new_email@example.com"`
	Phone             string  `json:"phone,omitempty" example:"+1987654321"`
	Firstname         string  `json:"firstname,omitempty" example:"Jane"`
	Lastname          string  `json:"lastname,omitempty" example:"Doe"`
	ProfilePictureURL *string `json:"profile_picture_url,omitempty" example:"http://example.com/pic.jpg"`
	IsActive          *bool   `json:"is_active,omitempty" example:"true"` // Pointer for optional boolean update
	Classroom         string  `json:"classroom,omitempty" example:"B202"`
	Number            *uint   `json:"number,omitempty" binding:"omitempty,number" example:"2"` // Pointer for optional int update
	Status            string  `json:"status,omitempty" example:"active"`
	Language          string  `json:"language,omitempty" example:"th"`
	// Role and SchoolID are typically not updated via this endpoint or require special permissions
	// Password update should be a separate endpoint
}

// UpdateUserProfile handles updating a user's profile.
// @Summary Update user profile
// @Description Update an authenticated user's profile.
// @Tags User
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "User ID to update"
// @Param user body UpdateUserProfileRequest true "User profile data to update"
// @Success 200 {object} models.User "User profile updated successfully"
// @Failure 400 {object} ErrorResponse "Invalid request payload or validation error"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 403 {object} ErrorResponse "Forbidden (cannot update other users or insufficient permissions)"
// @Failure 404 {object} ErrorResponse "User not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /user/{id} [put]
func (h *UserController) UpdateUserProfile(c *gin.Context) {
	claims, ok := middlewares.GetUserClaimsFromContext(c)
	if !ok {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "User claims not found in context"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid user ID"})
		return
	}

	// Authorization: User can only update their own profile unless ADMIN/SAMA_CREW
	if claims.UserID != uint(id) && claims.Role != "ADMIN" && claims.Role != "SAMA_CREW" {
		c.JSON(http.StatusForbidden, ErrorResponse{Message: "Forbidden: You can only update your own profile"})
		return
	}

	var req UpdateUserProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid request payload: " + err.Error()})
		return
	}

	// Fetch the existing user to apply updates
	userToUpdate, err := h.userService.GetUserByID(uint(id))
	if err != nil {
		if err.Error() == fmt.Sprintf("user with ID %d not found", id) {
			c.JSON(http.StatusNotFound, ErrorResponse{Message: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Failed to retrieve user for update: " + err.Error()})
		return
	}

	// Apply updates from request to the fetched user model
	// Only update fields that are provided in the request
	if req.Email != "" {
		userToUpdate.Email = req.Email
	}
	if req.Phone != "" {
		userToUpdate.Phone = req.Phone
	}
	if req.Firstname != "" {
		userToUpdate.Firstname = req.Firstname
	}
	if req.Lastname != "" {
		userToUpdate.Lastname = req.Lastname
	}
	if req.ProfilePictureURL != nil { // Check if pointer is not nil
		userToUpdate.ProfilePictureURL = req.ProfilePictureURL
	}
	// if req.IsActive != nil { // Check if pointer is not nil
	// 	userToUpdate.IsActive = *req.IsActive
	// }
	if req.Classroom != "" {
		userToUpdate.Classroom = req.Classroom
	}
	if req.Number != nil { // Check if pointer is not nil
		userToUpdate.Number = *req.Number
	}
	if req.Status != "" {
		userToUpdate.Status = req.Status
	}
	if req.Language != "" {
		userToUpdate.Language = req.Language
	}

	if err := h.userService.UpdateUserProfile(userToUpdate); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Failed to update user profile: " + err.Error()})
		return
	}

	userToUpdate.Password = "" // Omit password from response
	c.JSON(http.StatusOK, userToUpdate)
}

// DeleteUser handles deleting a user.
// @Summary Delete a user
// @Description Delete a user account by ID. Requires ADMIN or Sama Crew role, or user deleting self.
// @Tags User
// @Security BearerAuth
// @Produce json
// @Param id path int true "User ID to delete"
// @Success 204 "User deleted successfully"
// @Failure 400 {object} ErrorResponse "Invalid user ID"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 403 {object} ErrorResponse "Forbidden (insufficient permissions)"
// @Failure 404 {object} ErrorResponse "User not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /user/{id} [delete]
func (h *UserController) DeleteUser(c *gin.Context) {
	claims, ok := middlewares.GetUserClaimsFromContext(c)
	if !ok {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "User claims not found in context"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid user ID"})
		return
	}

	// Authorization: User can delete their own profile, or ADMIN/SAMA_CREW can delete any.
	if claims.UserID != uint(id) && claims.Role != "ADMIN" && claims.Role != "SAMA_CREW" {
		c.JSON(http.StatusForbidden, ErrorResponse{Message: "Forbidden: You can only delete your own profile or require higher permissions"})
		return
	}

	if err := h.userService.DeleteUser(uint(id)); err != nil {
		if err.Error() == fmt.Sprintf("user with ID %d not found", id) { // Check for specific not found error
			c.JSON(http.StatusNotFound, ErrorResponse{Message: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Failed to delete user: " + err.Error()})
		return
	}

	c.Status(http.StatusNoContent) // 204 No Content for successful deletion
}

// GetRelatedActivities retrieves a list of activities related to the authenticated user. All activities will contain "".
// This includes activities where the user is the owner, or part of exclusive classrooms/students.
// @Summary Get activities related to the authenticated user
// @Description Retrieve a list of activities that are assigned to or owned by the authenticated user.
// @Tags User
// @Security BearerAuth
// @Produce json
// @Success 200 {array} models.Activity "List of related activities retrieved successfully"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /user/activities [get]
func (c *UserController) GetRelatedActivities(ctx *gin.Context) {
	// claims, ok := middlewares.GetUserClaimsFromContext(ctx)
	// if !ok {
	// 	ctx.JSON(http.StatusInternalServerError, ErrorResponse{Message: "User claims not found in context"})
	// 	return
	// }

	// TODO: Implement the service call to fetch activities related to claims.UserID
	// This service method would need to query activities where:
	// 1. owner_id matches claims.UserID
	// 2. coverage_type is 'ALL' (if applicable to this user's school)
	// 3. user is in an exclusive_classroom (requires joining through activity_exclusive_classrooms and Classroom model's composite PK)
	// 4. user is in exclusive_student_ids (requires joining through activity_exclusive_student_ids)
	// This will be a more complex query in the repository.

	// Example placeholder for activities:
	// activities, err := c.activityService.GetActivitiesForUser(claims.UserID, claims.SchoolID, limit, offset)
	// if err != nil {
	//     ctx.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Failed to retrieve related activities: " + err.Error()})
	//     return
	// }

	// For now, returning a placeholder response
	ctx.JSON(http.StatusOK, []models.Activity{}) // Return an empty array or mock data
}
