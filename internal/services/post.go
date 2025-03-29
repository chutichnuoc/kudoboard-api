package services

import (
	"gorm.io/gorm"
	"kudoboard-api/internal/config"
	"kudoboard-api/internal/dto/requests"
	"kudoboard-api/internal/models"
	"kudoboard-api/internal/services/storage"
	"kudoboard-api/internal/utils"
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

// CreatePost creates a new post for an authenticated user
func (s *PostService) CreatePost(boardID, userID uint, input requests.CreatePostRequest) (*models.Post, error) {
	// Check if user has access to the board
	canAccess, err := s.boardService.CanAccessBoard(boardID, userID)
	if err != nil {
		return nil, err
	}
	if !canAccess {
		return nil, utils.NewForbiddenError("You don't have access to this board")
	}

	// Get board to check if it allows anonymous posts
	var board models.Board
	if result := s.db.First(&board, boardID); result.Error != nil {
		return nil, utils.NewNotFoundError("Board not found")
	}

	// If this is an anonymous post, check if the board allows it
	if input.IsAnonymous && !board.AllowAnonymous {
		return nil, utils.NewForbiddenError("This board does not allow anonymous posts")
	}

	// Get user for author details
	var user models.User
	if result := s.db.First(&user, userID); result.Error != nil {
		return nil, utils.NewInternalError("Failed to get user", result.Error)
	}

	// Get max position_order for this board
	var maxOrder int
	s.db.Model(&models.Post{}).Where("board_id = ?", boardID).Select("COALESCE(MAX(position_order), 0)").Scan(&maxOrder)

	// Create post
	post := models.Post{
		BoardID:         boardID,
		Content:         input.Content,
		BackgroundColor: input.BackgroundColor,
		TextColor:       input.TextColor,
		PositionX:       input.PositionX,
		PositionY:       input.PositionY,
		PositionOrder:   maxOrder + 1,
		IsAnonymous:     input.IsAnonymous,
	}

	// Set author details based on anonymity
	if input.IsAnonymous {
		post.AuthorName = input.AuthorName
		post.AuthorEmail = input.AuthorEmail
	} else {
		post.AuthorID = &userID
		post.AuthorName = user.Name
		post.AuthorEmail = user.Email
	}

	// Save post
	if result := s.db.Create(&post); result.Error != nil {
		return nil, utils.NewInternalError("Failed to create post", result.Error)
	}

	return &post, nil
}

// GetPostByID gets a post by ID
func (s *PostService) GetPostByID(postID uint) (*models.Post, error) {
	var post models.Post
	if result := s.db.First(&post, postID); result.Error != nil {
		return nil, utils.NewNotFoundError("Post not found")
	}
	return &post, nil
}

// UpdatePost updates a post
func (s *PostService) UpdatePost(postID, userID uint, input requests.UpdatePostRequest) (*models.Post, error) {
	// Find post
	var post models.Post
	if result := s.db.First(&post, postID); result.Error != nil {
		return nil, utils.NewNotFoundError("Post not found")
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
				return nil, utils.NewForbiddenError("You don't have permission to update this post")
			}
		}
	}

	// Update fields if provided
	if input.Content != nil {
		post.Content = *input.Content
	}
	if input.BackgroundColor != nil {
		post.BackgroundColor = *input.BackgroundColor
	}
	if input.TextColor != nil {
		post.TextColor = *input.TextColor
	}
	if input.PositionX != nil {
		post.PositionX = *input.PositionX
	}
	if input.PositionY != nil {
		post.PositionY = *input.PositionY
	}

	// Save changes
	if result := s.db.Save(&post); result.Error != nil {
		return nil, utils.NewInternalError("Failed to update post", result.Error)
	}

	return &post, nil
}

// DeletePost deletes a post
func (s *PostService) DeletePost(postID, userID uint) error {
	// Find post
	var post models.Post
	if result := s.db.First(&post, postID); result.Error != nil {
		return utils.NewNotFoundError("Post not found")
	}

	// Check if user has permission to delete this post
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
				return utils.NewForbiddenError("You don't have permission to delete this post")
			}
		}
	}

	// Start a transaction
	tx := s.db.Begin()

	// Delete likes
	if err := tx.Where("post_id = ?", postID).Delete(&models.PostLike{}).Error; err != nil {
		tx.Rollback()
		return utils.NewInternalError("Failed to delete post", err)
	}

	// Get media for this post
	var media []models.Media
	if err := tx.Where("post_id = ?", postID).Find(&media).Error; err != nil {
		tx.Rollback()
		return utils.NewInternalError("Failed to delete post", err)
	}

	// Delete media files from storage
	for _, m := range media {
		if m.SourceType == models.SourceTypeUpload {
			if err := s.storage.Delete(m.SourceURL); err != nil {
				// Log error but continue (we don't want to fail the entire transaction for a storage error)
			}
		}

		// Delete media record
		if err := tx.Delete(&m).Error; err != nil {
			tx.Rollback()
			return utils.NewInternalError("Failed to delete post", err)
		}
	}

	// Delete post
	if err := tx.Delete(&post).Error; err != nil {
		tx.Rollback()
		return utils.NewInternalError("Failed to delete post", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return utils.NewInternalError("Failed to delete post", err)
	}

	return nil
}

// LikePost adds a like to a post
func (s *PostService) LikePost(postID, userID uint) (int64, error) {
	// Find post
	var post models.Post
	if result := s.db.First(&post, postID); result.Error != nil {
		return 0, utils.NewNotFoundError("Post not found")
	}

	// Check if user already liked the post
	var existingLike models.PostLike
	result := s.db.Where("post_id = ? AND user_id = ?", postID, userID).First(&existingLike)
	if result.Error == nil {
		return 0, utils.NewBadRequestError("You have already liked this post")
	}

	// Create like
	like := models.PostLike{
		PostID: postID,
		UserID: userID,
	}

	// Save like
	if result := s.db.Create(&like); result.Error != nil {
		return 0, utils.NewInternalError("Failed to like post", result.Error)
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
		return 0, utils.NewNotFoundError("Post not found")
	}

	// Check if user has liked the post
	var like models.PostLike
	result := s.db.Where("post_id = ? AND user_id = ?", postID, userID).First(&like)
	if result.Error != nil {
		return 0, utils.NewBadRequestError("You have not liked this post")
	}

	// Delete like
	if result := s.db.Delete(&like); result.Error != nil {
		return 0, utils.NewInternalError("Failed to unlike post", result.Error)
	}

	// Count total likes
	var likesCount int64
	s.db.Model(&models.PostLike{}).Where("post_id = ?", postID).Count(&likesCount)

	return likesCount, nil
}

// ReorderPosts updates the order of posts on a board
func (s *PostService) ReorderPosts(boardID, userID uint, postOrders []requests.PostOrder) error {
	// Find board
	var board models.Board
	if result := s.db.First(&board, boardID); result.Error != nil {
		return utils.NewNotFoundError("Board not found")
	}

	// Check if user has permission to reorder posts
	if board.CreatorID != userID {
		// Check if user is a contributor with at least 'contributor' role
		var contributor models.BoardContributor
		result := s.db.Where("board_id = ? AND user_id = ? AND role IN ?",
			boardID, userID, []models.Role{models.RoleContributor, models.RoleAdmin}).
			First(&contributor)
		if result.Error != nil {
			return utils.NewForbiddenError("You don't have permission to reorder posts on this board")
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
			return utils.NewBadRequestError("Post with ID " + string(order.ID) + " does not belong to this board")
		}

		// Update position
		if err := tx.Model(&post).Update("position_order", order.PositionOrder).Error; err != nil {
			tx.Rollback()
			return utils.NewInternalError("Failed to reorder posts", err)
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return utils.NewInternalError("Failed to reorder posts", err)
	}

	return nil
}

// GetPostsForBoard gets all posts for a board
func (s *PostService) GetPostsForBoard(boardID uint, page, perPage int, sortBy, order string) ([]models.Post, int64, error) {
	// Build query
	query := s.db.Model(&models.Post{}).Where("board_id = ?", boardID)

	// Count total posts
	var total int64
	query.Count(&total)

	// Add pagination
	offset := (page - 1) * perPage
	query = query.Offset(offset).Limit(perPage)

	// Add ordering
	if sortBy == "" {
		sortBy = "position_order"
	}
	if order == "" {
		order = "asc"
	}
	orderClause := sortBy + " " + order
	query = query.Order(orderClause)

	// Execute query
	var posts []models.Post
	if result := query.Find(&posts); result.Error != nil {
		return nil, 0, utils.NewInternalError("Failed to fetch posts", result.Error)
	}

	return posts, total, nil
}

// GetMediaForPost gets all media for a post
func (s *PostService) GetMediaForPost(postID uint) ([]models.Media, error) {
	var media []models.Media
	if result := s.db.Where("post_id = ?", postID).Find(&media); result.Error != nil {
		return nil, utils.NewInternalError("Failed to fetch media", result.Error)
	}
	return media, nil
}

// CountPostLikes counts the number of likes for a post
func (s *PostService) CountPostLikes(postID uint) (int64, error) {
	var count int64
	if result := s.db.Model(&models.PostLike{}).Where("post_id = ?", postID).Count(&count); result.Error != nil {
		return 0, utils.NewInternalError("Failed to count likes", result.Error)
	}
	return count, nil
}

// HasUserLikedPost checks if a user has liked a post
func (s *PostService) HasUserLikedPost(postID, userID uint) (bool, error) {
	var like models.PostLike
	result := s.db.Where("post_id = ? AND user_id = ?", postID, userID).First(&like)
	return result.Error == nil, nil
}
