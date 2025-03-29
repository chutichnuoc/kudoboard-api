package requests

// CreateThemeRequest represents the request to create a new theme
type CreateThemeRequest struct {
	Category           string `json:"category" binding:"required"`
	Name               string `json:"name" binding:"required"`
	IconUrl            string `json:"icon_url" binding:"required"`
	BackgroundImageURL string `json:"background_image_url" binding:"required"`
}

// UpdateThemeRequest represents the request to update a theme
type UpdateThemeRequest struct {
	Category           *string `json:"category"`
	Name               *string `json:"name"`
	IconUrl            *string `json:"icon_url"`
	BackgroundImageURL *string `json:"background_image_url"`
}
