package services

import (
	"fmt"
	"mime/multipart"
	"path/filepath"
	"regexp"
	"strings"

	"gorm.io/gorm"

	"kudoboard-api/internal/config"
	"kudoboard-api/internal/models"
	"kudoboard-api/internal/services/storage"
	"kudoboard-api/internal/utils"
)

// MediaService handles media-related business logic
type MediaService struct {
	db           *gorm.DB
	storage      storage.StorageService
	cfg          *config.Config
	boardService *BoardService
}

// NewMediaService creates a new MediaService
func NewMediaService(db *gorm.DB, storage storage.StorageService, cfg *config.Config, boardService *BoardService) *MediaService {
	return &MediaService{
		db:           db,
		storage:      storage,
		cfg:          cfg,
		boardService: boardService,
	}
}

// UploadMedia handles file uploads for posts
func (s *MediaService) UploadMedia(file *multipart.FileHeader, postID, userID uint) (*models.Media, error) {
	// Find post
	var post models.Post
	if result := s.db.First(&post, postID); result.Error != nil {
		return nil, utils.NewNotFoundError("Post not found")
	}

	// Check if user has permission to add media to this post
	if post.AuthorID == nil || *post.AuthorID != userID {
		// Check if user is board creator or admin
		var board models.Board
		s.db.First(&board, post.BoardID)
		if board.CreatorID != userID {
			// Check if user is a board admin
			var contributor models.BoardContributor
			result := s.db.Where("board_id = ? AND user_id = ? AND role = ?",
				post.BoardID, userID, models.RoleAdmin).First(&contributor)
			if result.Error != nil {
				return nil, utils.NewForbiddenError("You don't have permission to add media to this post")
			}
		}
	}

	// Validate file size (max 10MB)
	if file.Size > 10*1024*1024 {
		return nil, utils.NewBadRequestError("File size exceeds 10MB limit")
	}

	// Determine media type based on file extension
	fileExt := strings.ToLower(filepath.Ext(file.Filename))
	var mediaType models.MediaType

	switch fileExt {
	case ".jpg", ".jpeg", ".png", ".webp":
		mediaType = models.MediaTypeImage
	case ".gif":
		mediaType = models.MediaTypeGif
	case ".mp4", ".webm", ".ogg":
		mediaType = models.MediaTypeVideo
	default:
		return nil, utils.NewBadRequestError("Unsupported file type. Allowed types: jpg, jpeg, png, webp, gif, mp4, webm, ogg")
	}

	// Upload file to storage
	fileInfo, err := s.storage.Save(file, fmt.Sprintf("posts/%d", post.ID))
	if err != nil {
		return nil, utils.NewInternalError("Failed to upload file", err)
	}

	// Create media record
	media := models.Media{
		PostID:       postID,
		Type:         mediaType,
		SourceType:   models.SourceTypeUpload,
		SourceURL:    fileInfo.URL,
		ThumbnailURL: fileInfo.URL, // For simplicity, use same URL for thumbnail
	}

	// Save media to database
	if result := s.db.Create(&media); result.Error != nil {
		// If database save fails, try to delete the uploaded file
		_ = s.storage.Delete(fileInfo.URL)
		return nil, utils.NewInternalError("Failed to save media", result.Error)
	}

	return &media, nil
}

// AddYoutubeVideo adds a YouTube video to a post
func (s *MediaService) AddYoutubeVideo(postID, userID uint, youtubeURL string) (*models.Media, error) {
	// Find post
	var post models.Post
	if result := s.db.First(&post, postID); result.Error != nil {
		return nil, utils.NewNotFoundError("Post not found")
	}

	// Check if user has permission to add media to this post
	if post.AuthorID == nil || *post.AuthorID != userID {
		// Check if user is board creator or admin
		var board models.Board
		s.db.First(&board, post.BoardID)
		if board.CreatorID != userID {
			// Check if user is a board admin
			var contributor models.BoardContributor
			result := s.db.Where("board_id = ? AND user_id = ? AND role = ?",
				post.BoardID, userID, models.RoleAdmin).First(&contributor)
			if result.Error != nil {
				return nil, utils.NewForbiddenError("You don't have permission to add media to this post")
			}
		}
	}

	// Extract YouTube video ID from URL
	videoID, err := extractYouTubeID(youtubeURL)
	if err != nil {
		return nil, utils.NewBadRequestError(err.Error())
	}

	// Create source URL and thumbnail URL
	sourceURL := fmt.Sprintf("https://www.youtube.com/embed/%s", videoID)
	thumbnailURL := fmt.Sprintf("https://img.youtube.com/vi/%s/hqdefault.jpg", videoID)

	// Create media record
	media := models.Media{
		PostID:       postID,
		Type:         models.MediaTypeYoutube,
		SourceType:   models.SourceTypeYoutube,
		SourceURL:    sourceURL,
		ExternalID:   videoID,
		ThumbnailURL: thumbnailURL,
	}

	// Save media to database
	if result := s.db.Create(&media); result.Error != nil {
		return nil, utils.NewInternalError("Failed to save media", result.Error)
	}

	return &media, nil
}

// DeleteMedia removes a media item
func (s *MediaService) DeleteMedia(mediaID, userID uint) error {
	// Find media
	var media models.Media
	if result := s.db.First(&media, mediaID); result.Error != nil {
		return utils.NewNotFoundError("Media not found")
	}

	// Find post
	var post models.Post
	if result := s.db.First(&post, media.PostID); result.Error != nil {
		return utils.NewNotFoundError("Post not found")
	}

	// Check if user has permission to delete media
	if post.AuthorID == nil || *post.AuthorID != userID {
		// Check if user is board creator or admin
		var board models.Board
		s.db.First(&board, post.BoardID)
		if board.CreatorID != userID {
			// Check if user is a board admin
			var contributor models.BoardContributor
			result := s.db.Where("board_id = ? AND user_id = ? AND role = ?",
				post.BoardID, userID, models.RoleAdmin).First(&contributor)
			if result.Error != nil {
				return utils.NewForbiddenError("You don't have permission to delete this media")
			}
		}
	}

	// Start a transaction
	tx := s.db.Begin()

	// Delete media from database
	if err := tx.Delete(&media).Error; err != nil {
		tx.Rollback()
		return utils.NewInternalError("Failed to delete media", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return utils.NewInternalError("Failed to delete media", err)
	}

	// Delete file from storage if it's an uploaded file
	if media.SourceType == models.SourceTypeUpload {
		if err := s.storage.Delete(media.SourceURL); err != nil {
			// Log the error but don't fail the request
			// In a real application, you might want to queue this for cleanup later
		}
	}

	return nil
}

// GetMediaByID gets a media item by ID
func (s *MediaService) GetMediaByID(mediaID uint) (*models.Media, error) {
	var media models.Media
	if result := s.db.First(&media, mediaID); result.Error != nil {
		return nil, utils.NewNotFoundError("Media not found")
	}
	return &media, nil
}

// Helper function to extract YouTube video ID from various URL formats
func extractYouTubeID(url string) (string, error) {
	// Match standard YouTube URL formats
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?:youtube\.com\/watch\?v=|youtu\.be\/|youtube\.com\/embed\/)([^&?/]+)`),
		regexp.MustCompile(`youtube\.com\/watch\?.*v=([^&]+)`),
		regexp.MustCompile(`youtube\.com\/shorts\/([^&?/]+)`),
	}

	for _, pattern := range patterns {
		matches := pattern.FindStringSubmatch(url)
		if len(matches) > 1 {
			return matches[1], nil
		}
	}

	return "", fmt.Errorf("invalid YouTube URL format")
}
