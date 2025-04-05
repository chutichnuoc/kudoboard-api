package handlers

import (
	"github.com/gin-gonic/gin"
	"kudoboard-api/internal/config"
	"kudoboard-api/internal/dto/responses"
	"kudoboard-api/internal/services"
	"net/http"
	"strconv"
)

// GiphyHandler handles Giphy API proxy requests
type GiphyHandler struct {
	giphyService *services.GiphyService
	cfg          *config.Config
}

// NewGiphyHandler creates a new GiphyHandler
func NewGiphyHandler(giphyService *services.GiphyService, cfg *config.Config) *GiphyHandler {
	return &GiphyHandler{
		giphyService: giphyService,
		cfg:          cfg,
	}
}

// Search handles GIF search requests
func (h *GiphyHandler) Search(c *gin.Context) {
	// Parse query parameters
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, responses.ErrorResponse("INVALID_QUERY", "Query parameter 'q' is required"))
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "25"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	rating := c.Query("rating")
	lang := c.Query("lang")

	// Call Giphy service
	result, err := h.giphyService.Search(query, limit, offset, rating, lang)
	if err != nil {
		c.JSON(http.StatusInternalServerError, responses.ErrorResponse("GIPHY_ERROR", err.Error()))
		return
	}

	c.JSON(http.StatusOK, responses.SuccessResponse(result))
}

// Trending handles trending GIFs requests
func (h *GiphyHandler) Trending(c *gin.Context) {
	// Parse query parameters
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "25"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	rating := c.Query("rating")

	// Call Giphy service
	result, err := h.giphyService.Trending(limit, offset, rating)
	if err != nil {
		c.JSON(http.StatusInternalServerError, responses.ErrorResponse("GIPHY_ERROR", err.Error()))
		return
	}

	c.JSON(http.StatusOK, responses.SuccessResponse(result))
}

// GetById handles retrieving a specific GIF by ID
func (h *GiphyHandler) GetById(c *gin.Context) {
	// Get GIF ID from URL
	gifId := c.Param("gifId")
	if gifId == "" {
		c.JSON(http.StatusBadRequest, responses.ErrorResponse("INVALID_ID", "GIF ID is required"))
		return
	}

	// Call Giphy service
	result, err := h.giphyService.GetById(gifId)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "GIF not found" {
			status = http.StatusNotFound
		}
		c.JSON(status, responses.ErrorResponse("GIPHY_ERROR", err.Error()))
		return
	}

	c.JSON(http.StatusOK, responses.SuccessResponse(result))
}

// Random handles random GIF requests
func (h *GiphyHandler) Random(c *gin.Context) {
	// Parse query parameters
	tag := c.Query("tag")
	rating := c.Query("rating")

	// Call Giphy service
	result, err := h.giphyService.Random(tag, rating)
	if err != nil {
		c.JSON(http.StatusInternalServerError, responses.ErrorResponse("GIPHY_ERROR", err.Error()))
		return
	}

	c.JSON(http.StatusOK, responses.SuccessResponse(result))
}
