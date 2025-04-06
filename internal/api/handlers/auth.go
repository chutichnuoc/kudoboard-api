package handlers

import (
	"github.com/gin-gonic/gin"
	"kudoboard-api/internal/config"
	"kudoboard-api/internal/dto/requests"
	"kudoboard-api/internal/dto/responses"
	"kudoboard-api/internal/services"
	"kudoboard-api/internal/utils"
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
		_ = c.Error(utils.NewValidationError(err.Error()))
		return
	}

	// Register user using auth service
	user, token, err := h.authService.RegisterUser(req.Name, req.Email, req.Password)
	if err != nil {
		_ = c.Error(err)
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
		_ = c.Error(utils.NewValidationError(err.Error()))
		return
	}

	// Login user using auth service
	user, token, err := h.authService.LoginUser(req.Email, req.Password)
	if err != nil {
		_ = c.Error(err)
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
		_ = c.Error(utils.NewUnauthorizedError("User not authenticated"))
		return
	}

	// Get user by ID
	user, err := h.authService.GetUserByID(userID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, responses.SuccessResponse(responses.NewUserResponse(user)))
}

// UpdateProfile updates the user's profile
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	// Get user ID from context
	userID := c.GetUint("userID")
	if userID == 0 {
		_ = c.Error(utils.NewUnauthorizedError("User not authenticated"))
		return
	}

	var req requests.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(utils.NewValidationError(err.Error()))
		return
	}

	// Update user profile
	user, err := h.authService.UpdateUser(userID, req.Name, req.ProfilePicture)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, responses.SuccessResponse(responses.NewUserResponse(user)))
}

// GoogleLogin handles Google OAuth login
func (h *AuthHandler) GoogleLogin(c *gin.Context) {
	var req requests.SocialLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(utils.NewValidationError(err.Error()))
		return
	}

	// Login with Google
	user, token, err := h.authService.GoogleLogin(req.AccessToken)
	if err != nil {
		_ = c.Error(err)
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
		_ = c.Error(utils.NewValidationError(err.Error()))
		return
	}

	// Login with Facebook
	user, token, err := h.authService.FacebookLogin(req.AccessToken)
	if err != nil {
		_ = c.Error(err)
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
		_ = c.Error(utils.NewValidationError(err.Error()))
	}

	// Initiate password reset
	err := h.authService.ForgotPassword(req.Email)
	if err != nil {
		_ = c.Error(err)
		return
	}

	// For security reasons, always return a positive response
	// even if the email doesn't exist
	c.JSON(http.StatusOK, responses.SuccessResponse(gin.H{
		"message": "If your email exists in our system, you will receive a password reset link",
	}))
}

// ResetPassword resets a user's password using a reset token
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req requests.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(utils.NewValidationError(err.Error()))
		return
	}

	// Reset password
	err := h.authService.ResetPassword(req.Token, req.Password)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, responses.SuccessResponse(gin.H{
		"message": "Password has been reset successfully",
	}))
}
