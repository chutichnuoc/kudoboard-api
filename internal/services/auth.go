package services

import (
	"gorm.io/gorm"
	"kudoboard-api/internal/config"
	"kudoboard-api/internal/models"
	"kudoboard-api/internal/utils"
)

// AuthService handles authentication logic
type AuthService struct {
	db  *gorm.DB
	cfg *config.Config
}

// NewAuthService creates a new AuthService
func NewAuthService(db *gorm.DB, cfg *config.Config) *AuthService {
	return &AuthService{
		db:  db,
		cfg: cfg,
	}
}

// RegisterUser registers a new user
func (s *AuthService) RegisterUser(name, email, password string) (*models.User, string, error) {
	// Check if user already exists
	var existingUser models.User
	if result := s.db.Where("email = ?", email).First(&existingUser); result.Error == nil {
		return nil, "", utils.NewBadRequestError("User with this email already exists")
	}

	// Create new user
	user := models.User{
		Name:     name,
		Email:    email,
		Password: password,
	}

	// Save user to database
	if result := s.db.Create(&user); result.Error != nil {
		return nil, "", utils.NewInternalError("Failed to create user", result.Error)
	}

	// Generate token
	token, err := utils.GenerateToken(user.ID, s.cfg.JWTSecret, s.cfg.JWTExpiresIn)
	if err != nil {
		return nil, "", utils.NewInternalError("Failed to generate token", err)
	}

	return &user, token, nil
}

// LoginUser authenticates a user
func (s *AuthService) LoginUser(email, password string) (*models.User, string, error) {
	// Find user by email
	var user models.User
	if result := s.db.Where("email = ?", email).First(&user); result.Error != nil {
		return nil, "", utils.NewUnauthorizedError("Invalid email or password")
	}

	// Check password
	if err := user.CheckPassword(password); err != nil {
		return nil, "", utils.NewUnauthorizedError("Invalid email or password")
	}

	// Generate token
	token, err := utils.GenerateToken(user.ID, s.cfg.JWTSecret, s.cfg.JWTExpiresIn)
	if err != nil {
		return nil, "", utils.NewInternalError("Failed to generate token", err)
	}

	return &user, token, nil
}

// GoogleLogin handles Google OAuth login
func (s *AuthService) GoogleLogin(accessToken string) (*models.User, string, error) {
	// In a real implementation, you would:
	// 1. Verify the Google token by calling Google's API
	// 2. Get user info from Google
	// 3. Create or update the user in your database
	// 4. Generate a JWT token

	// For this example, we'll mock it
	// This should be replaced with actual Google API calls
	googleID := "google_123456789"
	email := "example@gmail.com"
	name := "Google User"

	// Find user by Google ID or email
	var user models.User
	result := s.db.Where("google_id = ?", googleID).Or("email = ?", email).First(&user)

	if result.Error != nil {
		// User doesn't exist, create new user
		user = models.User{
			Name:         name,
			Email:        email,
			GoogleID:     &googleID,
			AuthProvider: "google",
			IsVerified:   true,
		}

		if result := s.db.Create(&user); result.Error != nil {
			return nil, "", utils.NewInternalError("Failed to create user", result.Error)
		}
	} else {
		// User exists, update Google ID if needed
		if user.GoogleID == nil {
			user.GoogleID = &googleID
			user.AuthProvider = "google"
			if result := s.db.Save(&user); result.Error != nil {
				return nil, "", utils.NewInternalError("Failed to update user", result.Error)
			}
		}
	}

	// Generate token
	token, err := utils.GenerateToken(user.ID, s.cfg.JWTSecret, s.cfg.JWTExpiresIn)
	if err != nil {
		return nil, "", utils.NewInternalError("Failed to generate token", err)
	}

	return &user, token, nil
}

// FacebookLogin handles Facebook OAuth login
func (s *AuthService) FacebookLogin(accessToken string) (*models.User, string, error) {
	// Similar to GoogleLogin, in a real implementation you would:
	// 1. Verify the Facebook token
	// 2. Get user info from Facebook
	// 3. Create or update the user in your database
	// 4. Generate a JWT token

	// Mock implementation
	facebookID := "facebook_123456789"
	email := "example@facebook.com"
	name := "Facebook User"

	// Find user by Facebook ID or email
	var user models.User
	result := s.db.Where("facebook_id = ?", facebookID).Or("email = ?", email).First(&user)

	if result.Error != nil {
		// User doesn't exist, create new user
		user = models.User{
			Name:         name,
			Email:        email,
			FacebookID:   &facebookID,
			AuthProvider: "facebook",
			IsVerified:   true,
		}

		if result := s.db.Create(&user); result.Error != nil {
			return nil, "", utils.NewInternalError("Failed to create user", result.Error)
		}
	} else {
		// User exists, update Facebook ID if needed
		if user.FacebookID == nil {
			user.FacebookID = &facebookID
			user.AuthProvider = "facebook"
			if result := s.db.Save(&user); result.Error != nil {
				return nil, "", utils.NewInternalError("Failed to update user", result.Error)
			}
		}
	}

	// Generate token
	token, err := utils.GenerateToken(user.ID, s.cfg.JWTSecret, s.cfg.JWTExpiresIn)
	if err != nil {
		return nil, "", utils.NewInternalError("Failed to generate token", err)
	}

	return &user, token, nil
}

// GetUserByID gets a user by ID
func (s *AuthService) GetUserByID(userID uint) (*models.User, error) {
	var user models.User
	if result := s.db.First(&user, userID); result.Error != nil {
		return nil, utils.NewNotFoundError("User not found")
	}
	return &user, nil
}

// UpdateUser updates a user's profile
func (s *AuthService) UpdateUser(userID uint, name, profilePicture string) (*models.User, error) {
	var user models.User
	if result := s.db.First(&user, userID); result.Error != nil {
		return nil, utils.NewNotFoundError("User not found")
	}

	// Update fields if provided
	if name != "" {
		user.Name = name
	}
	if profilePicture != "" {
		user.ProfilePicture = profilePicture
	}

	// Save changes
	if result := s.db.Save(&user); result.Error != nil {
		return nil, utils.NewInternalError("Failed to update user", result.Error)
	}

	return &user, nil
}

// ForgotPassword initiates the password reset process
func (s *AuthService) ForgotPassword(email string) error {
	var user models.User
	if result := s.db.Where("email = ?", email).First(&user); result.Error != nil {
		// Don't reveal if the email exists for security reasons
		return nil
	}

	// In a real implementation, you would:
	// 1. Generate a reset token
	// 2. Store it in the database with an expiration time
	// 3. Send an email with a reset link

	return nil
}

// ResetPassword resets a user's password
func (s *AuthService) ResetPassword(token, newPassword string) error {
	// In a real implementation, you would:
	// 1. Verify the reset token
	// 2. Check if it's expired
	// 3. Find the associated user
	// 4. Update their password

	return nil
}

// VerifyToken verifies a JWT token and returns the user ID
func (s *AuthService) VerifyToken(tokenString string) (uint, error) {
	claims, err := utils.VerifyToken(tokenString, s.cfg.JWTSecret)
	if err != nil {
		return 0, utils.NewUnauthorizedError("Invalid or expired token")
	}
	return claims.UserID, nil
}
