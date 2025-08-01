package controllers

import (
	"net/http"

	"sama/sama-backend-2025/src/models"
	"sama/sama-backend-2025/src/services"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	// Import bcrypt for password comparison
)

// AuthController manages HTTP requests for user authentication and account management.
type AuthController struct {
	authService *services.AuthService // Renamed from UserService for consistency with previous updates
	validate    *validator.Validate
}

// NewAuthController creates a new AuthController.
func NewAuthController(authService *services.AuthService, validate *validator.Validate) *AuthController {
	return &AuthController{
		authService: authService,
		validate:    validate,
	}
}

// RegisterRequest represents the request body for user registration.
type RegisterRequest struct {
	StudentID string  `json:"user_id,omitempty" example:"10101"` // UserID might be generated by system, or provided for specific roles
	Email     string  `json:"email" binding:"required,email" validate:"required,email" example:"user@example.com"`
	Password  string  `json:"password" binding:"required,min=8" validate:"required,min=8,alphanumunderscore" example:"Secure_P@ss1"` // Added alphanumunderscore to validate tag
	Firstname string  `json:"firstname" binding:"required" validate:"required" example:"John"`
	Lastname  string  `json:"lastname" binding:"required" validate:"required" example:"Doe"`
	Role      string  `json:"role" binding:"required,oneof=STD TCH ADMIN" validate:"required,oneof=STD TCH ADMIN" example:"STD"` // Validate against roles
	SchoolID  uint    `json:"school_id" binding:"required,gt=0" validate:"required,gt=0" example:"1"`
	Phone     string  `json:"phone,omitempty" example:"+1234567890"`
	Classroom *string `json:"classroom,omitempty" example:"A101"`
	Number    *uint   `json:"number,omitempty" example:"1"`
	Language  string  `json:"language,omitempty" example:"en"`
}

// LoginRequest represents the request body for user login.
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email" validate:"required,email" example:"user@example.com"`
	Password string `json:"password" binding:"required" validate:"required" example:"Secure_P@ss1"`
}

// LoginResponse represents the response body for successful login.
type LoginResponse struct {
	Token        string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken string `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

// RefreshTokenRequest represents the request body for generating new token.
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

// ValidateOtpRequest represents the request body for validating an OTP and resetting password.
type ValidateOtpRequest struct {
	Email       string `json:"email" binding:"required,email" validate:"required,email" example:"user@example.com"`
	Otp         string `json:"code" binding:"required,len=6" validate:"required,len=6" example:"123456"` // Assuming 6-digit OTP
	NewPassword string `json:"new_password" binding:"required,min=8" validate:"required,min=8,alphanumunderscore" example:"NewSecure_P@ss2"`
}

// RegisterUser handles user registration.
// @Summary Register a new user
// @Description Register a new user account (can be STD, TCH, ADMIN). UserID can be system-generated or provided.
// @Tags Auth
// @Accept json
// @Produce json
// @Param user body RegisterRequest true "User registration details"
// @Success 201 {object} models.User "User created successfully"
// @Failure 400 {object} ErrorResponse "Invalid request payload or validation error"
// @Failure 409 {object} ErrorResponse "User with this email already exists"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /register [post]
func (h *AuthController) RegisterUser(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid request payload: " + err.Error()})
		return
	}

	user := &models.User{
		StudentID: req.StudentID,
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

	if err := h.authService.RegisterUser(user); err != nil {
		if err.Error() == "user with this email already exists" {
			c.JSON(http.StatusConflict, ErrorResponse{Message: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Failed to register user: " + err.Error()})
		return
	}

	// Omit password from response for security
	user.Password = ""
	c.JSON(http.StatusCreated, user)
}

// Login handles user login and returns a JWT token.
// @Summary Log in a user
// @Description Authenticate user credentials and return a JWT token.
// @Tags Auth
// @Accept json
// @Produce json
// @Param credentials body LoginRequest true "User login credentials"
// @Success 200 {object} LoginResponse "Successful login with JWT token"
// @Failure 400 {object} ErrorResponse "Invalid request payload or validation error"
// @Failure 401 {object} ErrorResponse "Invalid credentials or account deactivated"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /login [post]
func (h *AuthController) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid request payload: " + err.Error()})
		return
	}

	token, refreshToken, err := h.authService.Login(req.Email, req.Password)
	if err != nil {
		if err.Error() == "invalid credentials" || err.Error() == "user account is deactivated" {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Message: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Failed to login: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, LoginResponse{Token: token, RefreshToken: refreshToken})
}

// RequestOtpRequest represents the request body for requesting an OTP.
type RequestOtpRequest struct {
	Email string `json:"email" binding:"required,email" validate:"required,email" example:"user@example.com"`
}

// RequestOtpResponse represents the response body for a successful OTP request.
type RequestOtpResponse struct {
	Message string `json:"message" example:"OTP sent to your email"`
}

// RequestOtp handles requesting an OTP for password reset.
// @Summary Request OTP for password reset
// @Description Sends a One-Time Password (OTP) to the user's registered email address to initiate a password reset.
// @Tags Auth
// @Accept json
// @Produce json
// @Param email_request body RequestOtpRequest true "User email to send OTP"
// @Success 200 {object} SuccessfulResponse "OTP sended"
// @Failure 400 {object} ErrorResponse "Invalid request payload or validation error"
// @Failure 404 {object} ErrorResponse "User with this email not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /password-reset/request-otp [post]
func (h *AuthController) RequestOtp(c *gin.Context) {
	var req RequestOtpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid request payload: " + err.Error()})
		return
	}

	// TODO: Implement actual OTP generation and sending logic in the service layer
	// Example:
	// err := h.userService.RequestPasswordResetOtp(req.Email)
	// if err != nil {
	//     if errors.Is(err, gorm.ErrRecordNotFound) {
	//         c.JSON(http.StatusNotFound, ErrorResponse{Message: "User with this email not found"})
	//         return
	//     }
	//     c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Failed to request OTP: " + err.Error()})
	//     return
	// }

	c.JSON(http.StatusOK, RequestOtpResponse{Message: "OTP sent to your email"})
}

// ValidateOtp handles validating an OTP and resetting the user's password.
// @Summary Validate OTP and reset password
// @Description Validates the provided OTP and email, then allows the user to set a new password. Returns a reset token or success message.
// @Tags Auth
// @Accept json
// @Produce json
// @Param otp_validation body ValidateOtpRequest true "OTP validation and new password details"
// @Success 200 {object} SuccessfulResponse "OTP validated and password reset successfully"
// @Failure 400 {object} ErrorResponse "Invalid request payload or validation error"
// @Failure 401 {object} ErrorResponse "Invalid OTP or email"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /password-reset/validate-otp [post]
func (h *AuthController) ValidateOtp(c *gin.Context) {
	var req ValidateOtpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid request payload: " + err.Error()})
		return
	}

	// TODO: Implement actual OTP validation and password reset logic in the service layer
	// Example:
	// token, err := h.userService.ValidateOtpAndResetPassword(req.Email, req.Otp, req.NewPassword)
	// if err != nil {
	//     if err.Error() == "invalid OTP or email" { // Custom error from service
	//         c.JSON(http.StatusUnauthorized, ErrorResponse{Message: err.Error()})
	//         return
	//     }
	//     c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Failed to validate OTP and reset password: " + err.Error()})
	//     return
	// }

	// For demonstration, returning a dummy token
	c.JSON(http.StatusOK, "Good to go")
}

// RefreshToken handles refreshing a JWT access token using a refresh token.
// @Summary Refresh access token
// @Description Exchanges a valid refresh token for a new access token and refresh token.
// @Tags Auth
// @Accept json
// @Produce json
// @Param refresh_token_request body RefreshTokenRequest true "Refresh token"
// @Success 200 {object} LoginResponse "New access and refresh tokens"
// @Failure 400 {object} ErrorResponse "Invalid request payload or validation error"
// @Failure 401 {object} ErrorResponse "Invalid or expired refresh token"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /refresh-token [post]
func (h *AuthController) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid request payload: " + err.Error()})
		return
	}

	// TODO: Call a service method to handle refresh token logic
	// This service method would:
	// 1. Validate the refresh token (e.g., check against a database of valid refresh tokens)
	// 2. If valid, generate a new access token and a new refresh token
	// 3. Invalidate the old refresh token (optional, but recommended for security)
	// Example:
	newAccessToken, newRefreshToken, err := h.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		if err.Error() == "invalid or expired refresh token" {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Message: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Failed to generate refresh token: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, LoginResponse{
		Token:        newAccessToken,
		RefreshToken: newRefreshToken,
	})
}
