package handlers

import (
	"github.com/gin-gonic/gin"
	"kudoboard-api/internal/config"
	"kudoboard-api/internal/dto/requests"
	"kudoboard-api/internal/dto/responses"
	"kudoboard-api/internal/services"
	"net/http"
)

// AuthHandler handles authentication-related requests
type AuthHandler struct {
	authService *services.AuthService
	cfg         *config.Config
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(authService *services.AuthService, cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		cfg:         cfg,
	}
}

// Register handles user registration
func (h *AuthHandler) Register(c *gin.Context) {
	var req requests.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, responses.ErrorResponse("VALIDATION_ERROR", err.Error()))
		return
	}

	// Register user using auth service
	user, token, err := h.authService.RegisterUser(req.Name, req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, responses.ErrorResponse("REGISTRATION_ERROR", err.Error()))
		return
	}

	// Create response
	c.JSON(http.StatusCreated, responses.SuccessResponse(responses.AuthResponse{
		Token: token,
		User:  responses.NewUserResponse(user),
	}))
}

// Login handles user login
func (h *AuthHandler) Login(c *gin.Context) {
	var req requests.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, responses.ErrorResponse("VALIDATION_ERROR", err.Error()))
		return
	}

	// Login user using auth service
	user, token, err := h.authService.LoginUser(req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, responses.ErrorResponse("AUTHENTICATION_ERROR", "Invalid email or password"))
		return
	}

	// Create response
	c.JSON(http.StatusOK, responses.SuccessResponse(responses.AuthResponse{
		Token: token,
		User:  responses.NewUserResponse(user),
	}))
}

// GetMe returns the currently authenticated user
func (h *AuthHandler) GetMe(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID := c.GetUint("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, responses.ErrorResponse("UNAUTHORIZED", "Not authenticated"))
		return
	}

	// Get user by ID
	user, err := h.authService.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, responses.ErrorResponse("USER_NOT_FOUND", "User not found"))
		return
	}

	c.JSON(http.StatusOK, responses.SuccessResponse(responses.NewUserResponse(user)))
}

// UpdateProfile updates the user's profile
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	// Get user ID from context
	userID := c.GetUint("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, responses.ErrorResponse("UNAUTHORIZED", "Not authenticated"))
		return
	}

	var req requests.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, responses.ErrorResponse("VALIDATION_ERROR", err.Error()))
		return
	}

	// Update user profile
	user, err := h.authService.UpdateUser(userID, req.Name, req.ProfilePicture)
	if err != nil {
		c.JSON(http.StatusInternalServerError, responses.ErrorResponse("UPDATE_ERROR", "Failed to update profile"))
		return
	}

	c.JSON(http.StatusOK, responses.SuccessResponse(responses.NewUserResponse(user)))
}

// GoogleLogin handles Google OAuth login
func (h *AuthHandler) GoogleLogin(c *gin.Context) {
	var req requests.SocialLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, responses.ErrorResponse("VALIDATION_ERROR", err.Error()))
		return
	}

	// Login with Google
	user, token, err := h.authService.GoogleLogin(req.AccessToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, responses.ErrorResponse("AUTHENTICATION_ERROR", err.Error()))
		return
	}

	c.JSON(http.StatusOK, responses.SuccessResponse(responses.AuthResponse{
		Token: token,
		User:  responses.NewUserResponse(user),
	}))
}

// FacebookLogin handles Facebook OAuth login
func (h *AuthHandler) FacebookLogin(c *gin.Context) {
	var req requests.SocialLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, responses.ErrorResponse("VALIDATION_ERROR", err.Error()))
		return
	}

	// Login with Facebook
	user, token, err := h.authService.FacebookLogin(req.AccessToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, responses.ErrorResponse("AUTHENTICATION_ERROR", err.Error()))
		return
	}

	c.JSON(http.StatusOK, responses.SuccessResponse(responses.AuthResponse{
		Token: token,
		User:  responses.NewUserResponse(user),
	}))
}

// ForgotPassword initiates the password reset process
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req requests.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, responses.ErrorResponse("VALIDATION_ERROR", err.Error()))
		return
	}

	// Initiate password reset
	err := h.authService.ForgotPassword(req.Email)
	if err != nil {
		// Don't reveal if the email exists for security reasons
		c.JSON(http.StatusInternalServerError, responses.ErrorResponse("RESET_ERROR", "Failed to process password reset"))
		return
	}

	c.JSON(http.StatusOK, responses.SuccessResponse(gin.H{
		"message": "If your email exists in our system, you will receive a password reset link",
	}))
}

// ResetPassword resets a user's password using a reset token
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req requests.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, responses.ErrorResponse("VALIDATION_ERROR", err.Error()))
		return
	}

	// Reset password
	err := h.authService.ResetPassword(req.Token, req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, responses.ErrorResponse("RESET_ERROR", "Invalid or expired reset token"))
		return
	}

	c.JSON(http.StatusOK, responses.SuccessResponse(gin.H{
		"message": "Password has been reset successfully",
	}))
}
