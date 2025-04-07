package services

import (
	"encoding/json"
	"fmt"
	"gorm.io/gorm"
	"kudoboard-api/internal/config"
	"kudoboard-api/internal/models"
	"kudoboard-api/internal/utils"
	"net/http"
)

// AuthService handles authentication logic
type AuthService struct {
	db         *gorm.DB
	cfg        *config.Config
	httpClient *http.Client
}

// NewAuthService creates a new AuthService
func NewAuthService(db *gorm.DB, cfg *config.Config) *AuthService {
	return &AuthService{
		db:  db,
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: cfg.HTTPClientTimeout,
		},
	}
}

// RegisterUser registers a new user
func (s *AuthService) RegisterUser(name, email, password string) (*models.User, string, error) {
	// Check if user already exists
	var existingUser models.User
	if result := s.db.Where("email = ?", email).First(&existingUser); result.Error == nil {
		return nil, "", utils.NewBadRequestError("User with this email already exists").
			WithField("email", email)
	}

	// Create new user
	user := models.User{
		Name:     name,
		Email:    email,
		Password: password,
	}

	// Save user to database
	if result := s.db.Create(&user); result.Error != nil {
		return nil, "", utils.NewInternalError("Account creation failed", result.Error).
			WithField("email", email).
			WithField("name", name)
	}

	// Generate token
	token, err := utils.GenerateToken(user.ID, s.cfg.JWTSecret, s.cfg.JWTExpiresIn)
	if err != nil {
		return nil, "", utils.NewInternalError("Failed to generate token", err).
			WithField("user_id", user.ID)
	}

	return &user, token, nil
}

// LoginUser authenticates a user
func (s *AuthService) LoginUser(email, password string) (*models.User, string, error) {
	// Find user by email
	var user models.User
	if result := s.db.Where("email = ?", email).First(&user); result.Error != nil {
		return nil, "", utils.NewUnauthorizedError("Invalid email or password").
			WithField("email", email).
			WithField("error_type", "user_not_found")
	}

	// Check password
	if err := user.CheckPassword(password); err != nil {
		return nil, "", utils.NewUnauthorizedError("Invalid email or password").
			WithField("email", email).
			WithField("user_id", user.ID).
			WithField("error_type", "invalid_password")
	}

	// Generate token
	token, err := utils.GenerateToken(user.ID, s.cfg.JWTSecret, s.cfg.JWTExpiresIn)
	if err != nil {
		return nil, "", utils.NewInternalError("Authentication failed", err).
			WithField("user_id", user.ID)
	}

	return &user, token, nil
}

// GoogleLogin handles Google OAuth login
func (s *AuthService) GoogleLogin(accessToken string) (*models.User, string, error) {
	// Verify the token by calling Google's API
	resp, err := s.httpClient.Get("https://www.googleapis.com/oauth2/v3/tokeninfo?id_token=" + accessToken)
	if err != nil {
		return nil, "", utils.NewInternalError("Failed to verify Google token", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", utils.NewUnauthorizedError("Invalid Google token").
			WithField("status_code", resp.StatusCode)
	}

	// Parse the response
	var tokenInfo struct {
		Sub           string `json:"sub"`
		Email         string `json:"email"`
		EmailVerified string `json:"email_verified"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenInfo); err != nil {
		return nil, "", utils.NewInternalError("Failed to parse Google token info", err)
	}

	// Validate email verification
	if tokenInfo.EmailVerified != "true" {
		return nil, "", utils.NewUnauthorizedError("Email not verified with Google").
			WithField("email", tokenInfo.Email)
	}

	// Find user by Google ID or email
	var user models.User
	result := s.db.Where("google_id = ?", tokenInfo.Sub).Or("email = ?", tokenInfo.Email).First(&user)

	if result.Error != nil {
		// User doesn't exist, create new user
		user = models.User{
			Name:           tokenInfo.Name,
			Email:          tokenInfo.Email,
			Password:       "", // No password for OAuth users
			GoogleID:       &tokenInfo.Sub,
			ProfilePicture: tokenInfo.Picture,
			AuthProvider:   "google",
			IsVerified:     true,
		}

		if result := s.db.Create(&user); result.Error != nil {
			return nil, "", utils.NewInternalError("Account creation failed", result.Error).
				WithField("email", tokenInfo.Email).
				WithField("google_id", tokenInfo.Sub)
		}
	} else {
		// User exists, update Google ID and profile if needed
		updates := false

		if user.GoogleID == nil || *user.GoogleID != tokenInfo.Sub {
			user.GoogleID = &tokenInfo.Sub
			user.AuthProvider = "google"
			updates = true
		}

		if tokenInfo.Picture != "" && user.ProfilePicture != tokenInfo.Picture {
			user.ProfilePicture = tokenInfo.Picture
			updates = true
		}

		if updates {
			if result := s.db.Save(&user); result.Error != nil {
				return nil, "", utils.NewInternalError("Failed to update user", result.Error).
					WithField("user_id", user.ID).
					WithField("google_id", tokenInfo.Sub)
			}
		}
	}

	// Generate JWT token
	token, err := utils.GenerateToken(user.ID, s.cfg.JWTSecret, s.cfg.JWTExpiresIn)
	if err != nil {
		return nil, "", utils.NewInternalError("Failed to generate token", err).
			WithField("user_id", user.ID)
	}

	return &user, token, nil
}

// FacebookLogin handles Facebook OAuth login
func (s *AuthService) FacebookLogin(accessToken string) (*models.User, string, error) {
	// Verify the token by calling Facebook's API to get user info
	// We need to include fields=id,name,email to get these fields
	fbURL := fmt.Sprintf("https://graph.facebook.com/me?fields=id,name,email,picture&access_token=%s", accessToken)
	resp, err := s.httpClient.Get(fbURL)
	if err != nil {
		return nil, "", utils.NewInternalError("Failed to verify Facebook token", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", utils.NewUnauthorizedError("Invalid Facebook token").
			WithField("status_code", resp.StatusCode)
	}

	// Parse the response
	var fbUserInfo struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		Email   string `json:"email"`
		Picture struct {
			Data struct {
				URL string `json:"url"`
			} `json:"data"`
		} `json:"picture"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&fbUserInfo); err != nil {
		return nil, "", utils.NewInternalError("Failed to parse Facebook user info", err)
	}

	// Ensure we got an email (Facebook might not return it if user hasn't verified it)
	if fbUserInfo.Email == "" {
		return nil, "", utils.NewUnauthorizedError("Email not provided by Facebook. Please ensure your email is verified with Facebook").
			WithField("facebook_id", fbUserInfo.ID)
	}

	// Find user by Facebook ID or email
	var user models.User
	result := s.db.Where("facebook_id = ?", fbUserInfo.ID).Or("email = ?", fbUserInfo.Email).First(&user)

	if result.Error != nil {
		// User doesn't exist, create new user
		facebookID := fbUserInfo.ID // Create a variable to store the ID
		user = models.User{
			Name:           fbUserInfo.Name,
			Email:          fbUserInfo.Email,
			Password:       "", // No password for OAuth users
			FacebookID:     &facebookID,
			ProfilePicture: fbUserInfo.Picture.Data.URL,
			AuthProvider:   "facebook",
			IsVerified:     true,
		}

		if result := s.db.Create(&user); result.Error != nil {
			return nil, "", utils.NewInternalError("Account creation failed", result.Error).
				WithField("email", fbUserInfo.Email).
				WithField("facebook_id", fbUserInfo.ID)
		}
	} else {
		// User exists, update Facebook ID and profile if needed
		updates := false

		if user.FacebookID == nil || *user.FacebookID != fbUserInfo.ID {
			facebookID := fbUserInfo.ID
			user.FacebookID = &facebookID
			user.AuthProvider = "facebook"
			updates = true
		}

		pictureURL := fbUserInfo.Picture.Data.URL
		if pictureURL != "" && user.ProfilePicture != pictureURL {
			user.ProfilePicture = pictureURL
			updates = true
		}

		if updates {
			if result := s.db.Save(&user); result.Error != nil {
				return nil, "", utils.NewInternalError("Failed to update user", result.Error).
					WithField("user_id", user.ID).
					WithField("facebook_id", fbUserInfo.ID)
			}
		}
	}

	// Generate JWT token
	token, err := utils.GenerateToken(user.ID, s.cfg.JWTSecret, s.cfg.JWTExpiresIn)
	if err != nil {
		return nil, "", utils.NewInternalError("Failed to generate token", err).
			WithField("user_id", user.ID)
	}

	return &user, token, nil
}

// GetUserByID gets a user by ID
func (s *AuthService) GetUserByID(userID uint) (*models.User, error) {
	var user models.User
	if result := s.db.First(&user, userID); result.Error != nil {
		return nil, utils.NewNotFoundError("User not found").
			WithField("user_id", userID)
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
		return nil, utils.NewInternalError("Failed to update user", result.Error).
			WithField("user_id", userID)
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
		return 0, utils.NewUnauthorizedError("Invalid or expired token").
			WithField("error", err.Error())
	}
	return claims.UserID, nil
}
