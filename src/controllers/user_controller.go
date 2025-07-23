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
	"golang.org/x/crypto/bcrypt"
)

// UserController manages HTTP requests for user accounts.
type UserController struct {
	userService *services.UserService
}

// NewUserController creates a new UserController.
func NewUserController(userService *services.UserService) *UserController {
	return &UserController{
		userService: userService,
	}
}

// RegisterRequest represents the request body for user registration.
type RegisterRequest struct {
	UserID    string `json:"user_id" example:"10101"`
	Email     string `json:"email" binding:"required,email" example:"user@example.com"`
	Password  string `json:"password" binding:"required,min=8" example:"Secure_P@ss1"` // Custom validation for password
	Firstname string `json:"firstname" binding:"required" example:"John"`
	Lastname  string `json:"lastname" binding:"required" example:"Doe"`
	Role      string `json:"role" binding:"required,oneof=STD TCH ADMIN" example:"STD"` // Validate against roles
	SchoolID  uint   `json:"school_id" binding:"required,gt=0" example:"1"`
	Phone     string `json:"phone,omitempty" example:"+1234567890"`
	Classroom string `json:"classroom,omitempty" example:"A101"`
	Number    uint   `json:"number,omitempty" example:"1"`
	Language  string `json:"language,omitempty" example:"en"`
}

// LoginRequest represents the request body for user login.
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email" example:"user@example.com"`
	Password string `json:"password" binding:"required" example:"Secure_P@ss1"`
}

// LoginResponse represents the response body for successful login.
type LoginResponse struct {
	Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

// ErrorResponse represents a generic error response.
type ErrorResponse struct {
	Message string `json:"message" example:"Error description"`
}

// RegisterUser handles user registration.
// @Summary Register a new user
// @Description Register a new user account (can be STD, TCH, ADMIN)
// @Tags Account
// @Accept json
// @Produce json
// @Param user body RegisterRequest true "User registration details"
// @Success 201 {object} models.User "User created successfully"
// @Failure 400 {object} ErrorResponse "Invalid request payload or validation error"
// @Failure 409 {object} ErrorResponse "User with this email already exists"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /register [post]
func (h *UserController) RegisterUser(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid request payload: " + err.Error()})
		return
	}

	user := &models.User{
		UserID:    req.UserID,
		Email:     req.Email,
		Password:  req.Password, // Plain password, will be hashed in service
		Firstname: req.Firstname,
		Lastname:  req.Lastname,
		Role:      req.Role,
		SchoolID:  req.SchoolID,
		Phone:     req.Phone,
		Classroom: req.Classroom,
		Number:    req.Number,
		Language:  req.Language,
	}

	if err := h.userService.RegisterUser(user); err != nil {
		if err.Error() == "user with this email already exists" {
			c.JSON(http.StatusConflict, ErrorResponse{Message: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Failed to register user: " + err.Error()})
		return
	}

	// Omit password from response
	user.Password = ""
	c.JSON(http.StatusCreated, user)
}

// Login handles user login and returns a JWT token.
// @Summary Log in a user
// @Description Authenticate user credentials and return a JWT token
// @Tags Account
// @Accept json
// @Produce json
// @Param credentials body LoginRequest true "User login credentials"
// @Success 200 {object} LoginResponse "Successful login with JWT token"
// @Failure 400 {object} ErrorResponse "Invalid request payload or validation error"
// @Failure 401 {object} ErrorResponse "Invalid credentials or account deactivated"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /login [post]
func (h *UserController) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid request payload: " + err.Error()})
		return
	}

	token, err := h.userService.Login(req.Email, req.Password)
	if err != nil {
		if err.Error() == "invalid credentials" || err.Error() == "user account is deactivated" {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Message: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Failed to login: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, LoginResponse{Token: token})
}

// GetMyProfile retrieves the profile of the authenticated user.
// @Summary Get authenticated user's profile
// @Description Retrieve the profile details of the currently authenticated user.
// @Tags Account
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
// @Tags Account
// @Security BearerAuth
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} models.User "User profile retrieved successfully"
// @Failure 400 {object} ErrorResponse "Invalid user ID"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 403 {object} ErrorResponse "Forbidden (insufficient permissions)"
// @Failure 404 {object} ErrorResponse "User not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /users/{id} [get]
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
// @Tags Account
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
// @Router /users/{id} [put]
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

// UpdateUserPasswordRequest represents the request body for updating a user's password.
type UpdateUserPasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required" example:"OldSecure_P@ss1"`
	NewPassword string `json:"new_password" binding:"required,min=8,alphanumunderscore" example:"NewSecure_P@ss2"` // Custom validation for password
}

// UpdateUserPassword handles updating a user's password.
// @Summary Update user password
// @Description Update an authenticated user's password.
// @Tags Account
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "User ID to update password for"
// @Param password_update body UpdateUserPasswordRequest true "Old and new password"
// @Success 200 {object} map[string]string "Password updated successfully"
// @Failure 400 {object} ErrorResponse "Invalid request payload or validation error"
// @Failure 401 {object} ErrorResponse "Unauthorized (invalid old password)"
// @Failure 403 {object} ErrorResponse "Forbidden (cannot update other users' passwords)"
// @Failure 404 {object} ErrorResponse "User not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /users/{id}/password [put]
func (h *UserController) UpdateUserPassword(c *gin.Context) {
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

	// Authorization: User can only update their own password
	if claims.UserID != uint(id) {
		c.JSON(http.StatusForbidden, ErrorResponse{Message: "Forbidden: You can only update your own password"})
		return
	}

	var req UpdateUserPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid request payload: " + err.Error()})
		return
	}

	// First, authenticate the old password
	user, err := h.userService.GetUserByID(uint(id))
	if err != nil {
		if err.Error() == fmt.Sprintf("user with ID %d not found", id) {
			c.JSON(http.StatusNotFound, ErrorResponse{Message: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Failed to retrieve user for password update: " + err.Error()})
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword))
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Message: "Invalid old password"})
		return
	}

	// Then, update with the new password
	if err := h.userService.UpdatePassword(uint(id), req.NewPassword); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Failed to update password: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password updated successfully"})
}

// DeleteUser handles deleting a user.
// @Summary Delete a user
// @Description Delete a user account by ID. Requires ADMIN or Sama Crew role, or user deleting self.
// @Tags Account
// @Security BearerAuth
// @Produce json
// @Param id path int true "User ID to delete"
// @Success 204 "User deleted successfully"
// @Failure 400 {object} ErrorResponse "Invalid user ID"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 403 {object} ErrorResponse "Forbidden (insufficient permissions)"
// @Failure 404 {object} ErrorResponse "User not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /users/{id} [delete]
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

// GetUsersBySchoolID handles retrieving users by school ID.
// @Summary Get users by school ID
// @Description Retrieve a list of users belonging to a specific school. Requires ADMIN or Sama Crew role.
// @Tags Account
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
// @Router /schools/{school_id}/users [get]
func (h *UserController) GetUsersBySchoolID(c *gin.Context) {
	claims, ok := middlewares.GetUserClaimsFromContext(c)
	if !ok {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "User claims not found in context"})
		return
	}

	// Authorization: Only ADMINs (for their school) or SAMA_CREW can access this
	if claims.Role != "ADMIN" && claims.Role != "SAMA_CREW" {
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

	users, err := h.userService.GetUsersBySchoolID(uint(schoolID), limit, offset)
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

// UpdateClassroomRequest represents the request body for updating a classroom.
type UpdateClassroomRequest struct {
	OldClassroom string `json:"old_classroom" binding:"required" example:"A101"`
	NewClassroom string `json:"new_classroom" binding:"required" example:"A102"`
}

// UpdateClassroomForSchool handles updating classrooms for all students in a school.
// @Summary Update classroom for a whole school
// @Description Update the classroom for all students in a given school from old classroom to new classroom. Requires ADMIN or Sama Crew role.
// @Tags Account
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param school_id path int true "School ID"
// @Param classroom_update body UpdateClassroomRequest true "Old and new classroom names"
// @Success 200 {object} map[string]int64 "Number of users updated"
// @Failure 400 {object} ErrorResponse "Invalid request payload or school ID"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 403 {object} ErrorResponse "Forbidden (insufficient permissions or not authorized for this school)"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /schools/{school_id}/classrooms [put]
func (h *UserController) UpdateClassroomForSchool(c *gin.Context) {
	claims, ok := middlewares.GetUserClaimsFromContext(c)
	if !ok {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "User claims not found in context"})
		return
	}

	// Authorization: Only ADMINs (for their school) or SAMA_CREW can access this
	if claims.Role != "ADMIN" && claims.Role != "SAMA_CREW" {
		c.JSON(http.StatusForbidden, ErrorResponse{Message: "Forbidden: Insufficient permissions"})
		return
	}

	schoolID, err := strconv.ParseUint(c.Param("school_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid school ID"})
		return
	}

	// If ADMIN, ensure they are updating classrooms in their own school
	if claims.Role == "ADMIN" && claims.SchoolID != uint(schoolID) {
		c.JSON(http.StatusForbidden, ErrorResponse{Message: "Forbidden: ADMIN can only update classrooms in their own school"})
		return
	}

	var req UpdateClassroomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid request payload: " + err.Error()})
		return
	}

	rowsAffected, err := h.userService.UpdateClassroomForSchool(uint(schoolID), req.OldClassroom, req.NewClassroom)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Failed to update classrooms: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("%d users' classrooms updated successfully", rowsAffected)})
}

// UpdateStudentClassroomRequest represents the request body for updating a single student's classroom.
type UpdateStudentClassroomRequest struct {
	NewClassroom string `json:"new_classroom" binding:"required" example:"C303"`
}

// UpdateClassroomForStudent handles updating a specific student's classroom.
// @Summary Update a specific student's classroom
// @Description Update the classroom for a specific student. Requires ADMIN or TCH role.
// @Tags Account
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param student_id path int true "Student User ID"
// @Param classroom_update body UpdateStudentClassroomRequest true "New classroom name"
// @Success 200 {object} map[string]string "Student classroom updated successfully"
// @Failure 400 {object} ErrorResponse "Invalid request payload or student ID"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 403 {object} ErrorResponse "Forbidden (insufficient permissions or not authorized for this student)"
// @Failure 404 {object} ErrorResponse "Student not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /users/{student_id}/classroom [put]
func (h *UserController) UpdateClassroomForStudent(c *gin.Context) {
	claims, ok := middlewares.GetUserClaimsFromContext(c)
	if !ok {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "User claims not found in context"})
		return
	}

	// Authorization: Only ADMIN or TCH can update a student's classroom
	if claims.Role != "ADMIN" && claims.Role != "TCH" && claims.Role != "SAMA_CREW" {
		c.JSON(http.StatusForbidden, ErrorResponse{Message: "Forbidden: Insufficient permissions"})
		return
	}

	studentID, err := strconv.ParseUint(c.Param("student_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid student ID"})
		return
	}

	var req UpdateStudentClassroomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid request payload: " + err.Error()})
		return
	}

	// Additional authorization: TCH can only update students in their school
	// ADMIN/SAMA_CREW can update any.
	if claims.Role == "TCH" {
		student, err := h.userService.GetUserByID(uint(studentID))
		if err != nil {
			if err.Error() == fmt.Sprintf("user with ID %d not found", studentID) {
				c.JSON(http.StatusNotFound, ErrorResponse{Message: err.Error()})
				return
			}
			c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Failed to retrieve student for authorization: " + err.Error()})
			return
		}
		if student.SchoolID != claims.SchoolID {
			c.JSON(http.StatusForbidden, ErrorResponse{Message: "Forbidden: Teacher can only update students within their own school"})
			return
		}
	}

	if err := h.userService.UpdateClassroomForStudent(uint(studentID), req.NewClassroom); err != nil {
		if err.Error() == fmt.Sprintf("user with ID %d not found", studentID) {
			c.JSON(http.StatusNotFound, err)
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Failed to update student classroom: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Student classroom updated successfully"})
}

// CheckStudentEmailRequest represents the request body for checking student email.
type CheckStudentEmailRequest struct {
	Email    string `json:"email" binding:"required,email" example:"student@example.com"`
	SchoolID uint   `json:"school_id" binding:"required,gt=0" example:"1"`
}

// CheckStudentEmailForRegistration handles checking if a student email is pre-registered for a school.
// @Summary Check if student email is eligible for registration
// @Description Checks if a student's email is known to the school (i.e., whitelisted) for registration. Requires ADMIN or Sama Crew role.
// @Tags Account
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param check_email body CheckStudentEmailRequest true "Email and School ID to check"
// @Success 200 {object} map[string]bool "Email eligibility status"
// @Failure 400 {object} ErrorResponse "Invalid request payload"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 403 {object} ErrorResponse "Forbidden (insufficient permissions or not authorized for this school)"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /check-student-email [post]
func (h *UserController) CheckStudentEmailForRegistration(c *gin.Context) {
	claims, ok := middlewares.GetUserClaimsFromContext(c)
	if !ok {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "User claims not found in context"})
		return
	}

	// Authorization: Only ADMINs (for their school) or SAMA_CREW can check this
	if claims.Role != "ADMIN" && claims.Role != "SAMA_CREW" {
		c.JSON(http.StatusForbidden, ErrorResponse{Message: "Forbidden: Insufficient permissions"})
		return
	}

	var req CheckStudentEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid request payload: " + err.Error()})
		return
	}

	// If ADMIN, ensure they are checking emails for their own school
	if claims.Role == "ADMIN" && claims.SchoolID != req.SchoolID {
		c.JSON(http.StatusForbidden, ErrorResponse{Message: "Forbidden: ADMIN can only check emails for their own school"})
		return
	}

	exists, err := h.userService.CheckStudentEmailForRegistration(req.Email, req.SchoolID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Failed to check email eligibility: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"is_eligible": !exists}) // Return true if email is NOT already registered in that school
}
