package handlers

import (
	"github.com/gin-gonic/gin"
	"kudoboard-api/internal/config"
	"kudoboard-api/internal/dto/requests"
	"kudoboard-api/internal/dto/responses"
	"kudoboard-api/internal/services"
	"kudoboard-api/internal/utils"
	"net/http"
	"strconv"
)

// ThemeHandler handles theme-related requests
type ThemeHandler struct {
	themeService *services.ThemeService
	cfg          *config.Config
}

// NewThemeHandler creates a new ThemeHandler
func NewThemeHandler(themeService *services.ThemeService, cfg *config.Config) *ThemeHandler {
	return &ThemeHandler{
		themeService: themeService,
		cfg:          cfg,
	}
}

// ListThemes lists all available themes
func (h *ThemeHandler) ListThemes(c *gin.Context) {
	// Get themes using service
	themes, err := h.themeService.GetThemes()
	if err != nil {
		_ = c.Error(err)
		return
	}

	// Build response
	themeResponses := make([]responses.ThemeResponse, len(themes))
	for i, theme := range themes {
		themeResponses[i] = responses.NewThemeResponse(&theme)
	}

	c.JSON(http.StatusOK, responses.SuccessResponse(themeResponses))
}

// GetTheme gets a theme by ID
func (h *ThemeHandler) GetTheme(c *gin.Context) {
	// Get theme ID from URL
	themeID, err := strconv.ParseUint(c.Param("themeId"), 10, 32)
	if err != nil {
		_ = c.Error(utils.NewBadRequestError("Invalid theme id"))
		return
	}

	// Get theme using service
	theme, err := h.themeService.GetThemeByID(uint(themeID))
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, responses.SuccessResponse(responses.NewThemeResponse(theme)))
}

// CreateTheme creates a new theme
func (h *ThemeHandler) CreateTheme(c *gin.Context) {
	// Parse request
	var req requests.CreateThemeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(utils.NewValidationError(err.Error()))
		return
	}

	// Create theme using service
	theme, err := h.themeService.CreateTheme(req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, responses.SuccessResponse(responses.NewThemeResponse(theme)))
}

// UpdateTheme updates an existing theme
func (h *ThemeHandler) UpdateTheme(c *gin.Context) {
	// Get theme ID from URL
	themeID, err := strconv.ParseUint(c.Param("themeId"), 10, 32)
	if err != nil {
		_ = c.Error(utils.NewBadRequestError("Invalid theme ID"))
		return
	}

	// Parse request
	var req requests.UpdateThemeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(utils.NewValidationError(err.Error()))
		return
	}

	// Update theme using service
	theme, err := h.themeService.UpdateTheme(uint(themeID), req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, responses.SuccessResponse(responses.NewThemeResponse(theme)))
}

// DeleteTheme deletes a theme
func (h *ThemeHandler) DeleteTheme(c *gin.Context) {
	// Get theme ID from URL
	themeID, err := strconv.ParseUint(c.Param("themeId"), 10, 32)
	if err != nil {
		_ = c.Error(utils.NewBadRequestError("Invalid theme ID"))
		return
	}

	// Delete theme using service
	err = h.themeService.DeleteTheme(uint(themeID))
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, responses.SuccessResponse(gin.H{"message": "Theme deleted successfully"}))
}
