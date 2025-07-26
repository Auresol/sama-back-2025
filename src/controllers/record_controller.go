package controllers

import (
	// For handling Data (map[string]interface{}) as raw JSON
	"fmt"
	"net/http"
	"strconv"

	"sama/sama-backend-2025/src/middlewares"
	"sama/sama-backend-2025/src/models"
	"sama/sama-backend-2025/src/services"

	"github.com/gin-gonic/gin"
)

// RecordController manages HTTP requests for records.
type RecordController struct {
	recordService *services.RecordService
}

// NewRecordController creates a new RecordController.
func NewRecordController(recordService *services.RecordService) *RecordController {
	return &RecordController{
		recordService: recordService,
	}
}

// CreateRecordRequest defines the request body for creating a new record.
type CreateRecordRequest struct {
	ActivityID uint                   `json:"activity_id" binding:"required,gt=0" example:"1"` // Assuming ActivityID is uint
	Data       map[string]interface{} `json:"data" binding:"required" swaggertype:"object,string" example:"field:test"`
	Advise     string                 `json:"advise,omitempty" example:"Good effort!"`

	StudentID uint `json:"student_id" binding:"required,gt=0" example:"101"`
	TeacherID uint `json:"teacher_id" binding:"required,gt=0" example:"201"`

	SchoolYear int `json:"school_year" binding:"required,gt=0" example:"2568"`
	Semester   int `json:"semester" binding:"required,gt=0" example:"1"`

	Amount int `json:"amount" binding:"required" example:"5"`

	Status string `json:"status" binding:"required,oneof=CREATED SENDED APPROVED REJECTED" example:"CREATED"`
}

// UpdateRecordRequest defines the request body for updating an existing record.
type UpdateRecordRequest struct {
	ActivityID *uint                  `json:"activity_id,omitempty" binding:"omitempty,gt=0" example:"2"` // Pointer for optional update
	Data       map[string]interface{} `json:"data,omitempty" swaggertype:"object,string" example:"field:test"`
	Advise     *string                `json:"advise,omitempty" example:"Needs more practice."`

	SchoolID  *uint `json:"school_id,omitempty" binding:"omitempty,gt=0" example:"1"`
	StudentID *uint `json:"student_id,omitempty" binding:"omitempty,gt=0" example:"101"`
	TeacherID *uint `json:"teacher_id,omitempty" binding:"omitempty,gt=0" example:"201"`

	SchoolYear *int `json:"school_year,omitempty" binding:"omitempty,gt=0" example:"2569"`
	Semester   *int `json:"semester,omitempty" binding:"omitempty,gt=0" example:"2"`

	Amount *int `json:"amount,omitempty" example:"7"`

	Status *string `json:"status,omitempty" binding:"omitempty,oneof=CREATED SENDED APPROVED REJECTED" example:"APPROVED"` // Status can be updated
}

// CreateRecord handles creating a new record.
// @Summary Create a new record
// @Description Create a new activity record with associated student, teacher, school, and activity details.
// @Tags Records
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param record body CreateRecordRequest true "Record creation details"
// @Success 201 {object} models.Record "Record created successfully"
// @Failure 400 {object} ErrorResponse "Invalid request payload or validation error"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 403 {object} ErrorResponse "Forbidden (insufficient permissions)"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /record [post]
func (c *RecordController) CreateRecord(ctx *gin.Context) {
	claims, ok := middlewares.GetUserClaimsFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Message: "User claims not found in context"})
		return
	}

	// Authorization: Example - only teachers can create records for now.
	// You'll need to refine this based on your business logic (e.g., students creating their own records, etc.)
	if claims.Role != "TCH" && claims.Role != "ADMIN" && claims.Role != "SAMA_CREW" {
		ctx.JSON(http.StatusForbidden, ErrorResponse{Message: "Forbidden: Insufficient permissions to create records"})
		return
	}

	var req CreateRecordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid request payload: " + err.Error()})
		return
	}

	record := &models.Record{
		ActivityID: req.ActivityID, // Assuming this is uint
		Advise:     req.Advise,
		StudentID:  req.StudentID,
		TeacherID:  req.TeacherID,
		SchoolYear: req.SchoolYear,
		Semester:   req.Semester,
		Amount:     req.Amount,
		Status:     req.Status,
	}

	// Unmarshal raw JSON into map[string]interface{}
	// if len(req.Data) > 0 {
	// 	if err := json.Unmarshal(req.Data, &record.Data); err != nil {
	// 		ctx.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid data JSON: " + err.Error()})
	// 		return
	// 	}
	// } else {
	// 	record.Data = make(map[string]interface{}) // Ensure it's an empty map if not provided
	// }

	// Pass the authenticated user's ID for status log
	if err := c.recordService.CreateRecord(record, claims.UserID); err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Failed to create record: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, record)
}

// GetRecordByID retrieves a record by its ID.
// @Summary Get record by ID
// @Description Retrieve details of a specific record by its ID. Accessible by relevant student/teacher/admin, or Sama Crew.
// @Tags Records
// @Security BearerAuth
// @Produce json
// @Param id path int true "Record ID"
// @Success 200 {object} models.Record "Record retrieved successfully"
// @Failure 400 {object} ErrorResponse "Invalid record ID"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 403 {object} ErrorResponse "Forbidden (not authorized to view this record)"
// @Failure 404 {object} ErrorResponse "Record not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /record/{id} [get]
func (c *RecordController) GetRecordByID(ctx *gin.Context) {
	claims, ok := middlewares.GetUserClaimsFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Message: "User claims not found in context"})
		return
	}

	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid record ID"})
		return
	}

	record, err := c.recordService.GetRecordByID(uint(id))
	if err != nil {
		if err.Error() == fmt.Sprintf("record with ID %d not found", id) {
			ctx.JSON(http.StatusNotFound, ErrorResponse{Message: err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Failed to retrieve record: " + err.Error()})
		return
	}

	// Authorization logic for viewing a record:
	// 1. Sama Crew can view any record.
	// 2. Student can view their own records.
	// 3. Teacher can view records where they are the assigned teacher, or records for students in their school.
	// 4. Admin can view records for their school.
	if claims.Role != "SAMA_CREW" {
		isAuthorized := false
		// if claims.Role == "STD" && claims.UserID == record.StudentID {
		// 	isAuthorized = true
		// } else if claims.Role == "TCH" && (claims.UserID == record.TeacherID || claims.SchoolID == record.SchoolID) {
		// 	// Teacher can see records they are assigned to, or records for students in their school
		// 	isAuthorized = true
		// }
		// } else if claims.Role == "ADMIN" && claims.SchoolID == record.SchoolID {
		// 	isAuthorized = true
		// }

		if !isAuthorized {
			ctx.JSON(http.StatusForbidden, ErrorResponse{Message: "Forbidden: Not authorized to view this record."})
			return
		}
	}

	ctx.JSON(http.StatusOK, record)
}

// GetAllRecords retrieves a list of records with filtering and pagination.
// @Summary Get all records
// @Description Retrieve a list of records with optional filters (school, student, teacher, activity, status).
// @Tags Records
// @Security BearerAuth
// @Produce json
// @Param school_id query int false "Filter by School ID"
// @Param student_id query int false "Filter by Student ID"
// @Param teacher_id query int false "Filter by Teacher ID"
// @Param activity_id query int false "Filter by Activity ID"
// @Param status query string false "Filter by Status (CREATED, SENDED, APPROVED, REJECTED)"
// @Param limit query int false "Limit for pagination" default(10)
// @Param offset query int false "Offset for pagination" default(0)
// @Success 200 {array} models.Record "List of records retrieved successfully"
// @Failure 400 {object} ErrorResponse "Invalid query parameters"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 403 {object} ErrorResponse "Forbidden (insufficient permissions)"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /records [get]
func (c *RecordController) GetAllRecords(ctx *gin.Context) {
	claims, ok := middlewares.GetUserClaimsFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Message: "User claims not found in context"})
		return
	}

	// Authorization:
	// SAMA_CREW can fetch all records.
	// ADMIN can fetch records for their school.
	// TCH can fetch records for their school or where they are the teacher.
	// STD can only fetch their own records.
	var filterSchoolID, filterStudentID, filterTeacherID, filterActivityID uint
	var filterStatus string

	// Parse query parameters
	if sID, err := strconv.ParseUint(ctx.DefaultQuery("school_id", "0"), 10, 64); err == nil {
		filterSchoolID = uint(sID)
	}
	if stID, err := strconv.ParseUint(ctx.DefaultQuery("student_id", "0"), 10, 64); err == nil {
		filterStudentID = uint(stID)
	}
	if tID, err := strconv.ParseUint(ctx.DefaultQuery("teacher_id", "0"), 10, 64); err == nil {
		filterTeacherID = uint(tID)
	}
	if aID, err := strconv.ParseUint(ctx.DefaultQuery("activity_id", "0"), 10, 64); err == nil {
		filterActivityID = uint(aID)
	}
	filterStatus = ctx.DefaultQuery("status", "")

	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(ctx.DefaultQuery("offset", "0"))

	// Apply authorization filtering
	switch claims.Role {
	case "STD":
		// Student can only see their own records
		if filterStudentID != 0 && filterStudentID != claims.UserID {
			ctx.JSON(http.StatusForbidden, ErrorResponse{Message: "Forbidden: Students can only view their own records."})
			return
		}
		filterStudentID = claims.UserID
	case "TCH":
		// Teacher can see records in their school, or where they are the teacher
		// If a school_id filter is provided, it must match their school_id
		if filterSchoolID != 0 && filterSchoolID != claims.SchoolID {
			ctx.JSON(http.StatusForbidden, ErrorResponse{Message: "Forbidden: Teachers can only view records within their school."})
			return
		}
		filterSchoolID = claims.SchoolID // Always filter by teacher's school
		// If a teacher_id filter is provided, it must match their user_id
		if filterTeacherID != 0 && filterTeacherID != claims.UserID {
			ctx.JSON(http.StatusForbidden, ErrorResponse{Message: "Forbidden: Teachers can only filter by their own teacher ID."})
			return
		}
		// If no teacher_id filter is provided, they can view all records in their school.
		// If filterTeacherID is 0, it means no specific teacher filter was requested, so we don't add it.
	case "ADMIN":
		// Admin can see records in their school
		if filterSchoolID != 0 && filterSchoolID != claims.SchoolID {
			ctx.JSON(http.StatusForbidden, ErrorResponse{Message: "Forbidden: Admins can only view records within their school."})
			return
		}
		filterSchoolID = claims.SchoolID // Always filter by admin's school
	case "SAMA_CREW":
		// Sama Crew can see all records, no additional filtering needed based on their claims
	default:
		ctx.JSON(http.StatusForbidden, ErrorResponse{Message: "Forbidden: Insufficient permissions to list records"})
		return
	}

	records, err := c.recordService.GetAllRecords(
		filterSchoolID, filterStudentID, filterTeacherID, filterActivityID,
		filterStatus,
		limit, offset,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Failed to retrieve records: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, records)
}

// UpdateRecord handles updating an existing record.
// @Summary Update a record
// @Description Update an existing record by ID. Requires relevant student/teacher/admin, or Sama Crew role.
// @Tags Records
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Record ID to update"
// @Param record body UpdateRecordRequest true "Record update details"
// @Success 200 {object} models.Record "Record updated successfully"
// @Failure 400 {object} ErrorResponse "Invalid request payload or validation error"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 403 {object} ErrorResponse "Forbidden (insufficient permissions or not authorized for this record)"
// @Failure 404 {object} ErrorResponse "Record not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /record/{id} [put]
func (c *RecordController) UpdateRecord(ctx *gin.Context) {
	claims, ok := middlewares.GetUserClaimsFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Message: "User claims not found in context"})
		return
	}

	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid record ID"})
		return
	}

	var req UpdateRecordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid request payload: " + err.Error()})
		return
	}

	// Fetch existing record for authorization and to apply updates
	recordToUpdate, err := c.recordService.GetRecordByID(uint(id))
	if err != nil {
		if err.Error() == fmt.Sprintf("record with ID %d not found", id) {
			ctx.JSON(http.StatusNotFound, ErrorResponse{Message: err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Failed to retrieve record for update: " + err.Error()})
		return
	}

	// Authorization:
	// SAMA_CREW can update any record.
	// ADMIN can update records in their school.
	// TCH can update records they are assigned to or for students in their school.
	// STD can update their own records (e.g., changing status from CREATED to SENDED).
	isAuthorized := false
	// if claims.Role == "SAMA_CREW" {
	// 	isAuthorized = true
	// } else if claims.Role == "STD" && claims.UserID == recordToUpdate.StudentID {
	// 	isAuthorized = true
	// } else if claims.Role == "TCH" && (claims.UserID == recordToUpdate.TeacherID || claims.SchoolID == recordToUpdate.SchoolID) {
	// 	isAuthorized = true
	// } else if claims.Role == "ADMIN" && claims.SchoolID == recordToUpdate.SchoolID {
	// 	isAuthorized = true
	// }

	if !isAuthorized {
		ctx.JSON(http.StatusForbidden, ErrorResponse{Message: "Forbidden: Not authorized to update this record."})
		return
	}

	// Apply updates from request to the fetched record model
	if req.ActivityID != nil {
		recordToUpdate.ActivityID = *req.ActivityID
	}
	if req.Data != nil {
		// if err := json.Unmarshal(req.Data, &recordToUpdate.Data); err != nil {
		// 	ctx.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid data JSON: " + err.Error()})
		// 	return
		// }
	}
	if req.Advise != nil {
		recordToUpdate.Advise = *req.Advise
	}
	if req.StudentID != nil {
		recordToUpdate.StudentID = *req.StudentID
	}
	if req.TeacherID != nil {
		recordToUpdate.TeacherID = *req.TeacherID
	}
	if req.SchoolYear != nil {
		recordToUpdate.SchoolYear = *req.SchoolYear
	}
	if req.Semester != nil {
		recordToUpdate.Semester = *req.Semester
	}
	if req.Amount != nil {
		recordToUpdate.Amount = *req.Amount
	}
	if req.Status != nil {
		recordToUpdate.Status = *req.Status
	}

	if err := c.recordService.UpdateRecord(recordToUpdate, claims.UserID); err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Failed to update record: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, recordToUpdate)
}

// DeleteRecord handles deleting a record.
// @Summary Delete a record
// @Description Delete a record by ID. Requires relevant teacher/admin, or Sama Crew role.
// @Tags Records
// @Security BearerAuth
// @Produce json
// @Param id path int true "Record ID to delete"
// @Success 204 "Record deleted successfully"
// @Failure 400 {object} ErrorResponse "Invalid record ID"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 403 {object} ErrorResponse "Forbidden (insufficient permissions or not authorized for this record)"
// @Failure 404 {object} ErrorResponse "Record not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /record/{id} [delete]
func (c *RecordController) DeleteRecord(ctx *gin.Context) {
	_, ok := middlewares.GetUserClaimsFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Message: "User claims not found in context"})
		return
	}

	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid record ID"})
		return
	}

	// Fetch existing record for authorization
	_, err = c.recordService.GetRecordByID(uint(id))
	if err != nil {
		if err.Error() == fmt.Sprintf("record with ID %d not found", id) {
			ctx.JSON(http.StatusNotFound, ErrorResponse{Message: err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Failed to retrieve record for deletion: " + err.Error()})
		return
	}

	// Authorization:
	// SAMA_CREW can delete any record.
	// ADMIN can delete records in their school.
	// TCH can delete records they are assigned to or for students in their school.
	// Students typically cannot delete records.
	isAuthorized := false
	// if claims.Role == "SAMA_CREW" {
	// 	isAuthorized = true
	// } else if claims.Role == "TCH" && (claims.UserID == recordToDelete.TeacherID || claims.SchoolID == recordToDelete.SchoolID) {
	// 	isAuthorized = true
	// } else if claims.Role == "ADMIN" && claims.SchoolID == recordToDelete.SchoolID {
	// 	isAuthorized = true
	// }

	if !isAuthorized {
		ctx.JSON(http.StatusForbidden, ErrorResponse{Message: "Forbidden: Not authorized to delete this record."})
		return
	}

	if err := c.recordService.DeleteRecord(uint(id)); err != nil {
		if err.Error() == fmt.Sprintf("record with ID %d not found for deletion", id) {
			ctx.JSON(http.StatusNotFound, ErrorResponse{Message: err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Failed to delete record: " + err.Error()})
		return
	}

	ctx.Status(http.StatusNoContent) // 204 No Content for successful deletion
}
