package services

import (
	"errors"
	"fmt"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"kudoboard-api/internal/config"
	"kudoboard-api/internal/dto/requests"
	"kudoboard-api/internal/log"
	"kudoboard-api/internal/models"
	"kudoboard-api/internal/services/storage"
	"kudoboard-api/internal/utils"
	"regexp"
)

// PostService handles post-related business logic
type PostService struct {
	db           *gorm.DB
	storage      storage.StorageService
	cfg          *config.Config
	boardService *BoardService
}

// NewPostService creates a new PostService
func NewPostService(db *gorm.DB, storage storage.StorageService, cfg *config.Config, boardService *BoardService) *PostService {
	return &PostService{
		db:           db,
		storage:      storage,
		cfg:          cfg,
		boardService: boardService,
	}
}

// CreatePost creates a new post
func (s *PostService) CreatePost(boardID, userID uint, input requests.CreatePostRequest) (*models.Post, error) {
	// Check if board exists
	var board models.Board
	if result := s.db.First(&board, boardID); result.Error != nil {
		return nil, utils.NewNotFoundError("Board not found").
			WithField("board_id", boardID)
	}

	// Check if board is locked
	if board.IsLocked {
		return nil, utils.NewForbiddenError("This board is locked and doesn't allow new posts")
	}

	// Determine if this is an anonymous post based on authentication
	isAnonymous := userID == 0

	// If anonymous, check if the board allows anonymous posts
	if isAnonymous && !board.AllowAnonymous {
		return nil, utils.NewForbiddenError("This board does not allow anonymous posts")
	}

	// For authenticated users, check access
	if !isAnonymous {
		// Check if user has access to the board
		canAccess, err := s.boardService.CanAccessBoard(boardID, userID)
		if err != nil {
			return nil, err
		}
		if !canAccess {
			return nil, utils.NewForbiddenError("You don't have access to this board")
		}
	}

	// If media type is YouTube, extract video id from media path and format it
	mediaPath := input.MediaPath
	if input.MediaType == "youtube" {
		var videoID string
		videoID, err := extractYouTubeID(input.MediaPath)
		if err != nil {
			return nil, utils.NewBadRequestError(err.Error())
		}
		mediaPath = fmt.Sprintf("https://www.youtube.com/embed/%s", videoID)
	}

	// Create post
	post := models.Post{
		BoardID:         boardID,
		Content:         input.Content,
		MediaPath:       mediaPath,
		MediaType:       input.MediaType,
		MediaSource:     input.MediaSource,
		BackgroundColor: input.BackgroundColor,
		TextColor:       input.TextColor,
		Position:        0, // Will be updated in the transaction
	}

	// Set author details based on authentication status
	if isAnonymous {
		post.AuthorName = input.AuthorName
	} else {
		// Get user for author name
		var user models.User
		if result := s.db.First(&user, userID); result.Error != nil {
			return nil, utils.NewInternalError("Failed to get user", result.Error)
		}
		post.AuthorID = &userID
		post.AuthorName = user.Name
	}

	// Save post and update position in a transaction
	err := utils.WithTransaction(s.db, func(tx *gorm.DB) error {
		// Save the post first to get an ID
		if result := tx.Create(&post).Error; result != nil {
			return utils.NewInternalError("Failed to create post", result)
		}

		// Now update the position using a direct SQL query with atomic increment
		// This ensures each post gets a unique position even with concurrent requests
		updateResult := tx.Exec(`
			UPDATE posts 
			SET position = (
				SELECT COALESCE(MAX(position), 0) + 1 
				FROM posts 
				WHERE board_id = ? AND id != ?
			)
			WHERE id = ?
		`, boardID, post.ID, post.ID)

		if updateResult.Error != nil {
			return utils.NewInternalError("Failed to update post position", updateResult.Error)
		}

		// If authenticated and not already a contributor, add as contributor
		if !isAnonymous {
			var contributor models.BoardContributor
			result := tx.Where("board_id = ? AND user_id = ?", boardID, userID).First(&contributor)
			if result.Error != nil {
				// User is not a contributor yet, create a contributor record
				newContributor := models.BoardContributor{
					BoardID: boardID,
					UserID:  userID,
					Role:    models.RoleContributor,
				}
				if err := tx.Create(&newContributor).Error; err != nil {
					log.Warn("Failed to create contributor record",
						zap.Uint("board_id", boardID),
						zap.Uint("user_id", userID),
						zap.Error(err))
					// Continue without failing the transaction
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Reload the post to get the updated position
	if err := s.db.First(&post, post.ID).Error; err != nil {
		return nil, utils.NewInternalError("Failed to reload post", err)
	}

	return &post, nil
}

// GetPostByID gets a post by ID
func (s *PostService) GetPostByID(postID uint) (*models.Post, error) {
	var post models.Post
	if result := s.db.First(&post, postID); result.Error != nil {
		return nil, utils.NewNotFoundError("Post not found").
			WithField("post_id", postID)
	}
	return &post, nil
}

// UpdatePost updates a post
func (s *PostService) UpdatePost(postID, userID uint, input requests.UpdatePostRequest) (*models.Post, error) {
	// Find post
	var post models.Post
	if result := s.db.First(&post, postID); result.Error != nil {
		return nil, utils.NewNotFoundError("Post not found").
			WithField("post_id", postID)
	}

	// Get the board to check if it's locked
	var board models.Board
	if result := s.db.First(&board, post.BoardID); result.Error != nil {
		return nil, utils.NewNotFoundError("Board not found").
			WithField("board_id", post.BoardID)
	}

	// Check if board is locked
	if board.IsLocked {
		return nil, utils.NewForbiddenError("This board is locked and doesn't allow modifications").
			WithField("board_Id", board.ID)
	}

	// Check if user has permission to update this post
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
				return nil, utils.NewForbiddenError("You don't have permission to update this post").
					WithField("post_id", postID).
					WithField("user_id", userID)
			}
		}
	}

	oldMediaPath := post.MediaPath
	oldMediaSource := post.MediaSource

	// Update fields if provided
	if input.Content != nil {
		post.Content = *input.Content
	}
	if input.AuthorName != nil {
		post.AuthorName = *input.AuthorName
	}
	// MediaType and MediaSource is not null, checked by binding
	if input.MediaPath != nil {
		if *input.MediaType == "youtube" {
			var videoID string
			videoID, err := extractYouTubeID(*input.MediaPath)
			if err != nil {
				return nil, utils.NewBadRequestError(err.Error()).
					WithField("media_path", *input.MediaPath)
			}
			post.MediaPath = fmt.Sprintf("https://www.youtube.com/embed/%s", videoID)
		} else {
			post.MediaPath = *input.MediaPath
		}
		post.MediaType = *input.MediaType
		post.MediaSource = *input.MediaSource
	}
	if input.BackgroundColor != nil {
		post.BackgroundColor = *input.BackgroundColor
	}
	if input.TextColor != nil {
		post.TextColor = *input.TextColor
	}

	// Save changes
	if result := s.db.Save(&post); result.Error != nil {
		return nil, utils.NewInternalError("Failed to update post", result.Error).
			WithField("post_id", post.ID)
	}

	if oldMediaPath != "" && oldMediaPath != post.MediaPath && oldMediaSource == "internal" {
		if err := s.storage.Delete(oldMediaPath); err != nil {
			log.Warn("Failed to delete old media",
				zap.Uint("post_id", postID),
				zap.String("file_path", oldMediaPath),
				zap.Error(err))
		}
	}

	return &post, nil
}

// DeletePost deletes a post
func (s *PostService) DeletePost(postID, userID uint) error {
	// Find post
	var post models.Post
	if result := s.db.First(&post, postID); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return utils.NewNotFoundError("Post not found").
				WithField("post_id", postID)
		}
		return utils.NewInternalError("Failed to query post", result.Error).
			WithField("post_id", postID)
	}

	// Get the board to check if it's locked
	var board models.Board
	if result := s.db.First(&board, post.BoardID); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return utils.NewNotFoundError("Board not found").
				WithField("board_id", post.BoardID)
		}
		return utils.NewInternalError("Failed to query board", result.Error).
			WithField("board_id", post.BoardID)
	}

	// Check if board is locked
	if board.IsLocked {
		return utils.NewForbiddenError("This board is locked and doesn't allow modifications").
			WithField("board_id", post.BoardID)
	}

	// Check if user has permission to delete this post
	if post.AuthorID == nil || *post.AuthorID != userID {
		// Check if user is board creator or admin
		if board.CreatorID != userID {
			// Check if user is a board admin
			var contributor models.BoardContributor
			result := s.db.Where("board_id = ? AND user_id = ? AND role = ?",
				post.BoardID, userID, models.RoleAdmin).First(&contributor)
			if result.Error != nil {
				return utils.NewForbiddenError("You don't have permission to delete this post").
					WithField("post_id", postID).
					WithField("user_id", userID).
					WithField("board_id", post.BoardID)
			}
		}
	}

	// Store media path for deletion after transaction
	mediaPath := post.MediaPath
	mediaSource := post.MediaSource

	// Delete the post and its likes in a transaction
	err := utils.WithTransaction(s.db, func(tx *gorm.DB) error {
		// Delete likes
		if err := tx.Where("post_id = ?", postID).Delete(&models.PostLike{}).Error; err != nil {
			return utils.NewInternalError("Failed to delete post likes", err).
				WithField("post_id", postID)
		}

		// Delete post
		if err := tx.Delete(&post).Error; err != nil {
			return utils.NewInternalError("Failed to delete post", err).
				WithField("post_id", postID)
		}

		return nil
	})

	if err != nil {
		return err
	}

	if mediaPath != "" && mediaSource == "internal" {
		if err := s.storage.Delete(mediaPath); err != nil {
			log.Warn("Failed to delete media",
				zap.Uint("post_id", postID),
				zap.String("file_path", mediaPath),
				zap.Error(err))
		}
	}

	return nil
}

// LikePost adds a like to a post
func (s *PostService) LikePost(postID, userID uint) (int64, error) {
	// Find post
	var post models.Post
	if result := s.db.First(&post, postID); result.Error != nil {
		return 0, utils.NewNotFoundError("Post not found").
			WithField("post_id", postID)
	}

	// Get the board to check if it's locked
	var board models.Board
	if result := s.db.First(&board, post.BoardID); result.Error != nil {
		return 0, utils.NewNotFoundError("Board not found").
			WithField("post_id", postID)
	}

	// Check if board is locked
	if board.IsLocked {
		return 0, utils.NewForbiddenError("This board is locked and doesn't allow new likes").
			WithField("post_id", postID)
	}

	// Check if user already liked the post
	var existingLike models.PostLike
	result := s.db.Where("post_id = ? AND user_id = ?", postID, userID).First(&existingLike)
	if result.Error == nil {
		return 0, utils.NewBadRequestError("You have already liked this post").
			WithField("post_id", postID)
	}

	// Create like
	like := models.PostLike{
		PostID: postID,
		UserID: userID,
	}

	// Save like
	if result := s.db.Create(&like); result.Error != nil {
		return 0, utils.NewInternalError("Failed to like post", result.Error).
			WithField("post_id", postID)
	}

	// Count total likes
	var likesCount int64
	s.db.Model(&models.PostLike{}).Where("post_id = ?", postID).Count(&likesCount)

	return likesCount, nil
}

// UnlikePost removes a like from a post
func (s *PostService) UnlikePost(postID, userID uint) (int64, error) {
	// Find post
	var post models.Post
	if result := s.db.First(&post, postID); result.Error != nil {
		return 0, utils.NewNotFoundError("Post not found").
			WithField("post_id", postID)
	}

	// Check if user has liked the post
	var like models.PostLike
	result := s.db.Where("post_id = ? AND user_id = ?", postID, userID).First(&like)
	if result.Error != nil {
		return 0, utils.NewBadRequestError("You have not liked this post").
			WithField("post_id", postID)
	}

	// Delete like
	if result := s.db.Delete(&like); result.Error != nil {
		return 0, utils.NewInternalError("Failed to unlike post", result.Error).
			WithField("post_id", postID)
	}

	// Count total likes
	var likesCount int64
	s.db.Model(&models.PostLike{}).Where("post_id = ?", postID).Count(&likesCount)

	return likesCount, nil
}

// ReorderPosts updates the order of posts on a board
func (s *PostService) ReorderPosts(boardID, userID uint, postOrders []requests.PostPosition) error {
	// Find board
	var board models.Board
	if result := s.db.First(&board, boardID); result.Error != nil {
		return utils.NewNotFoundError("Board not found").
			WithField("board_id", boardID)
	}

	// Check if board is locked
	if board.IsLocked {
		return utils.NewForbiddenError("This board is locked and doesn't allow reordering posts").
			WithField("board_id", boardID)
	}

	// Check if user has permission to reorder posts
	if board.CreatorID != userID {
		// Check if user is a contributor with at least 'contributor' role
		var contributor models.BoardContributor
		result := s.db.Where("board_id = ? AND user_id = ? AND role IN ?",
			boardID, userID, []models.Role{models.RoleAdmin}).
			First(&contributor)
		if result.Error != nil {
			return utils.NewForbiddenError("You don't have permission to reorder posts on this board").
				WithField("board_id", boardID).
				WithField("user_id", userID)
		}
	}

	// Start a transaction
	tx := s.db.Begin()

	// Update each post's position
	for _, order := range postOrders {
		// Verify post belongs to this board
		var post models.Post
		if err := tx.Where("id = ? AND board_id = ?", order.ID, boardID).First(&post).Error; err != nil {
			tx.Rollback()
			return utils.NewBadRequestError("Post does not belong to this board").
				WithField("board_id", boardID).
				WithField("post_id", order.ID)
		}

		// Update position
		if err := tx.Model(&post).Update("position", order.Position).Error; err != nil {
			tx.Rollback()
			return utils.NewInternalError("Failed to reorder posts", err).
				WithField("board_id", boardID)
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return utils.NewInternalError("Failed to reorder posts", err).
			WithField("board_id", boardID)
	}

	return nil
}

// GetPostsForBoard gets all posts for a board
func (s *PostService) GetPostsForBoard(boardID uint, page, perPage int, sortBy, order string) ([]models.Post, error) {
	// Build query
	query := s.db.Model(&models.Post{}).Where("board_id = ?", boardID)

	// Add pagination
	offset := (page - 1) * perPage
	query = query.Offset(offset).Limit(perPage)

	// Add ordering
	if sortBy == "" {
		sortBy = "position"
	}
	if order == "" {
		order = "asc"
	}
	orderClause := sortBy + " " + order
	query = query.Order(orderClause)

	// Execute query
	var posts []models.Post
	if result := query.Find(&posts); result.Error != nil {
		return nil, utils.NewInternalError("Failed to fetch posts", result.Error).
			WithField("board_id", boardID)
	}

	return posts, nil
}

// CountPostsInBoard count all posts for a board
func (s *PostService) CountPostsInBoard(boardID uint) int64 {
	// Build query
	query := s.db.Model(&models.Post{}).Where("board_id = ?", boardID)

	// Count total posts
	var total int64
	query.Count(&total)

	return total
}

// CountPostLikes counts the number of likes for a post
func (s *PostService) CountPostLikes(postID uint) (int64, error) {
	var count int64
	if result := s.db.Model(&models.PostLike{}).Where("post_id = ?", postID).Count(&count); result.Error != nil {
		return 0, utils.NewInternalError("Failed to count likes", result.Error).
			WithField("post_id", postID)
	}
	return count, nil
}

// HasUserLikedPost checks if a user has liked a post
func (s *PostService) HasUserLikedPost(postID, userID uint) (bool, error) {
	var like models.PostLike
	result := s.db.Where("post_id = ? AND user_id = ?", postID, userID).First(&like)
	return result.Error == nil, nil
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
