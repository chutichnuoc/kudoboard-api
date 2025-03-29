package services

import (
	"gorm.io/gorm"
	"kudoboard-api/internal/config"
	"kudoboard-api/internal/dto/requests"
	"kudoboard-api/internal/models"
	"kudoboard-api/internal/services/storage"
	"kudoboard-api/internal/utils"
)

// ThemeService handles theme-related business logic
type ThemeService struct {
	db      *gorm.DB
	storage storage.StorageService
	cfg     *config.Config
}

// NewThemeService creates a new ThemeService
func NewThemeService(db *gorm.DB, storage storage.StorageService, cfg *config.Config) *ThemeService {
	return &ThemeService{
		db:      db,
		storage: storage,
		cfg:     cfg,
	}
}

// GetThemes gets all available themes
func (s *ThemeService) GetThemes() ([]models.Theme, error) {
	var themes []models.Theme
	if result := s.db.Find(&themes); result.Error != nil {
		return nil, utils.NewInternalError("Failed to fetch themes", result.Error)
	}
	return themes, nil
}

// GetThemeByID gets a theme by ID
func (s *ThemeService) GetThemeByID(themeID uint) (*models.Theme, error) {
	var theme models.Theme
	if result := s.db.First(&theme, themeID); result.Error != nil {
		return nil, utils.NewNotFoundError("Theme not found")
	}
	return &theme, nil
}

// CreateTheme creates a new theme
func (s *ThemeService) CreateTheme(input requests.CreateThemeRequest) (*models.Theme, error) {
	theme := models.Theme{
		Category:           input.Category,
		Name:               input.Name,
		IconUrl:            input.IconUrl,
		BackgroundImageURL: input.BackgroundImageURL,
	}

	if result := s.db.Create(&theme); result.Error != nil {
		return nil, utils.NewInternalError("Failed to create theme", result.Error)
	}

	return &theme, nil
}

// UpdateTheme updates an existing theme
func (s *ThemeService) UpdateTheme(themeID uint, input requests.UpdateThemeRequest) (*models.Theme, error) {
	// Find theme
	var theme models.Theme
	if result := s.db.First(&theme, themeID); result.Error != nil {
		return nil, utils.NewNotFoundError("Theme not found")
	}

	// Update fields if provided
	if input.Category != nil {
		theme.Category = *input.Category
	}
	if input.Name != nil {
		theme.Name = *input.Name
	}
	if input.IconUrl != nil {
		theme.IconUrl = *input.IconUrl
	}
	if input.BackgroundImageURL != nil {
		theme.BackgroundImageURL = *input.BackgroundImageURL
	}

	// Save changes
	if result := s.db.Save(&theme); result.Error != nil {
		return nil, utils.NewInternalError("Failed to update theme", result.Error)
	}

	return &theme, nil
}

// DeleteTheme deletes a theme
func (s *ThemeService) DeleteTheme(themeID uint) error {
	// Check if theme is in use by any boards
	var count int64
	if err := s.db.Model(&models.Board{}).Where("theme_id = ?", themeID).Count(&count).Error; err != nil {
		return utils.NewInternalError("Failed to check theme usage", err)
	}

	if count > 0 {
		return utils.NewBadRequestError("Cannot delete theme as it is being used by boards")
	}

	// Delete theme
	result := s.db.Delete(&models.Theme{}, themeID)
	if result.Error != nil {
		return utils.NewInternalError("Failed to delete theme", result.Error)
	}

	if result.RowsAffected == 0 {
		return utils.NewNotFoundError("Theme not found")
	}

	return nil
}
