package handlers

import (
	"github.com/gin-gonic/gin"
	"kudoboard-api/internal/config"
	"kudoboard-api/internal/dto/requests"
	"kudoboard-api/internal/dto/responses"
	"kudoboard-api/internal/services"
	"net/http"
)

// FileHandler handles file-related requests
type FileHandler struct {
	fileService *services.FileService
	cfg         *config.Config
}

// NewFileHandler creates a new FileHandler
func NewFileHandler(fileService *services.FileService, cfg *config.Config) *FileHandler {
	return &FileHandler{
		fileService: fileService,
		cfg:         cfg,
	}
}

// UploadFile handles file uploads
func (h *FileHandler) UploadFile(c *gin.Context) {
	// Get user ID from context if authenticated
	userID := uint(0)
	user, exists := c.Get("user")
	if exists && user != nil {
		userID = c.GetUint("userID")
	}

	// Get category from form
	category := c.PostForm("category")
	if category == "" {
		category = "general" // Default category
	}

	// Get file from form
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, responses.ErrorResponse("INVALID_FILE", "Failed to read file"))
		return
	}

	// Upload file using service
	fileInfo, err := h.fileService.UploadFile(file, userID, category)
	if err != nil {
		c.JSON(http.StatusInternalServerError, responses.ErrorResponse("UPLOAD_ERROR", err.Error()))
		return
	}

	// Return file information
	c.JSON(http.StatusCreated, responses.SuccessResponse(fileInfo))
}

// DeleteFile handles file deletion
func (h *FileHandler) DeleteFile(c *gin.Context) {
	// Get user ID from context
	userID := c.GetUint("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, responses.ErrorResponse("UNAUTHORIZED", "User not authenticated"))
		return
	}

	// Get file path from request
	var req requests.DeleteFileRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, responses.ErrorResponse("VALIDATION_ERROR", err.Error()))
		return
	}

	// Delete file using service
	err := h.fileService.DeleteFile(req.FilePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, responses.ErrorResponse("DELETE_ERROR", err.Error()))
		return
	}

	c.JSON(http.StatusOK, responses.SuccessResponse(gin.H{"message": "File deleted successfully"}))
}
