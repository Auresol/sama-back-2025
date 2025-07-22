package controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"sama/sama-backend-2025/src/middlewares" // Renamed from middleware
	"sama/sama-backend-2025/src/models"
	"sama/sama-backend-2025/src/services" // Renamed from service

	"github.com/gin-gonic/gin"
)

// ActivityController manages HTTP requests for activities.
type ActivityController struct {
	activityService *services.ActivityService
}

// NewActivityController creates a new ActivityController.
func NewActivityController(activityService *services.ActivityService) *ActivityController {
	return &ActivityController{
		activityService: activityService,
	}
}

// CreateActivityRequest defines the request body for creating an activity.
type CreateActivityRequest struct {
	Name              string                  `json:"activity_name" binding:"required" example:"Tree Planting Event"`
	Template          models.ActivityTemplate `json:"template" binding:"required"` // Can use map[string]interface{} or models.ActivityTemplate
	CoverageType      string                  `json:"coverage_type" binding:"required,oneof=REQUIRE CUSTOM" example:"CUSTOM"`
	CustomStudentIDs  []uint                  `json:"custom_student_ids,omitempty"` // IDs of students, not full User objects
	FinishedCondition string                  `json:"finished_condition" binding:"required,oneof=TIMES HOURS" example:"HOURS"`
	Status            string                  `json:"status" binding:"required,oneof=REQUIRE CUSTOM" example:"REQUIRE"`
	UpdateProtocol    string                  `json:"update_protocol" binding:"required,oneof=RE_EVALUATE_ALL_RECORDS IGNORE_PAST_RECORDS" example:"RE_EVALUATE_ALL_RECORDS"`
	SchoolYear        int                     `json:"school_year" binding:"required,gt=0" example:"2568"`
	Semester          int                     `json:"semester" binding:"required,gt=0" example:"1"`
}

// UpdateActivityRequest defines the request body for updating an activity.
type UpdateActivityRequest struct {
	Name              string                  `json:"activity_name,omitempty" binding:"omitempty" example:"School Cleanup"`
	Template          models.ActivityTemplate `json:"template,omitempty"` // Can use map[string]interface{} or models.ActivityTemplate
	CoverageType      string                  `json:"coverage_type,omitempty" binding:"omitempty,oneof=REQUIRE CUSTOM" example:"REQUIRE"`
	CustomStudentIDs  []uint                  `json:"custom_student_ids,omitempty" ` // Provide empty array to clear
	IsActive          *bool                   `json:"is_active,omitempty" example:"true"`
	FinishedCondition string                  `json:"finished_condition,omitempty" binding:"omitempty,oneof=TIMES HOURS" example:"TIMES"`
	Status            string                  `json:"status,omitempty" binding:"omitempty,oneof=REQUIRE CUSTOM" example:"REQUIRE"`
	UpdateProtocol    string                  `json:"update_protocol,omitempty" binding:"omitempty,oneof=RE_EVALUATE_ALL_RECORDS IGNORE_PAST_RECORDS" example:"IGNORE_PAST_RECORDS"`
	SchoolYear        int                     `json:"school_year,omitempty" binding:"omitempty,gt=0" example:"2569"`
	Semester          int                     `json:"semester,omitempty" binding:"omitempty,gt=0" example:"2"`
}

// CreateActivity handles creating a new activity.
// @Summary Create a new activity
// @Description Create a new activity record with specified details, including template and student coverage. Requires TCH, ADMIN or Sama Crew role.
// @Tags Activities
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param activity body CreateActivityRequest true "Activity creation details"
// @Success 201 {object} models.Activity "Activity created successfully"
// @Failure 400 {object} ErrorResponse "Invalid request payload or validation error"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 403 {object} ErrorResponse "Forbidden (insufficient permissions)"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /activities [post]
func (c *ActivityController) CreateActivity(ctx *gin.Context) {
	claims, ok := middlewares.GetUserClaimsFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Message: "User claims not found in context"})
		return
	}

	// Authorization: Only Teachers, Admins, or Sama Crew can create activities
	if claims.Role != "TCH" && claims.Role != "ADMIN" && claims.Role != "SAMA_CREW" {
		ctx.JSON(http.StatusForbidden, ErrorResponse{Message: "Forbidden: Insufficient permissions to create activities"})
		return
	}

	var req CreateActivityRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid request payload: " + err.Error()})
		return
	}

	activity := &models.Activity{
		Name:              req.Name,
		Template:          req.Template,
		CoverageType:      req.CoverageType,
		FinishedCondition: req.FinishedCondition,
		Status:            req.Status,
		UpdateProtocol:    req.UpdateProtocol,
		SchoolYear:        req.SchoolYear,
		Semester:          req.Semester,
		OwnerID:           claims.UserID, // Set owner from authenticated user
		IsActive:          true,          // Default to active on creation
	}

	// Prepare CustomStudentIDs for the service.
	// We need to convert []uint to []models.User with only ID populated for GORM association.
	if req.CoverageType == "CUSTOM" {
		if len(req.CustomStudentIDs) == 0 {
			ctx.JSON(http.StatusBadRequest, ErrorResponse{Message: "CustomStudentIDs must not be empty when CoverageType is CUSTOM"})
			return
		}
		// for _, studentID := range req.CustomStudentIDs {
		// 	activity.CustomStudentIDs = append(activity.CustomStudentIDs, models.User{ID: studentID})
		// }
	} else {
		// Ensure CustomStudentIDs is empty if CoverageType is not CUSTOM
		activity.CustomStudentIDs = nil
	}

	if err := c.activityService.CreateActivity(activity); err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Failed to create activity: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, activity)
}

// GetActivityByID retrieves an activity by its ID.
// @Summary Get activity by ID
// @Description Retrieve details of a specific activity by its ID. Accessible by owner, relevant school ADMIN/TCH, or Sama Crew.
// @Tags Activities
// @Security BearerAuth
// @Produce json
// @Param id path int true "Activity ID"
// @Success 200 {object} models.Activity "Activity retrieved successfully"
// @Failure 400 {object} ErrorResponse "Invalid activity ID"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 403 {object} ErrorResponse "Forbidden (not authorized to view this activity)"
// @Failure 404 {object} ErrorResponse "Activity not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /activities/{id} [get]
func (c *ActivityController) GetActivityByID(ctx *gin.Context) {
	claims, ok := middlewares.GetUserClaimsFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Message: "User claims not found in context"})
		return
	}

	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid activity ID"})
		return
	}

	activity, err := c.activityService.GetActivityByID(uint(id))
	if err != nil {
		if err.Error() == fmt.Sprintf("activity with ID %d not found", id) {
			ctx.JSON(http.StatusNotFound, ErrorResponse{Message: err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Failed to retrieve activity: " + err.Error()})
		return
	}

	// Authorization logic for viewing an activity:
	// 1. Sama Crew can view any activity.
	// 2. Owner of the activity can view it.
	// 3. ADMIN/TCH of the same school as the activity's owner (assuming owner's school is tied to activity)
	//    or if the activity is school-wide for their school, can view it.
	if claims.Role != "SAMA_CREW" && claims.UserID != activity.OwnerID {
		// // Need to fetch owner's school ID to compare
		// owner, err := c.activityService.userRepo.GetUserByID(activity.OwnerID)
		// if err != nil {
		// 	ctx.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Failed to verify activity owner's school for authorization"})
		// 	return
		// }
		// if (claims.Role == "ADMIN" || claims.Role == "TCH") && claims.SchoolID != owner.SchoolID {
		// 	ctx.JSON(http.StatusForbidden, ErrorResponse{Message: "Forbidden: Not authorized to view this activity."})
		// 	return
		// }
		// If the activity is CUSTOM coverage, and the current user is a CustomStudent, they can view it.
		// This requires checking if claims.UserID is in activity.CustomStudentIDs.
		if activity.CoverageType == "CUSTOM" {
			isCustomStudent := false
			for _, student := range activity.CustomStudentIDs {
				if student.ID == claims.UserID {
					isCustomStudent = true
					break
				}
			}
			if !isCustomStudent {
				ctx.JSON(http.StatusForbidden, ErrorResponse{Message: "Forbidden: Not authorized to view this activity."})
				return
			}
		}
	}

	ctx.JSON(http.StatusOK, activity)
}

// GetAllActivities retrieves a list of activities.
// @Summary Get all activities
// @Description Retrieve a list of activities with optional filters by owner, school year, and semester. Requires ADMIN or Sama Crew role, or TCH for their own activities.
// @Tags Activities
// @Security BearerAuth
// @Produce json
// @Param owner_id query int false "Filter by owner User ID"
// @Param school_id query int false "Filter by School ID (Requires SAMA_CREW)"
// @Param school_year query int false "Filter by School Year"
// @Param semester query int false "Filter by Semester"
// @Param limit query int false "Limit for pagination" default(10)
// @Param offset query int false "Offset for pagination" default(0)
// @Success 200 {array} models.Activity "List of activities retrieved successfully"
// @Failure 400 {object} ErrorResponse "Invalid query parameters"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 403 {object} ErrorResponse "Forbidden (insufficient permissions)"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /activities [get]
func (c *ActivityController) GetAllActivities(ctx *gin.Context) {
	claims, ok := middlewares.GetUserClaimsFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Message: "User claims not found in context"})
		return
	}

	// Authorization:
	// SAMA_CREW can fetch all activities (or filtered by any owner/school).
	// ADMIN can fetch activities for their school (optionally filtered by owner in their school).
	// TCH can only fetch their own activities.
	if claims.Role != "SAMA_CREW" && claims.Role != "ADMIN" && claims.Role != "TCH" {
		ctx.JSON(http.StatusForbidden, ErrorResponse{Message: "Forbidden: Insufficient permissions to list activities"})
		return
	}

	ownerID, _ := strconv.ParseUint(ctx.DefaultQuery("owner_id", "0"), 10, 64)
	schoolID, _ := strconv.ParseUint(ctx.DefaultQuery("school_id", "0"), 10, 64)
	schoolYear, _ := strconv.Atoi(ctx.DefaultQuery("school_year", "0"))
	semester, _ := strconv.Atoi(ctx.DefaultQuery("semester", "0"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(ctx.DefaultQuery("offset", "0"))

	// Apply authorization filtering
	if claims.Role == "TCH" {
		// Teacher can only see their own activities
		ownerID = uint64(claims.UserID)
		// Optionally, restrict by their school ID too if the activity model has SchoolID
		// For now, relies on owner_id for TCH.
	} else if claims.Role == "ADMIN" {
		// Admin can only see activities within their school.
		schoolID = uint64(claims.SchoolID)
		// If owner_id is also provided by ADMIN, ensure that owner belongs to the same school.
	}
	// SAMA_CREW has no restrictions on ownerID or schoolID.

	activities, err := c.activityService.GetAllActivities(uint(ownerID), uint(schoolID), schoolYear, semester, limit, offset)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Failed to retrieve activities: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, activities)
}

// UpdateActivity handles updating an existing activity.
// @Summary Update an activity
// @Description Update an existing activity record by ID. Requires activity owner (TCH/ADMIN), or Sama Crew role.
// @Tags Activities
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Activity ID to update"
// @Param activity body UpdateActivityRequest true "Activity update details"
// @Success 200 {object} models.Activity "Activity updated successfully"
// @Failure 400 {object} ErrorResponse "Invalid request payload or validation error"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 403 {object} ErrorResponse "Forbidden (insufficient permissions or not owner)"
// @Failure 404 {object} ErrorResponse "Activity not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /activities/{id} [put]
func (c *ActivityController) UpdateActivity(ctx *gin.Context) {
	claims, ok := middlewares.GetUserClaimsFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Message: "User claims not found in context"})
		return
	}

	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid activity ID"})
		return
	}

	existingActivity, err := c.activityService.GetActivityByID(uint(id))
	if err != nil {
		if err.Error() == fmt.Sprintf("activity with ID %d not found", id) {
			ctx.JSON(http.StatusNotFound, ErrorResponse{Message: err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Failed to retrieve activity for update: " + err.Error()})
		return
	}

	// Authorization: Only owner or SAMA_CREW can update
	if claims.Role != "SAMA_CREW" && claims.UserID != existingActivity.OwnerID {
		ctx.JSON(http.StatusForbidden, ErrorResponse{Message: "Forbidden: You are not authorized to update this activity"})
		return
	}

	var req UpdateActivityRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid request payload: " + err.Error()})
		return
	}

	// Create a new models.Activity and populate fields from request
	// This helps with GORM's .Save() to only update provided fields if you load the existing
	// and then update specific fields. Or, pass the new values to the service.
	activityToUpdate := &models.Activity{ID: uint(id)} // Ensure ID is set

	// Manually map fields from request to activityToUpdate, respecting omitempty
	if req.Name != "" {
		activityToUpdate.Name = req.Name
	} else {
		activityToUpdate.Name = existingActivity.Name
	}
	if req.Template != nil {
		activityToUpdate.Template = req.Template
	} else {
		activityToUpdate.Template = existingActivity.Template
	}
	if req.CoverageType != "" {
		activityToUpdate.CoverageType = req.CoverageType
	} else {
		activityToUpdate.CoverageType = existingActivity.CoverageType
	}
	// For CustomStudentIDs, handle explicitly:
	if req.CustomStudentIDs != nil { // if CustomStudentIDs is provided in the request
		// if req.CoverageType == "CUSTOM" { // if CoverageType is specified as CUSTOM
		// 	for _, studentID := range req.CustomStudentIDs {
		// 		activityToUpdate.CustomStudentIDs = append(activityToUpdate.CustomStudentIDs, models.User{ID: studentID})
		// 	}
		// } else if existingActivity.CoverageType == "CUSTOM" && req.CoverageType == "" { // if it was CUSTOM but req doesn't specify it, use existing
		// 	for _, studentID := range req.CustomStudentIDs {
		// 		activityToUpdate.CustomStudentIDs = append(activityToUpdate.CustomStudentIDs, models.User{ID: studentID})
		// 	}
		// } else { // if it's not CUSTOM (either changing from custom, or already non-custom), clear them
		// 	activityToUpdate.CustomStudentIDs = nil
		// }
	} else { // if CustomStudentIDs is not provided in the request, retain existing
		activityToUpdate.CustomStudentIDs = existingActivity.CustomStudentIDs
	}

	if req.IsActive != nil {
		activityToUpdate.IsActive = *req.IsActive
	} else {
		activityToUpdate.IsActive = existingActivity.IsActive
	}
	if req.FinishedCondition != "" {
		activityToUpdate.FinishedCondition = req.FinishedCondition
	} else {
		activityToUpdate.FinishedCondition = existingActivity.FinishedCondition
	}
	if req.Status != "" {
		activityToUpdate.Status = req.Status
	} else {
		activityToUpdate.Status = existingActivity.Status
	}
	if req.UpdateProtocol != "" {
		activityToUpdate.UpdateProtocol = req.UpdateProtocol
	} else {
		activityToUpdate.UpdateProtocol = existingActivity.UpdateProtocol
	}
	if req.SchoolYear != 0 {
		activityToUpdate.SchoolYear = req.SchoolYear
	} else {
		activityToUpdate.SchoolYear = existingActivity.SchoolYear
	}
	if req.Semester != 0 {
		activityToUpdate.Semester = req.Semester
	} else {
		activityToUpdate.Semester = existingActivity.Semester
	}

	// OwnerID and Timestamps are usually not updated via public API
	activityToUpdate.OwnerID = existingActivity.OwnerID
	activityToUpdate.CreatedAt = existingActivity.CreatedAt
	activityToUpdate.UpdatedAt = time.Now() // Explicitly set updated time

	if err := c.activityService.UpdateActivity(activityToUpdate); err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Failed to update activity: " + err.Error()})
		return
	}

	// Re-fetch to get updated associations
	updatedActivity, err := c.activityService.GetActivityByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Failed to retrieve updated activity: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, updatedActivity)
}

// DeleteActivity handles deleting an activity.
// @Summary Delete an activity
// @Description Delete an activity record by ID. Requires activity owner (TCH/ADMIN), or Sama Crew role.
// @Tags Activities
// @Security BearerAuth
// @Produce json
// @Param id path int true "Activity ID to delete"
// @Success 204 "Activity deleted successfully"
// @Failure 400 {object} ErrorResponse "Invalid activity ID"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 403 {object} ErrorResponse "Forbidden (insufficient permissions or not owner)"
// @Failure 404 {object} ErrorResponse "Activity not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /activities/{id} [delete]
func (c *ActivityController) DeleteActivity(ctx *gin.Context) {
	claims, ok := middlewares.GetUserClaimsFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Message: "User claims not found in context"})
		return
	}

	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid activity ID"})
		return
	}

	existingActivity, err := c.activityService.GetActivityByID(uint(id))
	if err != nil {
		if err.Error() == fmt.Sprintf("activity with ID %d not found", id) {
			ctx.JSON(http.StatusNotFound, ErrorResponse{Message: err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Failed to retrieve activity for deletion: " + err.Error()})
		return
	}

	// Authorization: Only owner or SAMA_CREW can delete
	if claims.Role != "SAMA_CREW" && claims.UserID != existingActivity.OwnerID {
		ctx.JSON(http.StatusForbidden, ErrorResponse{Message: "Forbidden: You are not authorized to delete this activity"})
		return
	}

	if err := c.activityService.DeleteActivity(uint(id)); err != nil {
		if err.Error() == fmt.Sprintf("activity with ID %d not found for deletion", id) {
			ctx.JSON(http.StatusNotFound, ErrorResponse{Message: err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Failed to delete activity: " + err.Error()})
		return
	}

	ctx.Status(http.StatusNoContent) // 204 No Content for successful deletion
}
