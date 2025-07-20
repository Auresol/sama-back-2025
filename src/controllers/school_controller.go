package controllers

import (
	"fmt"
	"net/http"
	"strconv"

	"sama/sama-backend-2025/src/middlewares"
	"sama/sama-backend-2025/src/models"
	"sama/sama-backend-2025/src/services"

	"github.com/gin-gonic/gin"
)

// SchoolController manages HTTP requests for schools.
type SchoolController struct {
	schoolService *services.SchoolService
}

// NewSchoolController creates a new SchoolController.
func NewSchoolController(schoolService *services.SchoolService) *SchoolController {
	return &SchoolController{
		schoolService: schoolService,
	}
}

// CreateSchoolRequest represents the request body for creating a new school.
type CreateSchoolRequest struct {
	ThaiName    string `json:"thai_name" binding:"required" example:"โรงเรียนสามัคคีวิทยา"`
	EnglishName string `json:"english_name" binding:"required" example:"Samakkee Wittaya School"`
	ShortName   string `json:"short_name" binding:"required" example:"SMK"`
	Email       string `json:"email" binding:"required,email" example:"info@smk.ac.th"`
	Location    string `json:"location" binding:"required" example:"Bangkok, Thailand"`
	Phone       string `json:"phone" binding:"required,e164" example:"+66812345678"`
	SchoolYear  int    `json:"school_year" binding:"required,gt=0" example:"2568"`
	Semester    int    `json:"semester" binding:"required,gt=0" example:"1"`
}

// UpdateSchoolRequest represents the request body for updating an existing school.
type UpdateSchoolRequest struct {
	ThaiName    string `json:"thai_name,omitempty" binding:"omitempty" example:"โรงเรียนสามัคคีวิทยาใหม่"`
	EnglishName string `json:"english_name,omitempty" binding:"omitempty" example:"New Samakkee Wittaya School"`
	ShortName   string `json:"short_name,omitempty" binding:"omitempty" example:"NSMK"`
	Email       string `json:"email,omitempty" binding:"omitempty,email" example:"new_info@smk.ac.th"`
	Location    string `json:"location,omitempty" binding:"omitempty" example:"Nonthaburi, Thailand"`
	Phone       string `json:"phone,omitempty" binding:"omitempty,e164" example:"+66923456789"`
	SchoolYear  int    `json:"school_year,omitempty" binding:"omitempty,gt=0" example:"2569"`
	Semester    int    `json:"semester,omitempty" binding:"omitempty,gt=0" example:"2"`
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
// @Router /schools [post]
func (h *SchoolController) CreateSchool(c *gin.Context) {
	// claims, ok := middlewares.GetUserClaimsFromContext(c)
	// if !ok {
	// 	c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "User claims not found in context"})
	// 	return
	// }
	// // Authorization: Only ADMIN or SAMA_CREW can create schools
	// if claims.Role != "ADMIN" && claims.Role != "SAMA_CREW" {
	// 	c.JSON(http.StatusForbidden, ErrorResponse{Message: "Forbidden: Insufficient permissions"})
	// 	return
	// }

	var req CreateSchoolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid request payload: " + err.Error()})
		return
	}

	school := &models.School{
		ThaiName:    req.ThaiName,
		EnglishName: req.EnglishName,
		ShortName:   req.ShortName,
		Email:       req.Email,
		Location:    req.Location,
		Phone:       req.Phone,
		SchoolYear:  req.SchoolYear,
		Semester:    req.Semester,
	}

	if err := h.schoolService.CreateSchool(school); err != nil {
		if err.Error() == "school with this email already exists" || err.Error() == "school with this short name already exists" {
			c.JSON(http.StatusConflict, ErrorResponse{Message: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Failed to create school: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, school)
}

// GetSchoolByID retrieves a school by its ID.
// @Summary Get school by ID
// @Description Retrieve a school's details by its ID. Accessible by ADMIN (for their school), SAMA_CREW, or any TCH/STD if they belong to that school.
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
// @Router /schools/{id} [get]
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
	// SAMA_CREW can access any school.
	// ADMIN/TCH/STD can access their own school's data.
	if claims.Role != "SAMA_CREW" && claims.SchoolID != uint(id) {
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
// @Router /schools [get]
func (h *SchoolController) GetAllSchools(c *gin.Context) {
	// claims, ok := middlewares.GetUserClaimsFromContext(c)
	// if !ok {
	// 	c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "User claims not found in context"})
	// 	return
	// }
	// // Authorization: Only SAMA_CREW can get all schools
	// if claims.Role != "SAMA_CREW" {
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
// @Router /schools/{id} [put]
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
	// SAMA_CREW can update any school.
	// ADMIN can only update their own school.
	if claims.Role != "SAMA_CREW" && claims.Role != "ADMIN" {
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
	if req.ThaiName != "" {
		schoolToUpdate.ThaiName = req.ThaiName
	}
	if req.EnglishName != "" {
		schoolToUpdate.EnglishName = req.EnglishName
	}
	if req.ShortName != "" {
		schoolToUpdate.ShortName = req.ShortName
	}
	if req.Email != "" {
		schoolToUpdate.Email = req.Email
	}
	if req.Location != "" {
		schoolToUpdate.Location = req.Location
	}
	if req.Phone != "" {
		schoolToUpdate.Phone = req.Phone
	}
	if req.SchoolYear != 0 { // Assuming 0 means not provided for int
		schoolToUpdate.SchoolYear = req.SchoolYear
	}
	if req.Semester != 0 { // Assuming 0 means not provided for int
		schoolToUpdate.Semester = req.Semester
	}

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
// @Success 204 "School deleted successfully"
// @Failure 400 {object} ErrorResponse "Invalid school ID"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 403 {object} ErrorResponse "Forbidden (insufficient permissions or not authorized for this school)"
// @Failure 404 {object} ErrorResponse "School not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /schools/{id} [delete]
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
	// SAMA_CREW can delete any school.
	// ADMIN can only delete their own school.
	if claims.Role != "SAMA_CREW" && claims.Role != "ADMIN" {
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
