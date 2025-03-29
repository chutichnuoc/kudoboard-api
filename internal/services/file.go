package services

import (
	"fmt"
	"kudoboard-api/internal/config"
	"kudoboard-api/internal/dto/responses"
	"kudoboard-api/internal/services/storage"
	"kudoboard-api/internal/utils"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"
)

// File categories for organization
const (
	CategoryImage   = "image"
	CategoryGif     = "gif"
	CategoryVideo   = "video"
	CategoryTheme   = "theme"
	CategoryIcon    = "icon"
	CategoryAvatar  = "avatar"
	CategoryDefault = "general"
)

// FileService handles file uploads independently of posts or themes
type FileService struct {
	storage storage.StorageService
	cfg     *config.Config
}

// NewFileService creates a new FileService
func NewFileService(storage storage.StorageService, cfg *config.Config) *FileService {
	return &FileService{
		storage: storage,
		cfg:     cfg,
	}
}

// UploadFile handles file uploads and returns file information
func (s *FileService) UploadFile(file *multipart.FileHeader, userID uint, category string) (*responses.FileInfo, error) {
	// Validate file size (max 10MB)
	if file.Size > 10*1024*1024 {
		return nil, utils.NewBadRequestError("File size exceeds 10MB limit")
	}

	// Validate and set default category if needed
	if category == "" {
		category = CategoryDefault
	} else if !isValidCategory(category) {
		return nil, utils.NewBadRequestError("Invalid category. Allowed categories: image, gif, video, theme, icon, avatar, general")
	}

	fileExt := strings.ToLower(filepath.Ext(file.Filename))
	var fileType string

	switch fileExt {
	case ".jpg", ".jpeg", ".png", ".webp":
		fileType = "image"
	case ".gif":
		fileType = "gif"
	case ".mp4", ".webm", ".ogg":
		fileType = "video"
	default:
		return nil, utils.NewBadRequestError("Unsupported file type. Allowed types: jpg, jpeg, png, webp, gif, mp4, webm, ogg")
	}

	// Generate a unique directory path based on category and user
	var dirPath string
	if category == CategoryDefault || category == CategoryTheme || category == CategoryIcon {
		// Shared resources don't include user ID in the path
		dirPath = fmt.Sprintf("%s", category)
	} else if userID > 0 {
		// User-specific resources include user ID
		dirPath = fmt.Sprintf("%s/user_%d", category, userID)
	} else {
		// Anonymous uploads
		dirPath = fmt.Sprintf("%s/anonymous", category)
	}

	// Open the file to pass to storage service
	src, err := file.Open()
	if err != nil {
		return nil, utils.NewInternalError("Failed to open uploaded file", err)
	}
	defer src.Close()

	// Upload file using storage service
	storageInfo, err := s.storage.Save(file, dirPath)
	if err != nil {
		return nil, utils.NewInternalError("Failed to upload file", err)
	}

	// Return file info
	return &responses.FileInfo{
		FileName:    storageInfo.Filename,
		FilePath:    storageInfo.URL, // Use the URL directly from storage
		FileType:    fileType,
		FileSize:    storageInfo.Size,
		ContentType: storageInfo.ContentType,
		UploadedAt:  time.Now(),
	}, nil
}

// DeleteFile deletes a file from storage
func (s *FileService) DeleteFile(filePath string) error {
	// Delete file using storage service - pass the URL directly
	// The storage service will handle extracting the actual path
	err := s.storage.Delete(filePath)
	if err != nil {
		return utils.NewInternalError("Failed to delete file", err)
	}

	return nil
}

// isValidCategory check if category is valid
func isValidCategory(category string) bool {
	validCategories := map[string]bool{
		CategoryImage:   true,
		CategoryGif:     true,
		CategoryVideo:   true,
		CategoryTheme:   true,
		CategoryIcon:    true,
		CategoryAvatar:  true,
		CategoryDefault: true,
	}

	return validCategories[category]
}
