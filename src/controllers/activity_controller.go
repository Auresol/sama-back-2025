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
	"github.com/go-playground/validator/v10"
)

// ActivityController manages HTTP requests for activities.
type ActivityController struct {
	activityService *services.ActivityService
	validate        *validator.Validate
}

// NewActivityController creates a new ActivityController.
func NewActivityController(activityService *services.ActivityService, validate *validator.Validate) *ActivityController {
	return &ActivityController{
		activityService: activityService,
		validate:        validate,
	}
}

// CreateActivityRequest defines the request body for creating a new activity.
type CreateActivityRequest struct {
	Name                string                 `json:"name" binding:"required" example:"School Cleanup Drive"`
	Template            map[string]interface{} `json:"template" binding:"required" swaggertype:"object,string" example:"field:test"`
	CoverImageUrl       *string                `json:"cover_image_url" example:"test/example"`
	IsRequired          bool                   `json:"is_required" binding:"required" example:"true"`
	IsForJunior         bool                   `json:"is_for_junior" validate:"required" example:"true"`
	IsForSenior         bool                   `json:"is_for_senior" validate:"required" example:"true"`
	ExclusiveClassrooms []string               `json:"exclusive_classrooms"  binding:"required" example:"1/1"`
	ExclusiveStudentIDs []uint                 `json:"exclusive_student_ids"  binding:"required" example:"101"`
	Deadline            *time.Time             `json:"deadline,omitempty" example:"2025-07-28T15:49:03.123Z"`
	FinishedUnit        string                 `json:"finished_unit" binding:"required,oneof=TIMES HOURS" example:"HOURS"`
	FinishedAmount      uint                   `json:"finished_amount" binding:"required" example:"10"`
	CanExceedLimit      bool                   `json:"can_exceed_limit" biding:"required" example:"false"`
	Semester            uint                   `json:"semester,omitempty" example:"1"`
	SchoolYear          uint                   `json:"school_year,omitempty" example:"2568"`
	UpdateProtocol      string                 `json:"update_protocol" binding:"required,oneof=RE_EVALUATE_ALL_RECORDS IGNORE_PAST_RECORDS" example:"RE_EVALUATE_ALL_RECORDS"`
}

// UpdateActivityRequest defines the request body for updating an activity.
type UpdateActivityRequest struct {
	Name                string                 `json:"name" binding:"required" example:"School Cleanup Drive"`
	Template            map[string]interface{} `json:"template" binding:"required" swaggertype:"object,string" example:"field:test"`
	CoverImageUrl       *string                `json:"cover_image_url" example:"test/example"`
	IsRequired          bool                   `json:"is_required" binding:"required" example:"true"`
	IsForJunior         bool                   `json:"is_for_junior" validate:"required" example:"true"`
	IsForSenior         bool                   `json:"is_for_senior" validate:"required" example:"true"`
	ExclusiveClassrooms []string               `json:"exclusive_classrooms"  binding:"required" example:"1/1"`
	ExclusiveStudentIDs []uint                 `json:"exclusive_student_ids"  binding:"required" example:"101"`
	Deadline            *time.Time             `json:"deadline,omitempty" example:"2025-07-28T15:49:03.123Z"`
	FinishedUnit        string                 `json:"finished_unit" binding:"required,oneof=TIMES HOURS" example:"HOURS"`
	FinishedAmount      uint                   `json:"finished_amount" binding:"required" example:"10"`
	CanExceedLimit      bool                   `json:"can_exceed_limit" biding:"required" example:"false"`
	UpdateProtocol      string                 `json:"update_protocol" binding:"required,oneof=RE_EVALUATE_ALL_RECORDS IGNORE_PAST_RECORDS" example:"RE_EVALUATE_ALL_RECORDS"`
}

// CreateActivity handles creating a new activity.
// @Summary Create a new activity
// @Description Create a new activity record with specified details, including template and student coverage. Requires TCH, ADMIN or Sama Crew role.
// @Tags Activity
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param activity body CreateActivityRequest true "Activity creation details"
// @Success 201 {object} models.Activity "Activity created successfully"
// @Failure 400 {object} ErrorResponse "Invalid request payload or validation error"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 403 {object} ErrorResponse "Forbidden (insufficient permissions)"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /activity [post]
func (c *ActivityController) CreateActivity(ctx *gin.Context) {
	claims, ok := middlewares.GetUserClaimsFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Message: "User claims not found in context"})
		return
	}

	// Authorization: Only Teachers, Admins, or Sama Crew can create activities
	if claims.Role != "TCH" && claims.Role != "ADMIN" && claims.Role != "SAMA" {
		ctx.JSON(http.StatusForbidden, ErrorResponse{Message: "Forbidden: Insufficient permissions to create activities"})
		return
	}

	var req CreateActivityRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid request payload: " + err.Error()})
		return
	}

	activity := &models.Activity{
		Name:                req.Name,
		Template:            req.Template,
		CoverImageUrl:       req.CoverImageUrl,
		SchoolID:            claims.SchoolID,
		IsRequired:          req.IsRequired,
		IsForJunior:         req.IsForJunior,
		IsForSenior:         req.IsForSenior,
		FinishedUnit:        req.FinishedUnit,
		FinishedAmount:      req.FinishedAmount,
		ExclusiveClassrooms: req.ExclusiveClassrooms,
		ExclusiveStudentIDs: req.ExclusiveStudentIDs,
		Semester:            req.Semester,
		SchoolYear:          req.SchoolYear,
		CanExceedLimit:      req.CanExceedLimit,
		UpdateProtocol:      req.UpdateProtocol,
		OwnerID:             claims.UserID,
		IsActive:            true,
	}

	// Prepare CustomStudentIDs for the service.
	// We need to convert []uint to []models.User with only ID populated for GORM association.
	// if req.CoverageType == "CUSTOM" {
	// 	if len(req.CustomStudentIDs) == 0 {
	// 		ctx.JSON(http.StatusBadRequest, ErrorResponse{Message: "CustomStudentIDs must not be empty when CoverageType is CUSTOM"})
	// 		return
	// 	}
	// 	// for _, studentID := range req.CustomStudentIDs {
	// 	// 	activity.CustomStudentIDs = append(activity.CustomStudentIDs, models.User{ID: studentID})
	// 	// }
	// } else {
	// 	// Ensure CustomStudentIDs is empty if CoverageType is not CUSTOM
	// 	//activity.CustomStudentIDs = nil
	// }

	if err := c.activityService.CreateActivity(activity); err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Failed to create activity: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, activity)
}

// GetActivityByID retrieves an activity by its ID.
// @Summary Get activity by ID
// @Description Retrieve details of a specific activity by its ID. Accessible by owner, relevant school ADMIN/TCH, or Sama Crew.
// @Tags Activity
// @Security BearerAuth
// @Produce json
// @Param id path int true "Activity ID"
// @Success 200 {object} models.ActivityWithStatistic "Activity retrieved successfully"
// @Failure 400 {object} ErrorResponse "Invalid activity ID"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 403 {object} ErrorResponse "Forbidden (not authorized to view this activity)"
// @Failure 404 {object} ErrorResponse "Activity not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /activity/{id} [get]
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
	if claims.Role != "SAMA" && claims.UserID != activity.OwnerID {
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
		// if activity.CoverageType == "CUSTOM" {
		// 	isCustomStudent := false
		// 	for _, student := range activity.CustomStudentIDs {
		// 		if student.ID == claims.UserID {
		// 			isCustomStudent = true
		// 			break
		// 		}
		// 	}
		// 	if !isCustomStudent {
		// 		ctx.JSON(http.StatusForbidden, ErrorResponse{Message: "Forbidden: Not authorized to view this activity."})
		// 		return
		// 	}
		// }
	}

	ctx.JSON(http.StatusOK, activity)
}

// GetAllActivity retrieves a list of activities.
// @Summary Get all activities
// @Description Retrieve a list of activities with optional filters by owner, school year, and semester. Requires ADMIN or Sama Crew role, or TCH for their own activities.
// @Tags Activity
// @Security BearerAuth
// @Produce json
// @Param owner_id query int false "Filter by owner User ID"
// @Param school_id query int false "Filter by School ID (Requires SAMA)"
// @Param classroom query string false "Filter by classroom"
// @Param semester query int false "Filter by Semester"
// @Param school_year query int false "Filter by School Year"
// @Param limit query int false "Limit for pagination" default(10)
// @Param offset query int false "Offset for pagination" default(0)
// @Success 200 {object} PaginateActivitiesResponse "List of activities retrieved successfully"
// @Failure 400 {object} ErrorResponse "Invalid query parameters"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 403 {object} ErrorResponse "Forbidden (insufficient permissions)"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /activity [get]
func (c *ActivityController) GetAllActivities(ctx *gin.Context) {
	claims, ok := middlewares.GetUserClaimsFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Message: "User claims not found in context"})
		return
	}

	// Authorization:
	// SAMA can fetch all activities (or filtered by any owner/school).
	// ADMIN can fetch activities for their school (optionally filtered by owner in their school).
	// TCH can only fetch their own activities.
	if claims.Role != "SAMA" && claims.Role != "ADMIN" && claims.Role != "TCH" {
		ctx.JSON(http.StatusForbidden, ErrorResponse{Message: "Forbidden: Insufficient permissions to list activities"})
		return
	}

	// classroom := ctx.DefaultQuery("classroom", "")
	semester, _ := strconv.ParseUint(ctx.DefaultQuery("semester", "0"), 10, 64)
	schoolYear, _ := strconv.ParseUint(ctx.DefaultQuery("school_year", "0"), 10, 64)
	ownerID, _ := strconv.ParseUint(ctx.DefaultQuery("owner_id", "0"), 10, 64)
	schoolID, _ := strconv.ParseUint(ctx.DefaultQuery("school_id", "0"), 10, 64)
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(ctx.DefaultQuery("offset", "0"))

	// Apply authorization filtering
	if claims.Role == "TCH" {
		// Teacher can only see their own activities
		// ownerID = uint64(claims.UserID)
		// Optionally, restrict by their school ID too if the activity model has SchoolID
		// For now, relies on owner_id for TCH.
	} else if claims.Role == "ADMIN" {
		// Admin can only see activities within their school.
		schoolID = uint64(claims.SchoolID)
		// If owner_id is also provided by ADMIN, ensure that owner belongs to the same school.
	}
	// SAMA has no restrictions on ownerID or schoolID.

	activities, count, err := c.activityService.GetAllActivities(uint(ownerID), uint(schoolID), uint(semester), uint(schoolYear), limit, offset)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Failed to retrieve activities: " + err.Error()})
		return
	}

	response := PaginateActivitiesResponse{
		Activities: activities,
		Limit:      limit,
		Offset:     offset,
		Total:      count,
	}

	ctx.JSON(http.StatusOK, response)
}

// UpdateActivity handles updating an existing activity.
// @Summary Update an activity
// @Description Update an existing activity record by ID. Requires activity owner (TCH/ADMIN), or Sama Crew role.
// @Tags Activity
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
// @Router /activity/{id} [put]
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

	// Authorization: Only owner or SAMA can update
	if claims.Role != "SAMA" && claims.UserID != existingActivity.OwnerID {
		ctx.JSON(http.StatusForbidden, ErrorResponse{Message: "Forbidden: You are not authorized to update this activity"})
		return
	}

	var req UpdateActivityRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid request payload: " + err.Error()})
		return
	}

	activity := &models.Activity{
		ID:                  existingActivity.ID,
		Name:                req.Name,
		Template:            req.Template,
		CoverImageUrl:       req.CoverImageUrl,
		SchoolID:            existingActivity.SchoolID,
		IsRequired:          req.IsRequired,
		IsForJunior:         req.IsForJunior,
		IsForSenior:         req.IsForSenior,
		FinishedUnit:        req.FinishedUnit,
		FinishedAmount:      req.FinishedAmount,
		ExclusiveClassrooms: req.ExclusiveClassrooms,
		ExclusiveStudentIDs: req.ExclusiveStudentIDs,
		CanExceedLimit:      req.CanExceedLimit,
		UpdateProtocol:      req.UpdateProtocol,
		OwnerID:             existingActivity.OwnerID,
		IsActive:            existingActivity.IsActive,
	}

	if err := c.activityService.UpdateActivity(activity); err != nil {
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
// @Tags Activity
// @Security BearerAuth
// @Produce json
// @Param id path int true "Activity ID to delete"
// @Success 204 {object} SuccessfulResponse "Activity deleted successfully"
// @Failure 400 {object} ErrorResponse "Invalid activity ID"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 403 {object} ErrorResponse "Forbidden (insufficient permissions or not owner)"
// @Failure 404 {object} ErrorResponse "Activity not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /activity/{id} [delete]
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

	// Authorization: Only owner or SAMA can delete
	if claims.Role != "SAMA" && claims.UserID != existingActivity.OwnerID {
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
