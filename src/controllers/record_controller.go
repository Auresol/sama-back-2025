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
	Amount     int                    `json:"amount" binding:"required" example:"5"`
}

// UpdateRecordRequest defines the request body for updating an existing record.
type UpdateRecordRequest struct {
	Data   map[string]interface{} `json:"data" binding:"required" swaggertype:"object,string" example:"field:test"`
	Amount int                    `json:"amount" binding:"reqiured,gt=0" example:"7"`
}

// UpdateRecordRequest defines the request body for updating an existing record.
type SendRecordRequest struct {
	TeacherID uint `json:"teacher_id" binding:"required" example:"1"`
}

// UpdateRecordRequest defines the request body for updating an existing record.
type ApproveRecordRequest struct {
	Advice *string `json:"advice" binding:"required" example:"Good jobs"`
}

// UpdateRecordRequest defines the request body for updating an existing record.
type RejectRecordRequest struct {
	Advice *string `json:"advice" binding:"required" example:"Not so good"`
}

type UnsendRecordRequest struct {
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

	// // Authorization: Example - only teachers can create records for now.
	// // You'll need to refine this based on your business logic (e.g., students creating their own records, etc.)
	// if claims.Role != "TCH" && claims.Role != "ADMIN" && claims.Role != "SAMA_CREW" {
	// 	ctx.JSON(http.StatusForbidden, ErrorResponse{Message: "Forbidden: Insufficient permissions to create records"})
	// 	return
	// }

	var req CreateRecordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid request payload: " + err.Error()})
		return
	}

	record := &models.Record{
		ActivityID: req.ActivityID,
		StudentID:  claims.UserID,
		Data:       req.Data,
		Amount:     req.Amount,
		Status:     "CREATED",
	}

	// Pass the authenticated user's ID for status log
	if err := c.recordService.CreateRecord(record, claims.SchoolID, claims.UserID); err != nil {
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
	// if claims.Role != "SAMA_CREW" {
	// 	isAuthorized := false
	// 	// if claims.Role == "STD" && claims.UserID == record.StudentID {
	// 	// 	isAuthorized = true
	// 	// } else if claims.Role == "TCH" && (claims.UserID == record.TeacherID || claims.SchoolID == record.SchoolID) {
	// 	// 	// Teacher can see records they are assigned to, or records for students in their school
	// 	// 	isAuthorized = true
	// 	// }
	// 	// } else if claims.Role == "ADMIN" && claims.SchoolID == record.SchoolID {
	// 	// 	isAuthorized = true
	// 	// }

	// 	if !isAuthorized {
	// 		ctx.JSON(http.StatusForbidden, ErrorResponse{Message: "Forbidden: Not authorized to view this record."})
	// 		return
	// 	}
	// }

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
		filterStudentID, filterTeacherID, filterActivityID,
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
// @Summary Update an existing record
// @Description Update an existing record's data and/or amount. Accessible by relevant student/teacher/admin, or Sama Crew.
// @Tags Records
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Record ID"
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

	recordID, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid record ID in path"})
		return
	}

	var req UpdateRecordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid request payload: " + err.Error()})
		return
	}

	// Fetch existing record for authorization and update
	existingRecord, err := c.recordService.GetRecordByID(uint(recordID))
	if err != nil {
		if err.Error() == fmt.Sprintf("record with ID %d not found", recordID) {
			ctx.JSON(http.StatusNotFound, ErrorResponse{Message: err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Failed to retrieve record for update: " + err.Error()})
		return
	}

	// Authorization logic for updating a record:
	// Example: Student can only update their own records if status is CREATED.
	// Teacher can update records for students in their school if status is CREATED/SENDED.
	// Admin/SAMA_CREW can update any record.
	isAuthorized := true
	// if claims.Role == "SAMA_CREW" || claims.Role == "ADMIN" { // Admins/Sama Crew can edit any record
	// 	isAuthorized = true
	// } else if claims.Role == "STD" && claims.UserID == existingRecord.StudentID && existingRecord.Status == "CREATED" {
	// 	isAuthorized = true
	// } else if claims.Role == "TCH" && claims.SchoolID == existingRecord.SchoolID && (existingRecord.Status == "CREATED" || existingRecord.Status == "SENDED") {
	// 	isAuthorized = true
	// }

	if !isAuthorized {
		ctx.JSON(http.StatusForbidden, ErrorResponse{Message: "Forbidden: Not authorized to update this record."})
		return
	}

	// Update the record fields
	existingRecord.Data = req.Data
	existingRecord.Amount = req.Amount

	// Pass the authenticated user's ID for status log
	if err := c.recordService.UpdateRecord(existingRecord, claims.UserID); err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Failed to update record: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, existingRecord)
}

// DeleteRecord handles deleting a record.
// @Summary Delete a record
// @Description Delete a record by ID. Requires relevant teacher/admin, or Sama Crew role.
// @Tags Records
// @Security BearerAuth
// @Produce json
// @Param id path int true "Record ID to delete"
// @Success 204 {object} SuccessfulResponse "Record deleted successfully"
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
	isAuthorized := true
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

// SendRecord handles sending a record for approval.
// @Summary Send a record
// @Description Change the status of a record to 'SENDED'.
// @Tags Records
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Record ID"
// @Param record body SendRecordRequest true "Teacher ID to send to"
// @Success 200 {object} models.Record "Record sent successfully"
// @Failure 400 {object} ErrorResponse "Invalid request payload or validation error"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 403 {object} ErrorResponse "Forbidden (insufficient permissions or record not in creatable status)"
// @Failure 404 {object} ErrorResponse "Record not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /record/{id}/send [patch]
func (c *RecordController) SendRecord(ctx *gin.Context) {
	claims, ok := middlewares.GetUserClaimsFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Message: "User claims not found in context"})
		return
	}

	recordID, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid record ID in path"})
		return
	}

	var req SendRecordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid request payload: " + err.Error()})
		return
	}

	// Fetch existing record for authorization and status check
	existingRecord, err := c.recordService.GetRecordByID(uint(recordID))
	if err != nil {
		if err.Error() == fmt.Sprintf("record with ID %d not found", recordID) {
			ctx.JSON(http.StatusNotFound, ErrorResponse{Message: err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Failed to retrieve record for sending: " + err.Error()})
		return
	}

	// Authorization & Status Check:
	// Only the student who owns the record, if status is 'CREATED', can send it.
	// Or ADMIN/SAMA_CREW can send any record.
	isAuthorized := false
	if claims.Role == "SAMA_CREW" || claims.Role == "ADMIN" {
		isAuthorized = true
	} else if claims.Role == "STD" && claims.UserID == existingRecord.StudentID && existingRecord.Status == "CREATED" {
		isAuthorized = true
	}

	if !isAuthorized {
		ctx.JSON(http.StatusForbidden, ErrorResponse{Message: "Forbidden: Not authorized to send this record, or record is not in 'CREATED' status."})
		return
	}

	// Call service method to change status to SENDED
	if err := c.recordService.SendRecord(uint(recordID), req.TeacherID, claims.UserID); err != nil {
		if err.Error() == fmt.Sprintf("record %d cannot be sent: invalid status", recordID) { // Example of a specific service error
			ctx.JSON(http.StatusBadRequest, ErrorResponse{Message: err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Failed to send record: " + err.Error()})
		return
	}

	// Retrieve the updated record to return
	updatedRecord, err := c.recordService.GetRecordByID(uint(recordID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Failed to retrieve updated record: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, updatedRecord)
}

// ApproveRecord handles approving a record.
// @Summary Approve a record
// @Description Change the status of a record to 'APPROVED'. Requires teacher or admin role.
// @Tags Records
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Record ID"
// @Param record body ApproveRecordRequest true "Optional advice for approval"
// @Success 200 {object} models.Record "Record approved successfully"
// @Failure 400 {object} ErrorResponse "Invalid request payload or validation error"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 403 {object} ErrorResponse "Forbidden (insufficient permissions or record not in sendable status)"
// @Failure 404 {object} ErrorResponse "Record not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /record/{id}/approve [patch]
func (c *RecordController) ApproveRecord(ctx *gin.Context) {
	claims, ok := middlewares.GetUserClaimsFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Message: "User claims not found in context"})
		return
	}

	recordID, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid record ID in path"})
		return
	}

	var req ApproveRecordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid request payload: " + err.Error()})
		return
	}

	// Fetch existing record for authorization and status check
	existingRecord, err := c.recordService.GetRecordByID(uint(recordID))
	if err != nil {
		if err.Error() == fmt.Sprintf("record with ID %d not found", recordID) {
			ctx.JSON(http.StatusNotFound, ErrorResponse{Message: err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Failed to retrieve record for approval: " + err.Error()})
		return
	}

	// Authorization & Status Check:
	// Only the assigned teacher or admin/SAMA_CREW can approve.
	// Record must be in 'SENDED' status.
	isAuthorized := false
	if claims.Role == "SAMA_CREW" || claims.Role == "ADMIN" {
		isAuthorized = true
	} else if claims.Role == "TCH" && claims.UserID == *existingRecord.TeacherID && existingRecord.Status == "SENDED" {
		isAuthorized = true
	}

	if !isAuthorized {
		ctx.JSON(http.StatusForbidden, ErrorResponse{Message: "Forbidden: Not authorized to approve this record, or record is not in 'SENDED' status."})
		return
	}

	// Call service method to change status to APPROVED
	if err := c.recordService.ApproveRecord(uint(recordID), req.Advice, claims.UserID); err != nil {
		if err.Error() == fmt.Sprintf("record %d cannot be approved: invalid status", recordID) { // Example of a specific service error
			ctx.JSON(http.StatusBadRequest, ErrorResponse{Message: err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Failed to approve record: " + err.Error()})
		return
	}

	// Retrieve the updated record to return
	updatedRecord, err := c.recordService.GetRecordByID(uint(recordID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Failed to retrieve updated record: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, updatedRecord)
}

// RejectRecord handles rejecting a record.
// @Summary Reject a record
// @Description Change the status of a record to 'REJECTED'. Requires teacher or admin role.
// @Tags Records
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Record ID"
// @Param record body RejectRecordRequest true "Optional advice for rejection"
// @Success 200 {object} models.Record "Record rejected successfully"
// @Failure 400 {object} ErrorResponse "Invalid request payload or validation error"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 403 {object} ErrorResponse "Forbidden (insufficient permissions or record not in sendable status)"
// @Failure 404 {object} ErrorResponse "Record not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /record/{id}/reject [patch]
func (c *RecordController) RejectRecord(ctx *gin.Context) {
	claims, ok := middlewares.GetUserClaimsFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Message: "User claims not found in context"})
		return
	}

	recordID, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid record ID in path"})
		return
	}

	var req RejectRecordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid request payload: " + err.Error()})
		return
	}

	// Fetch existing record for authorization and status check
	existingRecord, err := c.recordService.GetRecordByID(uint(recordID))
	if err != nil {
		if err.Error() == fmt.Sprintf("record with ID %d not found", recordID) {
			ctx.JSON(http.StatusNotFound, ErrorResponse{Message: err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Failed to retrieve record for rejection: " + err.Error()})
		return
	}

	// Authorization & Status Check:
	// Only the assigned teacher or admin/SAMA_CREW can reject.
	// Record must be in 'SENDED' status.
	isAuthorized := false
	if claims.Role == "SAMA_CREW" || claims.Role == "ADMIN" {
		isAuthorized = true
	} else if claims.Role == "TCH" && claims.UserID == *existingRecord.TeacherID && existingRecord.Status == "SENDED" {
		isAuthorized = true
	}

	if !isAuthorized {
		ctx.JSON(http.StatusForbidden, ErrorResponse{Message: "Forbidden: Not authorized to reject this record, or record is not in 'SENDED' status."})
		return
	}

	// Call service method to change status to REJECTED
	if err := c.recordService.RejectRecord(uint(recordID), req.Advice, claims.UserID); err != nil {
		if err.Error() == fmt.Sprintf("record %d cannot be rejected: invalid status", recordID) { // Example of a specific service error
			ctx.JSON(http.StatusBadRequest, ErrorResponse{Message: err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Failed to reject record: " + err.Error()})
		return
	}

	// Retrieve the updated record to return
	updatedRecord, err := c.recordService.GetRecordByID(uint(recordID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Failed to retrieve updated record: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, updatedRecord)
}

// UnsendRecord handles unsending a record.
// @Summary Unsend a record
// @Description Change the status of a record back to 'CREATED' from 'SENDED'.
// @Tags Records
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Record ID"
// @Param record body UnsendRecordRequest true "Empty request body as ID is in path"
// @Success 200 {object} models.Record "Record unsent successfully"
// @Failure 400 {object} ErrorResponse "Invalid request payload or validation error"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 403 {object} ErrorResponse "Forbidden (insufficient permissions or record not in sendable status)"
// @Failure 404 {object} ErrorResponse "Record not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /record/{id}/unsend [patch]
func (c *RecordController) UnsendRecord(ctx *gin.Context) {
	claims, ok := middlewares.GetUserClaimsFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Message: "User claims not found in context"})
		return
	}

	recordID, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid record ID in path"})
		return
	}

	// var req UnsendRecordRequest // Still bind to check for empty/malformed body if needed, though no fields
	// if err := ctx.ShouldBindJSON(&req); err != nil {
	// 	// Depending on your gin setup, an empty JSON body might still trigger an error here.
	// 	// If you expect a truly empty body ({}), this check might be too strict.
	// 	// For PATCH, it's safer to always allow an empty body for request structs with no fields.
	// 	// If `binding:"required"` was on internal fields, it'd still be relevant.
	// 	// Given UnsendRecordRequest has no fields, this `ShouldBindJSON` check might be simplified
	// 	// or even removed if you truly expect an empty body and don't need validation for it.
	// 	// For now, keeping it to be consistent with other methods' error handling pattern.
	// 	ctx.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid request payload: " + err.Error()})
	// 	return
	// }

	// Fetch existing record for authorization and status check
	existingRecord, err := c.recordService.GetRecordByID(uint(recordID))
	if err != nil {
		if err.Error() == fmt.Sprintf("record with ID %d not found", recordID) {
			ctx.JSON(http.StatusNotFound, ErrorResponse{Message: err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Failed to retrieve record for unsending: " + err.Error()})
		return
	}

	// Authorization & Status Check:
	// Only the student who sent the record (if status is 'SENDED'), or ADMIN/SAMA_CREW can unsend it.
	isAuthorized := false
	if claims.Role == "SAMA_CREW" || claims.Role == "ADMIN" {
		isAuthorized = true
	} else if claims.Role == "STD" && claims.UserID == existingRecord.StudentID && existingRecord.Status == "SENDED" {
		isAuthorized = true
	}

	if !isAuthorized {
		ctx.JSON(http.StatusForbidden, ErrorResponse{Message: "Forbidden: Not authorized to unsend this record, or record is not in 'SENDED' status."})
		return
	}

	// // Call service method to change status to CREATED
	if err := c.recordService.UnsendRecord(uint(recordID), claims.UserID); err != nil {
		if err.Error() == fmt.Sprintf("record %d cannot be unsent: invalid status", recordID) { // Example of a specific service error
			ctx.JSON(http.StatusBadRequest, ErrorResponse{Message: err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Failed to unsend record: " + err.Error()})
		return
	}

	// Retrieve the updated record to return
	updatedRecord, err := c.recordService.GetRecordByID(uint(recordID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Failed to retrieve updated record: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, updatedRecord)
}
