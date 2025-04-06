package handlers

import (
	"github.com/gin-gonic/gin"
	"kudoboard-api/internal/config"
	"kudoboard-api/internal/dto/responses"
	"kudoboard-api/internal/services"
	"kudoboard-api/internal/utils"
	"net/http"
	"strconv"
)

// UnsplashHandler handles Unsplash API proxy requests
type UnsplashHandler struct {
	unsplashService *services.UnsplashService
	cfg             *config.Config
}

// NewUnsplashHandler creates a new UnsplashHandler
func NewUnsplashHandler(unsplashService *services.UnsplashService, cfg *config.Config) *UnsplashHandler {
	return &UnsplashHandler{
		unsplashService: unsplashService,
		cfg:             cfg,
	}
}

// Search handles photo search requests
func (h *UnsplashHandler) Search(c *gin.Context) {
	// Parse query parameters
	query := c.Query("query")
	if query == "" {
		_ = c.Error(utils.NewBadRequestError("Query parameter 'query' is required"))
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "10"))
	orderBy := c.DefaultQuery("order_by", "relevant")

	// Call Unsplash service
	result, err := h.unsplashService.Search(query, page, perPage, orderBy)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, responses.SuccessResponse(result))
}

// Random handles random photo requests
func (h *UnsplashHandler) Random(c *gin.Context) {
	// Parse query parameters
	count, _ := strconv.Atoi(c.DefaultQuery("count", "1"))
	if count < 1 {
		count = 1
	} else if count > 30 {
		count = 30 // Unsplash limit
	}

	query := c.Query("query")
	topics := c.Query("topics")
	username := c.Query("username")
	collections := c.Query("collections")
	featured := c.DefaultQuery("featured", "false") == "true"

	// Call Unsplash service
	result, err := h.unsplashService.Random(count, query, topics, username, collections, featured)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, responses.SuccessResponse(result))
}

// GetById handles retrieving a specific photo by ID
func (h *UnsplashHandler) GetById(c *gin.Context) {
	// Get photo ID from URL
	photoID := c.Param("photoId")
	if photoID == "" {
		_ = c.Error(utils.NewBadRequestError("Photo ID is required"))
		return
	}

	// Call Unsplash service
	result, err := h.unsplashService.GetById(photoID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, responses.SuccessResponse(result))
}
