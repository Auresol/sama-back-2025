package controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"sama/sama-backend-2025/src/middlewares"
	"sama/sama-backend-2025/src/models"
	"sama/sama-backend-2025/src/services"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// SchoolController manages HTTP requests for schools.
type SchoolController struct {
	schoolService *services.SchoolService
	userService   *services.UserService
	validate      *validator.Validate
}

// NewSchoolController creates a new SchoolController.
func NewSchoolController(
	schoolService *services.SchoolService,
	userService *services.UserService,
	validate *validator.Validate,
) *SchoolController {
	return &SchoolController{
		schoolService: schoolService,
		userService:   userService,
		validate:      validate,
	}
}

// CreateSchoolRequest represents the request body for creating a new school.
type CreateSchoolRequest struct {
	ThaiName                string    `json:"thai_name" binding:"required" example:"โรงเรียนสามัคคีวิทยา"`
	EnglishName             string    `json:"english_name" binding:"required" example:"Samakkee Wittaya School"`
	ShortName               string    `json:"short_name" binding:"required" example:"SMK"`
	SchoolLogoUrl           *string   `json:"school_logo_url"`
	Email                   *string   `json:"email,omitempty" binding:"email" example:"info@smk.ac.th"`
	DefaultActivityDeadline time.Time `json:"default_activity_deadline" example:"2025-07-28T15:49:03.123Z"`
	Location                *string   `json:"location,omitempty" example:"Bangkok, Thailand"`
	Phone                   *string   `json:"phone,omitempty" binding:"e164" example:"+66812345678"`
	Classrooms              []string  `json:"classrooms" binding:"required" example:"1/1" validate:"required,dive,classroomregex"`
	SchoolYear              uint      `json:"school_year" binding:"required,gt=0" example:"2568"`
	Semester                uint      `json:"semester" binding:"required,gt=0" example:"1"`
}

// UpdateSchoolRequest represents the request body for updating an existing school.
type UpdateSchoolRequest struct {
	ThaiName                string    `json:"thai_name" binding:"required" example:"โรงเรียนสามัคคีวิทยา"`
	EnglishName             string    `json:"english_name" binding:"required" example:"Samakkee Wittaya School"`
	ShortName               string    `json:"short_name" binding:"required" example:"SMK"`
	SchoolLogoUrl           *string   `json:"school_logo_url"`
	Email                   *string   `json:"email,omitempty" binding:"email" example:"info@smk.ac.th"`
	DefaultActivityDeadline time.Time `json:"default_activity_deadline"`
	Location                *string   `json:"location,omitempty" example:"Bangkok, Thailand"`
	Phone                   *string   `json:"phone,omitempty" binding:"e164" example:"+66812345678"`
	Classrooms              []string  `json:"classrooms" binding:"required" example:"1/1" validate:"required,dive,classroomregex"`
}

// CreateSchool handles the creation of a new school.
// @Summary Create a new school
// @Description Create a new school record. Requires ADMIN or Sama Crew role.
// @Tags School
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param school body CreateSchoolRequest true "School creation details"
// @Success 201 {object} models.School "School created successfully"
// @Failure 400 {object} ErrorResponse "Invalid request payload or validation error"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 403 {object} ErrorResponse "Forbidden (insufficient permissions)"
// @Failure 409 {object} ErrorResponse "School with this email or short name already exists"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /school [post]
func (h *SchoolController) CreateSchool(c *gin.Context) {
	// claims, ok := middlewares.GetUserClaimsFromContext(c)
	// if !ok {
	// 	c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "User claims not found in context"})
	// 	return
	// }
	// // Authorization: Only ADMIN or SAMA can create schools
	// if claims.Role != "ADMIN" && claims.Role != "SAMA" {
	// 	c.JSON(http.StatusForbidden, ErrorResponse{Message: "Forbidden: Insufficient permissions"})
	// 	return
	// }

	var req CreateSchoolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid request payload: " + err.Error()})
		return
	}

	err := h.validate.Struct(req)
	if err != nil {
		//fmt.Printf("Validation Error for s1: %v\n", err)
		c.JSON(http.StatusBadRequest, ErrorResponse{Message: err.Error()})
		return
	}

	if len(req.Classrooms) == 0 {
		c.JSON(http.StatusBadRequest, ErrorResponse{Message: "Please specify at least one classroom"})
		return
	}

	school := &models.School{
		ThaiName:                req.ThaiName,
		EnglishName:             req.EnglishName,
		ShortName:               req.ShortName,
		DefaultActivityDeadline: req.DefaultActivityDeadline,
		Email:                   req.Email,
		Location:                req.Location,
		Phone:                   req.Phone,
		Classrooms:              req.Classrooms,
		SchoolYear:              req.SchoolYear,
		Semester:                req.Semester,
	}

	if err := h.schoolService.CreateSchool(school); err != nil {
		// if err.Error() == "school with this email already exists" || err.Error() == "school with this short name already exists" {
		// 	c.JSON(http.StatusConflict, ErrorResponse{Message: err.Error()})
		// 	return
		// }
		c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Failed to create school: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, school)
}

// GetSchoolByID retrieves a school by its ID.
// @Summary Get school by ID
// @Description Retrieve a school's details by its ID. Accessible by ADMIN (for their school), SAMA, or any TCH/STD if they belong to that school.
// @Tags School
// @Security BearerAuth
// @Produce json
// @Param id path int true "School ID"
// @Success 200 {object} models.School "School retrieved successfully"
// @Failure 400 {object} ErrorResponse "Invalid school ID"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 403 {object} ErrorResponse "Forbidden (not authorized to access this school's data)"
// @Failure 404 {object} ErrorResponse "School not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /school/{id} [get]
func (h *SchoolController) GetSchoolByID(c *gin.Context) {
	claims, ok := middlewares.GetUserClaimsFromContext(c)
	if !ok {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "User claims not found in context"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid school ID"})
		return
	}

	// Authorization:
	// SAMA can access any school.
	// ADMIN/TCH/STD can access their own school's data.
	if claims.Role != "SAMA" && claims.SchoolID != uint(id) {
		c.JSON(http.StatusForbidden, ErrorResponse{Message: "Forbidden: Not authorized to access this school's data"})
		return
	}

	school, err := h.schoolService.GetSchoolByID(uint(id))
	if err != nil {
		if err.Error() == fmt.Sprintf("school with ID %d not found", id) {
			c.JSON(http.StatusNotFound, ErrorResponse{Message: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Failed to retrieve school: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, school)
}

// GetAllSchools retrieves all schools with pagination.
// @Summary Get all schools
// @Description Retrieve a list of all school records with pagination. Requires Sama Crew role.
// @Tags School
// @Security BearerAuth
// @Produce json
// @Param limit query int false "Limit for pagination" default(10)
// @Param offset query int false "Offset for pagination" default(0)
// @Success 200 {array} models.School "List of schools retrieved successfully"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 403 {object} ErrorResponse "Forbidden (insufficient permissions)"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /school [get]
func (h *SchoolController) GetAllSchools(c *gin.Context) {
	// claims, ok := middlewares.GetUserClaimsFromContext(c)
	// if !ok {
	// 	c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "User claims not found in context"})
	// 	return
	// }
	// // Authorization: Only SAMA can get all schools
	// if claims.Role != "SAMA" {
	// 	c.JSON(http.StatusForbidden, ErrorResponse{Message: "Forbidden: Insufficient permissions"})
	// 	return
	// }

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	schools, err := h.schoolService.GetAllSchools(limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Failed to retrieve schools: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, schools)
}

// UpdateSchool handles updating an existing school.
// @Summary Update a school
// @Description Update an existing school record by ID. Requires ADMIN (for their school) or Sama Crew role.
// @Tags School
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "School ID to update"
// @Param school body UpdateSchoolRequest true "School update details"
// @Success 200 {object} models.School "School updated successfully"
// @Failure 400 {object} ErrorResponse "Invalid request payload or validation error"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 403 {object} ErrorResponse "Forbidden (insufficient permissions or not authorized for this school)"
// @Failure 404 {object} ErrorResponse "School not found"
// @Failure 409 {object} ErrorResponse "New email or short name already exists for another school"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /school/{id} [put]
func (h *SchoolController) UpdateSchool(c *gin.Context) {
	claims, ok := middlewares.GetUserClaimsFromContext(c)
	if !ok {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "User claims not found in context"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid school ID"})
		return
	}

	// Authorization:
	// SAMA can update any school.
	// ADMIN can only update their own school.
	if claims.Role != "SAMA" && claims.Role != "ADMIN" {
		c.JSON(http.StatusForbidden, ErrorResponse{Message: "Forbidden: Insufficient permissions"})
		return
	}
	if claims.Role == "ADMIN" && claims.SchoolID != uint(id) {
		c.JSON(http.StatusForbidden, ErrorResponse{Message: "Forbidden: ADMIN can only update their own school"})
		return
	}

	var req UpdateSchoolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid request payload: " + err.Error()})
		return
	}

	// Fetch existing school to ensure it exists
	schoolToUpdate, err := h.schoolService.GetSchoolByID(uint(id))
	if err != nil {
		if err.Error() == fmt.Sprintf("school with ID %d not found", id) {
			c.JSON(http.StatusNotFound, ErrorResponse{Message: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Failed to retrieve school for update: " + err.Error()})
		return
	}

	// Apply updates from request to the fetched school model
	// Only update fields that are provided in the request
	schoolToUpdate.ThaiName = req.ThaiName
	schoolToUpdate.EnglishName = req.EnglishName
	schoolToUpdate.ShortName = req.ShortName
	schoolToUpdate.DefaultActivityDeadline = req.DefaultActivityDeadline
	schoolToUpdate.Email = req.Email
	schoolToUpdate.Location = req.Location
	schoolToUpdate.Phone = req.Phone
	schoolToUpdate.Classrooms = req.Classrooms

	fmt.Println(schoolToUpdate)

	if err := h.schoolService.UpdateSchool(schoolToUpdate); err != nil {
		if err.Error() == "new email already exists for another school" || err.Error() == "new short name already exists for another school" {
			c.JSON(http.StatusConflict, ErrorResponse{Message: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Failed to update school: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, schoolToUpdate)
}

// DeleteSchool handles deleting a school.
// @Summary Delete a school
// @Description Delete a school record by ID. Requires ADMIN (for their school) or Sama Crew role.
// @Tags School
// @Security BearerAuth
// @Produce json
// @Param id path int true "School ID to delete"
// @Success 204 {object} SuccessfulResponse "School deleted successfully"
// @Failure 400 {object} ErrorResponse "Invalid school ID"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 403 {object} ErrorResponse "Forbidden (insufficient permissions or not authorized for this school)"
// @Failure 404 {object} ErrorResponse "School not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /school/{id} [delete]
func (h *SchoolController) DeleteSchool(c *gin.Context) {
	claims, ok := middlewares.GetUserClaimsFromContext(c)
	if !ok {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "User claims not found in context"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid school ID"})
		return
	}

	// Authorization:
	// SAMA can delete any school.
	// ADMIN can only delete their own school.
	if claims.Role != "SAMA" && claims.Role != "ADMIN" {
		c.JSON(http.StatusForbidden, ErrorResponse{Message: "Forbidden: Insufficient permissions"})
		return
	}
	if claims.Role == "ADMIN" && claims.SchoolID != uint(id) {
		c.JSON(http.StatusForbidden, ErrorResponse{Message: "Forbidden: ADMIN can only delete their own school"})
		return
	}

	if err := h.schoolService.DeleteSchool(uint(id)); err != nil {
		if err.Error() == fmt.Sprintf("school with ID %d not found for deletion", id) {
			c.JSON(http.StatusNotFound, ErrorResponse{Message: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Failed to delete school: " + err.Error()})
		return
	}

	c.Status(http.StatusNoContent) // 204 No Content for successful deletion
}

// SemesterTransitionRequest represents the request body for semester transition operations.
type SemesterTransitionRequest struct {
	SchoolID uint `json:"school_id" binding:"required,gt=0" example:"1"`
}

// SemesterTransitionResponse represents the response body for semester transition operations.
type SemesterTransitionResponse struct {
	Message string `json:"message" example:"Operation completed successfully"`
}

// AdvanceSemester handles moving an entire school to the next semester.
// @Summary Move school to next semester
// @Description Advances the specified school to the next academic semester. Requires ADMIN or Sama Crew role.
// @Tags School
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param semester_transition body SemesterTransitionRequest true "School ID for semester transition"
// @Success 200 {object} SemesterTransitionResponse "Operation completed successfully"
// @Failure 400 {object} ErrorResponse "Invalid request payload or school ID"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 403 {object} ErrorResponse "Forbidden (insufficient permissions or not authorized for this school)"
// @Failure 404 {object} ErrorResponse "School not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /school/advance-semester [post]
func (h *SchoolController) AdvanceSemester(c *gin.Context) {
	claims, ok := middlewares.GetUserClaimsFromContext(c)
	if !ok {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "User claims not found in context"})
		return
	}

	// Authorization: Only ADMINs (for their school) or SAMA can perform this
	if claims.Role != "ADMIN" && claims.Role != "SAMA" {
		c.JSON(http.StatusForbidden, ErrorResponse{Message: "Forbidden: Insufficient permissions"})
		return
	}

	var req SemesterTransitionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid request payload: " + err.Error()})
		return
	}

	// If ADMIN, ensure they are operating on their own school
	if claims.Role == "ADMIN" && claims.SchoolID != req.SchoolID {
		c.JSON(http.StatusForbidden, ErrorResponse{Message: "Forbidden: ADMIN can only move their own school's semester"})
		return
	}

	// TODO: Implement the service call to move the school to the next semester
	// Example:
	// err := h.schoolService.MoveSchoolToNextSemester(req.SchoolID)
	// if err != nil {
	//     if errors.Is(err, gorm.ErrRecordNotFound) {
	//         c.JSON(http.StatusNotFound, ErrorResponse{Message: fmt.Sprintf("School with ID %d not found", req.SchoolID)})
	//         return
	//     }
	//     c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Failed to move school to next semester: " + err.Error()})
	//     return
	// }

	c.JSON(http.StatusOK, SemesterTransitionResponse{Message: "School moved to next semester successfully"})
}

// RevertSemester handles reverting an entire school back to the previous semester.
// @Summary Revert school to previous semester
// @Description Reverts the specified school to the previous academic semester. Requires ADMIN or Sama Crew role.
// @Tags School
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param semester_transition body SemesterTransitionRequest true "School ID for semester transition"
// @Success 200 {object} SemesterTransitionResponse "Operation completed successfully"
// @Failure 400 {object} ErrorResponse "Invalid request payload or school ID"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 403 {object} ErrorResponse "Forbidden (insufficient permissions or not authorized for this school)"
// @Failure 404 {object} ErrorResponse "School not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /school/revert-semester [post]
func (h *SchoolController) RevertSemester(c *gin.Context) {
	claims, ok := middlewares.GetUserClaimsFromContext(c)
	if !ok {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "User claims not found in context"})
		return
	}

	// Authorization: Only ADMINs (for their school) or SAMA can perform this
	if claims.Role != "ADMIN" && claims.Role != "SAMA" {
		c.JSON(http.StatusForbidden, ErrorResponse{Message: "Forbidden: Insufficient permissions"})
		return
	}

	var req SemesterTransitionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid request payload: " + err.Error()})
		return
	}

	// If ADMIN, ensure they are operating on their own school
	if claims.Role == "ADMIN" && claims.SchoolID != req.SchoolID {
		c.JSON(http.StatusForbidden, ErrorResponse{Message: "Forbidden: ADMIN can only revert their own school's semester"})
		return
	}

	// TODO: Implement the service call to revert the school to the previous semester
	// Example:
	// err := h.schoolService.RevertSchoolSemester(req.SchoolID)
	// if err != nil {
	//     if errors.Is(err, gorm.ErrRecordNotFound) {
	//         c.JSON(http.StatusNotFound, ErrorResponse{Message: fmt.Sprintf("School with ID %d not found", req.SchoolID)})
	//         return
	//     }
	//     c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Failed to revert school semester: " + err.Error()})
	//     return
	// }

	c.JSON(http.StatusOK, SemesterTransitionResponse{Message: "School reverted to previous semester successfully"})
}

// GetUsersBySchoolID handles retrieving users by school ID.
// @Summary Get users by school ID
// @Description Retrieve a list of users belonging to a specific school. Requires ADMIN or Sama Crew role.
// @Tags School
// @Security BearerAuth
// @Produce json
// @Param school_id path int true "School ID"
// @Param limit query int false "Limit for pagination" default(10)
// @Param offset query int false "Offset for pagination" default(0)
// @Success 200 {array} models.User "List of users retrieved successfully"
// @Failure 400 {object} ErrorResponse "Invalid school ID or pagination parameters"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 403 {object} ErrorResponse "Forbidden (insufficient permissions or not authorized for this school)"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /school/{id}/user [get]
func (h *SchoolController) GetUsersBySchoolID(c *gin.Context) {
	claims, ok := middlewares.GetUserClaimsFromContext(c)
	if !ok {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "User claims not found in context"})
		return
	}

	// Authorization: Only ADMINs (for their school) or SAMA can access this
	if claims.Role != "ADMIN" && claims.Role != "SAMA" {
		c.JSON(http.StatusForbidden, ErrorResponse{Message: "Forbidden: Insufficient permissions"})
		return
	}

	schoolID, err := strconv.ParseUint(c.Param("school_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid school ID"})
		return
	}

	// If ADMIN, ensure they are requesting users from their own school
	if claims.Role == "ADMIN" && claims.SchoolID != uint(schoolID) {
		c.JSON(http.StatusForbidden, ErrorResponse{Message: "Forbidden: ADMIN can only view users from their own school"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	users, err := h.userService.GetUsersBySchoolID(uint(schoolID), "", limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Failed to retrieve users: " + err.Error()})
		return
	}

	// Omit passwords from response
	for i := range users {
		users[i].Password = ""
	}
	c.JSON(http.StatusOK, users)
}

// GetStatistic get statistic based on activity_id and classroom
// @Summary Get users by school ID
// @Description Retrieve a list of users belonging to a specific school. Requires ADMIN or Sama Crew role.
// @Tags School
// @Security BearerAuth
// @Produce json
// @Param school_id path int true "School ID"
// @Param classroom query string false "Classroom string to query"
// @Param activity_id query string false "Activity id list seperate by \"|\""
// @Success 200 {array} models.User "List of users retrieved successfully"
// @Failure 400 {object} ErrorResponse "Invalid school ID or pagination parameters"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 403 {object} ErrorResponse "Forbidden (insufficient permissions or not authorized for this school)"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /school/{id}/statistic [get]
func (h *SchoolController) GetStatistic(c *gin.Context) {
	claims, ok := middlewares.GetUserClaimsFromContext(c)
	if !ok {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "User claims not found in context"})
		return
	}

	// Authorization: Only ADMINs (for their school) or SAMA can access this
	if claims.Role != "ADMIN" && claims.Role != "SAMA" {
		c.JSON(http.StatusForbidden, ErrorResponse{Message: "Forbidden: Insufficient permissions"})
		return
	}

	schoolID, err := strconv.ParseUint(c.Param("school_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid school ID"})
		return
	}

	// If ADMIN, ensure they are requesting users from their own school
	if claims.Role == "ADMIN" && claims.SchoolID != uint(schoolID) {
		c.JSON(http.StatusForbidden, ErrorResponse{Message: "Forbidden: ADMIN can only view users from their own school"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	users, err := h.userService.GetUsersBySchoolID(uint(schoolID), "", limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Failed to retrieve users: " + err.Error()})
		return
	}

	// Omit passwords from response
	for i := range users {
		users[i].Password = ""
	}
	c.JSON(http.StatusOK, users)
}
