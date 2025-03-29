package requests

// RegisterRequest represents the user registration request
type RegisterRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// LoginRequest represents the user login request
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// SocialLoginRequest represents a login request with a social provider
type SocialLoginRequest struct {
	AccessToken string `json:"access_token" binding:"required"`
}

// UpdateProfileRequest represents a request to update user profile
type UpdateProfileRequest struct {
	Name           string `json:"name"`
	ProfilePicture string `json:"profile_picture"`
}

// ForgotPasswordRequest represents a request for password reset
type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// ResetPasswordRequest represents a request to reset password with token
type ResetPasswordRequest struct {
	Token    string `json:"token" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
}
