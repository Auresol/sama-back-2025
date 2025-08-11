package controllers

import (
	"fmt"
	"net/http"
	"strconv"

	"sama/sama-backend-2025/src/middlewares"
	"sama/sama-backend-2025/src/models"
	"sama/sama-backend-2025/src/services"
	"sama/sama-backend-2025/src/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// UserController manages HTTP requests for user accounts.
type UserController struct {
	userService     *services.UserService
	activityService *services.ActivityService
	recordService   *services.RecordService
	validate        *validator.Validate
}

// NewUserController creates a new UserController.
func NewUserController(
	userService *services.UserService,
	activityService *services.ActivityService,
	recordService *services.RecordService,
	validate *validator.Validate,
) *UserController {
	return &UserController{
		userService:     userService,
		activityService: activityService,
		recordService:   recordService,
		validate:        validate,
	}
}

// UpdateUserProfileRequest represents the request body for updating a user's profile.
// Use a separate struct for update requests to control what fields can be updated.
type UpdateUserProfileRequest struct {
	StudentID         string  `json:"user_id,omitempty" example:"10101"`
	Email             string  `json:"email" binding:"omitempty,email" example:"new_email@example.com"`
	Phone             string  `json:"phone" example:"+1987654321"`
	Firstname         string  `json:"firstname" example:"Jane"`
	Lastname          string  `json:"lastname" example:"Doe"`
	ProfilePictureURL *string `json:"profile_picture_url,omitempty" example:"http://example.com/pic.jpg"`
	Classroom         *string `json:"classroom,omitempty" example:"1/1" validate:"classroomregex"`
	Number            *uint   `json:"number,omitempty" binding:"omitempty,number" example:"2"` // Pointer for optional int update
	Language          string  `json:"language" example:"th"`
	BookmarkUserIDs   []uint  `json:"bookmark_user_ids" example:"1"`
}

type UserStatistic struct {
	NonCreatedPercent float32                        `json:"non_created_percent"`
	CreatedPercent    float32                        `json:"created_percent"`
	SendedPercnet     float32                        `json:"sended_percent"`
	ApprovedPercent   float32                        `json:"approved_percent"`
	RejectedPercent   float32                        `json:"rejected_percent"`
	Activities        []models.ActivityWithStatistic `json:"activities"`
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
// @Router /user/me [get]
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
	claims, ok := middlewares.GetUserClaimsFromContext(c)
	if !ok {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "User claims not found in context"})
		return
	}

	// STD are only allow for getMyProfile
	// if claims.Role == "STD" {
	// 	c.JSON(http.StatusForbidden, ErrorResponse{Message: "Forbidden: Insufficient permissions"})
	// 	return
	// }

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid user ID"})
		return
	}

	user, err := h.userService.GetUserByID(uint(id))
	if err != nil {
		if err.Error() == fmt.Sprintf("user with ID %d not found", id) {
			c.JSON(http.StatusNotFound, ErrorResponse{Message: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Failed to retrieve user: " + err.Error()})
		return
	}

	// Can get user outside their school only if they are SAMA
	if claims.Role != "SAMA" && claims.SchoolID != user.SchoolID {
		c.JSON(http.StatusForbidden, ErrorResponse{Message: "Forbidden: Insufficient permissions (user not in your school)"})
		return
	}

	c.JSON(http.StatusOK, user)
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

	userToUpdate, err := h.userService.GetUserByID(uint(id))
	if err != nil {
		if err.Error() == fmt.Sprintf("user with ID %d not found", id) {
			c.JSON(http.StatusNotFound, ErrorResponse{Message: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Failed to retrieve user for update: " + err.Error()})
		return
	}

	// For STD and TCH, do not allow to update other user
	if (claims.Role == "STD" || claims.Role == "TCH") && claims.UserID != userToUpdate.ID {
		c.JSON(http.StatusForbidden, ErrorResponse{Message: "Forbidden: Can only update your profile"})
		return
	}

	// For ADMIN, allow only their profile and other non-admin in the same school
	if claims.Role == "ADMIN" && userToUpdate.SchoolID != claims.SchoolID && !(userToUpdate.ID == claims.UserID || userToUpdate.Role == "STD" || userToUpdate.Role == "TCH") {
		c.JSON(http.StatusForbidden, ErrorResponse{Message: "Forbidden: Can only update your profile or anyone not ADMIN in your school"})
		return
	}

	var req UpdateUserProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid request payload: " + err.Error()})
		return
	}

	userToUpdate.Email = req.Email
	userToUpdate.Phone = req.Phone
	userToUpdate.Firstname = req.Firstname
	userToUpdate.Lastname = req.Lastname
	userToUpdate.ProfilePictureURL = req.ProfilePictureURL
	userToUpdate.Classroom = req.Classroom
	userToUpdate.Number = req.Number
	userToUpdate.Language = req.Language
	userToUpdate.BookmarkUserIDs = req.BookmarkUserIDs

	if err := h.userService.UpdateUserProfile(userToUpdate); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Failed to update user profile: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, userToUpdate)
}

// DeleteUser handles deleting a user.
// @Summary Delete a user
// @Description Delete a user account by ID. Requires ADMIN or Sama Crew role, or user deleting self.
// @Tags User
// @Security BearerAuth
// @Produce json
// @Param id path int true "User ID to delete"
// @Success 204 {object} SuccessfulResponse "User deleted successfully"
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

	user, err := h.userService.GetUserByID(uint(id))
	if err != nil {
		if err.Error() == fmt.Sprintf("user with ID %d not found", id) {
			c.JSON(http.StatusNotFound, ErrorResponse{Message: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Failed to retrieve user: " + err.Error()})
		return
	}

	// For STD and TCH, do not allow to update other user
	if (claims.Role == "STD" || claims.Role == "TCH") && claims.UserID != user.ID {
		c.JSON(http.StatusForbidden, ErrorResponse{Message: "Forbidden: Can only delete your profile"})
		return
	}

	// For ADMIN, allow only their profile and other non-admin in the same school
	if claims.Role == "ADMIN" && user.SchoolID != claims.SchoolID && !(user.ID == claims.UserID || user.Role == "STD" || user.Role == "TCH") {
		c.JSON(http.StatusForbidden, ErrorResponse{Message: "Forbidden: Can only delete your profile or anyone not ADMIN in your school"})
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

// GetAssignedActivity retrieves a list of activities related to the authenticated user.
// This includes activities where the user is the owner, or part of exclusive classrooms/students.
// @Summary Get activities related to the user
// @Description Retrieve a list of activities that are assigned to or owned by the authenticated user.
// @Tags User
// @Security BearerAuth
// @Param id path int true "User ID to get"
// @Param semester query int false "School semester"
// @Param school_year query int false "School year"
// @Produce json
// @Success 200 {array} models.ActivityWithStatistic "List of related activities retrieved successfully"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /user/{id}/activity [get]
func (c *UserController) GetAssignedActivities(ctx *gin.Context) {
	claims, ok := middlewares.GetUserClaimsFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Message: "User claims not found in context"})
		return
	}

	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid user ID"})
		return
	}

	semester, _ := strconv.ParseUint(ctx.DefaultQuery("semester", "0"), 10, 64)
	schoolYear, _ := strconv.ParseUint(ctx.DefaultQuery("school_year", "0"), 10, 64)

	// TODO: Implement the service call to fetch activities related to claims.UserID
	// This service method would need to query activities where:
	// 1. owner_id matches claims.UserID
	// 2. coverage_type is 'ALL' (if applicable to this user's school)
	// 3. user is in an exclusive_classroom (requires joining through activity_exclusive_classrooms and Classroom model's composite PK)
	// 4. user is in exclusive_student_ids (requires joining through activity_exclusive_student_ids)
	// This will be a more complex query in the repository.

	// Example placeholder for activities:
	activities, err := c.activityService.GetAssignedActivitiesByUserID(uint(id), claims.SchoolID, uint(semester), uint(schoolYear))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Failed to retrieve related activities: " + err.Error()})
		return
	}

	// For now, returning a placeholder response
	ctx.JSON(http.StatusOK, activities) // Return an empty array or mock data
}

// GetStatisticByID retrieves a statistic of user
// @Summary Get statistic of uesr
// @Description Retrieve a list of activities that are assigned to or owned by the authenticated user.
// @Tags User
// @Security BearerAuth
// @Param id path int true "User ID to get"
// @Param activity_id query string true "Activity id list seperate by |"
// @Param semester query int false "School semester"
// @Param school_year query int false "School year"
// @Produce json
// @Success 200 {object} UserStatistic "List of related activities retrieved successfully"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /user/{id}/statistic [get]
func (c *UserController) GetUserStatisticByID(ctx *gin.Context) {
	claims, ok := middlewares.GetUserClaimsFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Message: "User claims not found in context"})
		return
	}

	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid user ID"})
		return
	}

	activityIDs, err := utils.SplitQueryUint(ctx.Query("activity_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Message: "Failed to read activity_ids query: " + err.Error()})
		return
	}
	semester, _ := strconv.ParseUint(ctx.DefaultQuery("semester", "0"), 10, 64)
	schoolYear, _ := strconv.ParseUint(ctx.DefaultQuery("school_year", "0"), 10, 64)

	// Example placeholder for activities:
	activities,
		totalNonCreated,
		totalCreated,
		totalSended,
		totalApproved,
		totalRejected,
		err := c.userService.GetUserStatistic(uint(id), claims.SchoolID, activityIDs, uint(semester), uint(schoolYear))

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Failed to retrieve statistic: " + err.Error()})
		return
	}

	response := UserStatistic{
		NonCreatedPercent: totalNonCreated,
		CreatedPercent:    totalCreated,
		SendedPercnet:     totalSended,
		ApprovedPercent:   totalApproved,
		RejectedPercent:   totalRejected,
		Activities:        activities,
	}

	// For now, returning a placeholder response
	ctx.JSON(http.StatusOK, response) // Return an empty array or mock data
}
