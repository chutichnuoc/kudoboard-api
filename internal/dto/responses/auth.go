package responses

import (
	"kudoboard-api/internal/models"
	"time"
)

// AuthResponse represents the response for authentication requests
type AuthResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

// UserResponse represents user data in API responses
type UserResponse struct {
	ID             uint      `json:"id"`
	Name           string    `json:"name"`
	Email          string    `json:"email"`
	ProfilePicture string    `json:"profile_picture"`
	IsVerified     bool      `json:"is_verified"`
	AuthProvider   string    `json:"auth_provider"`
	CreatedAt      time.Time `json:"created_at"`
}

// FromUser converts a user model to a user response
func (ur *UserResponse) FromUser(user *models.User) {
	ur.ID = user.ID
	ur.Name = user.Name
	ur.Email = user.Email
	ur.ProfilePicture = user.ProfilePicture
	ur.IsVerified = user.IsVerified
	ur.AuthProvider = user.AuthProvider
	ur.CreatedAt = user.CreatedAt
}

// NewUserResponse creates a new user response from a user model
func NewUserResponse(user *models.User) UserResponse {
	var response UserResponse
	response.FromUser(user)
	return response
}
