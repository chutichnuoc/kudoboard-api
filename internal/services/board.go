package services

import (
	"gorm.io/gorm"
	"kudoboard-api/internal/config"
	"kudoboard-api/internal/dto/requests"
	"kudoboard-api/internal/models"
	"kudoboard-api/internal/services/storage"
	"kudoboard-api/internal/utils"
)

// BoardService handles board-related business logic
type BoardService struct {
	db      *gorm.DB
	storage storage.StorageService
	cfg     *config.Config
}

// NewBoardService creates a new BoardService
func NewBoardService(db *gorm.DB, storage storage.StorageService, cfg *config.Config) *BoardService {
	return &BoardService{
		db:      db,
		storage: storage,
		cfg:     cfg,
	}
}

// CreateBoard creates a new board
func (s *BoardService) CreateBoard(userID uint, input requests.CreateBoardRequest) (*models.Board, error) {
	// Create new board
	board := models.Board{
		Title:              input.Title,
		Description:        input.Description,
		CreatorID:          userID,
		BackgroundType:     input.BackgroundType,
		BackgroundImageURL: input.BackgroundImageURL,
		BackgroundColor:    input.BackgroundColor,
		ThemeID:            input.ThemeID,
		IsPrivate:          input.IsPrivate,
		AllowAnonymous:     input.AllowAnonymous,
		ExpiresAt:          input.ExpiresAt,
	}

	// Save board to database
	if result := s.db.Create(&board); result.Error != nil {
		return nil, utils.NewInternalError("Failed to create board", result.Error)
	}

	return &board, nil
}

// GetBoardByID gets a board by ID
func (s *BoardService) GetBoardByID(boardID uint) (*models.Board, error) {
	var board models.Board
	if result := s.db.First(&board, boardID); result.Error != nil {
		return nil, utils.NewNotFoundError("Board not found")
	}

	return &board, nil
}

// GetBoardBySlug gets a board by slug
func (s *BoardService) GetBoardBySlug(slug string) (*models.Board, *models.User, []models.Post, error) {
	// Find board by slug
	var board models.Board
	if result := s.db.Where("slug = ?", slug).First(&board); result.Error != nil {
		return nil, nil, nil, utils.NewNotFoundError("Board not found")
	}

	// Get board creator
	var creator models.User
	if result := s.db.First(&creator, board.CreatorID); result.Error != nil {
		return nil, nil, nil, utils.NewInternalError("Failed to fetch board creator", result.Error)
	}

	// Get posts
	var posts []models.Post
	if result := s.db.Where("board_id = ?", board.ID).Order("position_order asc, created_at desc").Find(&posts); result.Error != nil {
		return nil, nil, nil, utils.NewInternalError("Failed to fetch posts", result.Error)
	}

	return &board, &creator, posts, nil
}

// UpdateBoard updates a board
func (s *BoardService) UpdateBoard(boardID, userID uint, input requests.UpdateBoardRequest) (*models.Board, error) {
	// Find board
	var board models.Board
	if result := s.db.First(&board, boardID); result.Error != nil {
		return nil, utils.NewNotFoundError("Board not found")
	}

	// Check if user is the creator
	if board.CreatorID != userID {
		return nil, utils.NewForbiddenError("You don't have permission to update this board")
	}

	// Update fields if provided
	if input.Title != nil {
		board.Title = *input.Title
	}
	if input.Description != nil {
		board.Description = *input.Description
	}
	if input.BackgroundType != nil {
		board.BackgroundType = *input.BackgroundType
	}
	if input.BackgroundImageURL != nil {
		board.BackgroundImageURL = *input.BackgroundImageURL
	}
	if input.BackgroundColor != nil {
		board.BackgroundColor = *input.BackgroundColor
	}
	if input.ThemeID != nil {
		board.ThemeID = input.ThemeID
	}
	if input.IsPrivate != nil {
		board.IsPrivate = *input.IsPrivate
	}
	if input.AllowAnonymous != nil {
		board.AllowAnonymous = *input.AllowAnonymous
	}
	if input.ExpiresAt != nil {
		board.ExpiresAt = input.ExpiresAt
	}

	// Save changes
	if result := s.db.Save(&board); result.Error != nil {
		return nil, utils.NewInternalError("Failed to update board", result.Error)
	}

	return &board, nil
}

// DeleteBoard deletes a board
func (s *BoardService) DeleteBoard(boardID, userID uint) error {
	// Find board
	var board models.Board
	if result := s.db.First(&board, boardID); result.Error != nil {
		return utils.NewNotFoundError("Board not found")
	}

	// Check if user is the creator
	if board.CreatorID != userID {
		return utils.NewForbiddenError("You don't have permission to delete this board")
	}

	// Start a transaction
	tx := s.db.Begin()

	// Delete all associated posts likes
	if err := tx.Exec("DELETE FROM post_likes WHERE post_id IN (SELECT id FROM posts WHERE board_id = ?)", boardID).Error; err != nil {
		tx.Rollback()
		return utils.NewInternalError("Failed to delete board", err)
	}

	// Get all media for posts on this board
	var media []models.Media
	if err := tx.Where("post_id IN (SELECT id FROM posts WHERE board_id = ?)", boardID).Find(&media).Error; err != nil {
		tx.Rollback()
		return utils.NewInternalError("Failed to delete board", err)
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
			return utils.NewInternalError("Failed to delete board", err)
		}
	}

	// Delete all associated posts
	if err := tx.Where("board_id = ?", boardID).Delete(&models.Post{}).Error; err != nil {
		tx.Rollback()
		return utils.NewInternalError("Failed to delete board", err)
	}

	// Delete all associated contributors
	if err := tx.Where("board_id = ?", boardID).Delete(&models.BoardContributor{}).Error; err != nil {
		tx.Rollback()
		return utils.NewInternalError("Failed to delete board", err)
	}

	// Delete the board
	if err := tx.Delete(&board).Error; err != nil {
		tx.Rollback()
		return utils.NewInternalError("Failed to delete board", err)
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		return utils.NewInternalError("Failed to delete board", err)
	}

	return nil
}

// ListUserBoards lists all boards created by a user
func (s *BoardService) ListUserBoards(userID uint, page, perPage int, search, sortBy, order string) ([]models.Board, int64, error) {
	// Build query
	query := s.db.Model(&models.Board{}).Where("creator_id = ?", userID)

	// Add search if provided
	if search != "" {
		query = query.Where("title LIKE ? OR description LIKE ?", "%"+search+"%", "%"+search+"%")
	}

	// Count total boards
	var total int64
	query.Count(&total)

	// Add pagination
	offset := (page - 1) * perPage
	query = query.Offset(offset).Limit(perPage)

	// Add ordering
	if sortBy == "" {
		sortBy = "created_at"
	}
	if order == "" {
		order = "desc"
	}
	orderClause := sortBy + " " + order
	query = query.Order(orderClause)

	// Execute query
	var boards []models.Board
	if result := query.Find(&boards); result.Error != nil {
		return nil, 0, utils.NewInternalError("Failed to fetch boards", result.Error)
	}

	return boards, total, nil
}

// ListPublicBoards lists all public boards
func (s *BoardService) ListPublicBoards(page, perPage int, search, sortBy, order string) ([]models.Board, int64, error) {
	// Build query for public boards
	query := s.db.Model(&models.Board{}).Where("is_private = ?", false)

	// Add search if provided
	if search != "" {
		query = query.Where("title LIKE ? OR description LIKE ?", "%"+search+"%", "%"+search+"%")
	}

	// Count total boards
	var total int64
	query.Count(&total)

	// Add pagination
	offset := (page - 1) * perPage
	query = query.Offset(offset).Limit(perPage)

	// Add ordering
	if sortBy == "" {
		sortBy = "created_at"
	}
	if order == "" {
		order = "desc"
	}
	orderClause := sortBy + " " + order
	query = query.Order(orderClause)

	// Execute query
	var boards []models.Board
	if result := query.Find(&boards); result.Error != nil {
		return nil, 0, utils.NewInternalError("Failed to fetch boards", result.Error)
	}

	return boards, total, nil
}

// AddContributor adds a contributor to a board
func (s *BoardService) AddContributor(boardID, userID uint, email string, role models.Role) (*models.BoardContributor, *models.User, error) {
	// Find board
	var board models.Board
	if result := s.db.First(&board, boardID); result.Error != nil {
		return nil, nil, utils.NewNotFoundError("Board not found")
	}

	// Check if user is the creator
	if board.CreatorID != userID {
		return nil, nil, utils.NewForbiddenError("You don't have permission to add contributors to this board")
	}

	// Find user by email
	var contributorUser models.User
	if result := s.db.Where("email = ?", email).First(&contributorUser); result.Error != nil {
		return nil, nil, utils.NewNotFoundError("User not found with this email")
	}

	// Check if user is already a contributor
	var existingContributor models.BoardContributor
	result := s.db.Where("board_id = ? AND user_id = ?", boardID, contributorUser.ID).First(&existingContributor)
	if result.Error == nil {
		return nil, nil, utils.NewBadRequestError("User is already a contributor to this board")
	}

	// Create new contributor
	contributor := models.BoardContributor{
		BoardID: boardID,
		UserID:  contributorUser.ID,
		Role:    role,
	}

	// Save to database
	if result := s.db.Create(&contributor); result.Error != nil {
		return nil, nil, utils.NewInternalError("Failed to add contributor", result.Error)
	}

	return &contributor, &contributorUser, nil
}

// UpdateContributor updates a contributor's role
func (s *BoardService) UpdateContributor(boardID, userID, contributorID uint, role models.Role) (*models.BoardContributor, *models.User, error) {
	// Find board
	var board models.Board
	if result := s.db.First(&board, boardID); result.Error != nil {
		return nil, nil, utils.NewNotFoundError("Board not found")
	}

	// Check if user is the creator
	if board.CreatorID != userID {
		return nil, nil, utils.NewForbiddenError("You don't have permission to update contributors for this board")
	}

	// Find contributor
	var contributor models.BoardContributor
	result := s.db.Where("board_id = ? AND user_id = ?", boardID, contributorID).First(&contributor)
	if result.Error != nil {
		return nil, nil, utils.NewNotFoundError("Contributor not found")
	}

	// Update role
	contributor.Role = role

	// Save to database
	if result := s.db.Save(&contributor); result.Error != nil {
		return nil, nil, utils.NewInternalError("Failed to update contributor", result.Error)
	}

	// Get contributor user
	var contributorUser models.User
	if result := s.db.First(&contributorUser, contributorID); result.Error != nil {
		return nil, nil, utils.NewInternalError("Failed to get contributor user", result.Error)
	}

	return &contributor, &contributorUser, nil
}

// RemoveContributor removes a contributor from a board
func (s *BoardService) RemoveContributor(boardID, userID, contributorID uint) error {
	// Find board
	var board models.Board
	if result := s.db.First(&board, boardID); result.Error != nil {
		return utils.NewNotFoundError("Board not found")
	}

	// Check if user is the creator
	if board.CreatorID != userID {
		return utils.NewForbiddenError("You don't have permission to remove contributors from this board")
	}

	// Find contributor
	var contributor models.BoardContributor
	result := s.db.Where("board_id = ? AND user_id = ?", boardID, contributorID).First(&contributor)
	if result.Error != nil {
		return utils.NewNotFoundError("Contributor not found")
	}

	// Delete contributor
	if result := s.db.Delete(&contributor); result.Error != nil {
		return utils.NewInternalError("Failed to remove contributor", result.Error)
	}

	return nil
}

// ListBoardContributors lists all contributors for a board
func (s *BoardService) ListBoardContributors(boardID, userID uint) ([]models.BoardContributor, []models.User, error) {
	// Find board
	var board models.Board
	if result := s.db.First(&board, boardID); result.Error != nil {
		return nil, nil, utils.NewNotFoundError("Board not found")
	}

	// Check if user is the creator or a contributor
	if board.CreatorID != userID {
		var contributor models.BoardContributor
		result := s.db.Where("board_id = ? AND user_id = ?", boardID, userID).First(&contributor)
		if result.Error != nil {
			return nil, nil, utils.NewForbiddenError("You don't have permission to view this board's contributors")
		}
	}

	// Get contributors
	var contributors []models.BoardContributor
	if err := s.db.Where("board_id = ?", boardID).Find(&contributors).Error; err != nil {
		return nil, nil, utils.NewInternalError("Failed to fetch contributors", err)
	}

	// Get users for each contributor
	var users []models.User
	for _, contributor := range contributors {
		var user models.User
		if err := s.db.First(&user, contributor.UserID).Error; err != nil {
			continue // Skip if user not found
		}
		users = append(users, user)
	}

	return contributors, users, nil
}

// GetThemes gets all available themes
func (s *BoardService) GetThemes() ([]models.Theme, error) {
	var themes []models.Theme
	if result := s.db.Find(&themes); result.Error != nil {
		return nil, utils.NewInternalError("Failed to fetch themes", result.Error)
	}
	return themes, nil
}

// GetThemeByID gets a theme by ID
func (s *BoardService) GetThemeByID(themeID uint) (*models.Theme, error) {
	var theme models.Theme
	if result := s.db.First(&theme, themeID); result.Error != nil {
		return nil, utils.NewNotFoundError("Theme not found")
	}
	return &theme, nil
}

// CanAccessBoard checks if a user has access to a board
func (s *BoardService) CanAccessBoard(boardID, userID uint) (bool, error) {
	// Find board
	var board models.Board
	if result := s.db.First(&board, boardID); result.Error != nil {
		return false, utils.NewNotFoundError("Board not found")
	}

	// If board is public, anyone can access
	if !board.IsPrivate {
		return true, nil
	}

	// If user is the creator, they can access
	if board.CreatorID == userID {
		return true, nil
	}

	// Check if user is a contributor
	var contributor models.BoardContributor
	result := s.db.Where("board_id = ? AND user_id = ?", boardID, userID).First(&contributor)
	return result.Error == nil, nil
}
