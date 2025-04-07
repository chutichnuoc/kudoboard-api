package services

import (
	"errors"
	"fmt"
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
		Title:                input.Title,
		ReceiverName:         input.ReceiverName,
		CreatorID:            userID,
		FontName:             input.FontName,
		FontSize:             input.FontSize,
		HeaderColor:          input.HeaderColor,
		ThemeID:              input.ThemeID,
		Effect:               input.Effect,
		EnableIntroAnimation: input.EnableIntroAnimation,
		IsPrivate:            input.IsPrivate,
		AllowAnonymous:       input.AllowAnonymous,
	}

	// Use transaction to ensure both operations succeed or fail together
	err := utils.WithTransaction(s.db, func(tx *gorm.DB) error {
		// Save board to database
		if err := tx.Create(&board).Error; err != nil {
			return utils.NewInternalError("Failed to create board", err)
		}

		// Add creator as admin contributor
		contributor := models.BoardContributor{
			BoardID: board.ID,
			UserID:  userID,
			Role:    models.RoleAdmin,
		}

		if err := tx.Create(&contributor).Error; err != nil {
			return utils.NewInternalError("Failed to add creator as admin", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &board, nil
}

// GetBoardByID gets a board by ID
func (s *BoardService) GetBoardByID(boardID uint) (*models.Board, error) {
	var board models.Board
	if result := s.db.First(&board, boardID); result.Error != nil {
		return nil, utils.NewNotFoundError("Board not found").
			WithField("board_id", boardID)
	}

	return &board, nil
}

// GetBoardBySlug gets a board by slug
func (s *BoardService) GetBoardBySlug(slug string) (*models.Board, *models.User, []models.Post, error) {
	// Find board by slug
	var board models.Board
	if result := s.db.Where("slug = ?", slug).First(&board); result.Error != nil {
		return nil, nil, nil, utils.NewNotFoundError("Board not found").
			WithField("slug", slug)
	}

	// Get board creator
	var creator models.User
	if result := s.db.First(&creator, board.CreatorID); result.Error != nil {
		return nil, nil, nil, utils.NewInternalError("Unable to load board information", result.Error).
			WithField("slug", slug)
	}

	// Get posts
	var posts []models.Post
	if result := s.db.Where("board_id = ?", board.ID).Order("created_at desc").Find(&posts); result.Error != nil {
		return nil, nil, nil, utils.NewInternalError("Unable to load board content", result.Error).
			WithField("slug", slug)
	}

	return &board, &creator, posts, nil
}

// UpdateBoard updates a board
func (s *BoardService) UpdateBoard(boardID, userID uint, input requests.UpdateBoardRequest) (*models.Board, error) {
	// Find board
	var board models.Board
	if result := s.db.First(&board, boardID); result.Error != nil {
		return nil, utils.NewNotFoundError("Board not found").
			WithField("board_id", boardID)
	}

	// Check if board is locked
	if board.IsLocked {
		return nil, utils.NewForbiddenError("This board is locked and doesn't allow update").
			WithField("board_id", boardID)
	}

	// Check if user is the creator
	if board.CreatorID != userID {
		return nil, utils.NewForbiddenError("You don't have permission to update this board").
			WithField("board_id", boardID).
			WithField("user_id", userID)
	}

	// Update fields if provided
	if input.Title != nil {
		board.Title = *input.Title
	}
	if input.ReceiverName != nil {
		board.ReceiverName = *input.ReceiverName
	}
	if input.FontName != nil {
		board.FontName = *input.FontName
	}
	if input.FontSize != nil {
		board.FontSize = *input.FontSize
	}
	if input.HeaderColor != nil {
		board.HeaderColor = *input.HeaderColor
	}
	if input.ShowHeaderColor != nil {
		board.ShowHeaderColor = *input.ShowHeaderColor
	}
	if input.ThemeID != nil {
		board.ThemeID = input.ThemeID
	}
	if input.Effect != nil {
		board.Effect = *input.Effect
	}
	if input.EnableIntroAnimation != nil {
		board.EnableIntroAnimation = *input.EnableIntroAnimation
	}
	if input.IsPrivate != nil {
		board.IsPrivate = *input.IsPrivate
	}
	if input.AllowAnonymous != nil {
		board.AllowAnonymous = *input.AllowAnonymous
	}

	// Save changes
	if result := s.db.Save(&board); result.Error != nil {
		return nil, utils.NewInternalError("Failed to update board", result.Error).
			WithField("board_id", boardID)
	}

	return &board, nil
}

// DeleteBoard deletes a board
func (s *BoardService) DeleteBoard(boardID, userID uint) error {
	// Find board
	var board models.Board
	if result := s.db.First(&board, boardID); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return utils.NewNotFoundError("Board not found").
				WithField("board_id", boardID)
		}
		return utils.NewInternalError("Failed to query board", result.Error).
			WithField("board_id", boardID)
	}

	// Check if user is the creator
	if board.CreatorID != userID {
		return utils.NewForbiddenError("You don't have permission to delete this board").
			WithField("board_id", boardID).
			WithField("user_id", userID).
			WithField("creator_id", board.CreatorID)
	}

	// Get all media for posts on this board to delete after transaction
	var posts []models.Post
	if err := s.db.Where("board_id = ?", boardID).Find(&posts).Error; err != nil {
		return utils.NewInternalError("Failed to fetch board posts for media cleanup", err).
			WithField("board_id", boardID)
	}

	// Use transaction for all database operations
	err := utils.WithTransaction(s.db, func(tx *gorm.DB) error {
		// Delete all associated posts likes
		if err := tx.Exec("DELETE FROM post_likes WHERE post_id IN (SELECT id FROM posts WHERE board_id = ?)", boardID).Error; err != nil {
			return utils.NewInternalError("Failed to delete board post likes", err).
				WithField("board_id", boardID)
		}

		// Delete all associated posts
		if err := tx.Where("board_id = ?", boardID).Delete(&models.Post{}).Error; err != nil {
			return utils.NewInternalError("Failed to delete board posts", err).
				WithField("board_id", boardID)
		}

		// Delete all associated contributors
		if err := tx.Where("board_id = ?", boardID).Delete(&models.BoardContributor{}).Error; err != nil {
			return utils.NewInternalError("Failed to delete board contributors", err).
				WithField("board_id", boardID)
		}

		// Delete the board
		if err := tx.Delete(&board).Error; err != nil {
			return utils.NewInternalError("Failed to delete board", err).
				WithField("board_id", boardID)
		}

		return nil
	})

	if err != nil {
		return err
	}

	// Now handle media deletion outside the transaction
	var mediaErrors []string
	for _, post := range posts {
		if post.MediaPath != "" && post.MediaSource == "internal" {
			if err := s.storage.Delete(post.MediaPath); err != nil {
				// Continue attempting to delete other files, but track the error
				mediaErrors = append(mediaErrors, fmt.Sprintf("Media %s: %s", post.MediaPath, err.Error()))
			}
		}
	}

	// If we had media deletion errors, include them in the response
	// but still consider the deletion successful
	if len(mediaErrors) > 0 {
		return utils.NewInternalError("Board deleted but some media files could not be removed",
			fmt.Errorf("media deletion errors: %v", mediaErrors)).
			WithField("board_id", boardID).
			WithField("media_errors", mediaErrors)
	}

	return nil
}

// ToggleBoardLock changes the locked status of a board
func (s *BoardService) ToggleBoardLock(boardID, userID uint, isLocked bool) (*models.Board, error) {
	// Find board
	var board models.Board
	if result := s.db.First(&board, boardID); result.Error != nil {
		return nil, utils.NewNotFoundError("Board not found").
			WithField("board_id", boardID)
	}

	// Check if user is the creator or admin
	if board.CreatorID != userID {
		// Check if user is a board admin
		var contributor models.BoardContributor
		result := s.db.Where("board_id = ? AND user_id = ? AND role = ?",
			boardID, userID, models.RoleAdmin).First(&contributor)
		if result.Error != nil {
			return nil, utils.NewForbiddenError("You don't have permission to lock/unlock this board").
				WithField("board_id", boardID)
		}
	}

	// Update locked status
	board.IsLocked = isLocked

	// Save changes
	if result := s.db.Save(&board); result.Error != nil {
		return nil, utils.NewInternalError("Failed to update board lock status", result.Error).
			WithField("board_id", boardID)
	}

	return &board, nil
}

// ListUserBoards lists all boards where the user is owner or contributor
func (s *BoardService) ListUserBoards(userID uint, page, perPage int, search, sortBy, order string) ([]struct {
	models.Board
	IsOwner    bool
	IsFavorite bool
	IsArchived bool
	Creator    models.User
}, int64, error) {
	// Create a subquery to get all board IDs where user is a contributor
	var contributorBoardIDs []uint
	if err := s.db.Model(&models.BoardContributor{}).
		Select("board_id").
		Where("user_id = ?", userID).
		Find(&contributorBoardIDs).Error; err != nil {
		return nil, 0, utils.NewInternalError("Failed to fetch contributor boards", err).WithField("user_id", userID)
	}

	// Build main query to get all boards where user is creator OR contributor
	query := s.db.Model(&models.Board{}).
		Distinct().
		Where("creator_id = ? OR id IN ?", userID, contributorBoardIDs)

	// Add search if provided
	if search != "" {
		query = query.Where("title LIKE ? OR receiver_name LIKE ?", "%"+search+"%", "%"+search+"%")
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

	// Execute query for boards
	var boards []models.Board
	if result := query.Find(&boards); result.Error != nil {
		return nil, 0, utils.NewInternalError("Failed to fetch boards", result.Error).
			WithField("user_id", userID)
	}

	// Get contributor info for these boards
	var contributors []models.BoardContributor
	if err := s.db.Where("user_id = ? AND board_id IN ?", userID,
		func() []uint {
			ids := make([]uint, len(boards))
			for i, b := range boards {
				ids[i] = b.ID
			}
			return ids
		}()).
		Find(&contributors).Error; err != nil {
		return nil, 0, utils.NewInternalError("Failed to fetch board contributors", err).
			WithField("user_id", userID)
	}

	// Create a map for quick lookup of contributor info
	contributorMap := make(map[uint]models.BoardContributor)
	for _, c := range contributors {
		contributorMap[c.BoardID] = c
	}

	// Build response with additional fields
	result := make([]struct {
		models.Board
		IsOwner    bool
		IsFavorite bool
		IsArchived bool
		Creator    models.User
	}, len(boards))

	for i, board := range boards {
		var creator models.User
		if err := s.db.First(&creator, board.CreatorID).Error; err != nil {
			continue
		}
		result[i].Board = board
		result[i].IsOwner = board.CreatorID == userID

		// Set favorite/archived status from contributor record if it exists
		if contributor, exists := contributorMap[board.ID]; exists {
			result[i].IsFavorite = contributor.IsFavorite
			result[i].IsArchived = contributor.IsArchived
		} else {
			// For boards where user is creator but not in contributors table yet
			result[i].IsFavorite = false
			result[i].IsArchived = false
		}
		result[i].Creator = creator
	}

	return result, total, nil
}

// UpdateBoardPreferences updates a user's preferences for a board (favorite/archived status)
func (s *BoardService) UpdateBoardPreferences(boardID, userID uint, isFavorite, isArchived *bool) error {
	// Find the contributor record
	var contributor models.BoardContributor
	result := s.db.Where("board_id = ? AND user_id = ?", boardID, userID).First(&contributor)
	if result.Error != nil {
		// If the board-user relationship doesn't exist
		return utils.NewNotFoundError("Board not found or you don't have access to it").
			WithField("board_id", boardID).
			WithField("user_id", userID)
	}

	// Update only the fields that are provided
	updates := make(map[string]interface{})

	if isFavorite != nil {
		updates["is_favorite"] = *isFavorite
	}

	if isArchived != nil {
		updates["is_archived"] = *isArchived
	}

	// Only update if there are changes
	if len(updates) > 0 {
		result = s.db.Model(&contributor).Updates(updates)
		if result.Error != nil {
			return utils.NewInternalError("Failed to update board preferences", result.Error).
				WithField("board_id", boardID).
				WithField("user_id", userID)
		}
	}

	return nil
}

// AddContributor adds a contributor to a board
func (s *BoardService) AddContributor(boardID, userID uint, email string, role models.Role) (*models.BoardContributor, *models.User, error) {
	// Find board
	var board models.Board
	if result := s.db.First(&board, boardID); result.Error != nil {
		return nil, nil, utils.NewNotFoundError("Board not found").
			WithField("board_id", boardID)
	}

	// Check if user is the creator
	if board.CreatorID != userID {
		return nil, nil, utils.NewForbiddenError("You don't have permission to add contributors to this board").
			WithField("board_id", boardID).
			WithField("user_id", userID)
	}

	// Find user by email
	var contributorUser models.User
	if result := s.db.Where("email = ?", email).First(&contributorUser); result.Error != nil {
		return nil, nil, utils.NewNotFoundError("User not found with this email").
			WithField("email", email)
	}

	// Check if user is already a contributor
	var existingContributor models.BoardContributor
	result := s.db.Where("board_id = ? AND user_id = ?", boardID, contributorUser.ID).First(&existingContributor)
	if result.Error == nil {
		return nil, nil, utils.NewBadRequestError("User is already a contributor to this board").
			WithField("board_id", boardID).
			WithField("contributor_id", contributorUser.ID)
	}

	// Create new contributor
	contributor := models.BoardContributor{
		BoardID: boardID,
		UserID:  contributorUser.ID,
		Role:    role,
	}

	// Save to database
	if result := s.db.Create(&contributor); result.Error != nil {
		return nil, nil, utils.NewInternalError("Failed to add contributor", result.Error).
			WithField("board_id", boardID)
	}

	return &contributor, &contributorUser, nil
}

// UpdateContributor updates a contributor's role
func (s *BoardService) UpdateContributor(boardID, userID, contributorID uint, role models.Role) (*models.BoardContributor, *models.User, error) {
	// Find board
	var board models.Board
	if result := s.db.First(&board, boardID); result.Error != nil {
		return nil, nil, utils.NewNotFoundError("Board not found").
			WithField("board_id", boardID)
	}

	// Check if user is the creator
	if board.CreatorID != userID {
		return nil, nil, utils.NewForbiddenError("You don't have permission to update contributors for this board").
			WithField("board_id", boardID).
			WithField("user_id", userID)
	}

	// Find contributor
	var contributor models.BoardContributor
	result := s.db.Where("board_id = ? AND user_id = ?", boardID, contributorID).First(&contributor)
	if result.Error != nil {
		return nil, nil, utils.NewNotFoundError("Contributor not found").
			WithField("board_id", boardID).
			WithField("contributor_id", contributorID)
	}

	// Update role
	contributor.Role = role

	// Save to database
	if result := s.db.Save(&contributor); result.Error != nil {
		return nil, nil, utils.NewInternalError("Failed to update contributor", result.Error).
			WithField("board_id", boardID).
			WithField("contributor_id", contributorID)
	}

	// Get contributor user
	var contributorUser models.User
	if result := s.db.First(&contributorUser, contributorID); result.Error != nil {
		return nil, nil, utils.NewInternalError("Failed to get contributor user", result.Error).
			WithField("board_id", boardID).
			WithField("contributor_id", contributorID)
	}

	return &contributor, &contributorUser, nil
}

// RemoveContributor removes a contributor from a board
func (s *BoardService) RemoveContributor(boardID, userID, contributorID uint) error {
	// Find board
	var board models.Board
	if result := s.db.First(&board, boardID); result.Error != nil {
		return utils.NewNotFoundError("Board not found").
			WithField("board_id", boardID)
	}

	// Check if user is the creator
	if board.CreatorID != userID {
		return utils.NewForbiddenError("You don't have permission to remove contributors from this board").
			WithField("board_id", boardID).
			WithField("user_id", userID)
	}

	// Find contributor
	var contributor models.BoardContributor
	result := s.db.Where("board_id = ? AND user_id = ?", boardID, contributorID).First(&contributor)
	if result.Error != nil {
		return utils.NewNotFoundError("Contributor not found").
			WithField("board_id", boardID).
			WithField("contributor_id", contributorID)
	}

	// Delete contributor
	if result := s.db.Delete(&contributor); result.Error != nil {
		return utils.NewInternalError("Failed to remove contributor", result.Error).
			WithField("board_id", boardID).
			WithField("contributor_id", contributorID)
	}

	return nil
}

// ListBoardContributors lists all contributors for a board
func (s *BoardService) ListBoardContributors(boardID, userID uint) ([]models.BoardContributor, []models.User, error) {
	// Find board
	var board models.Board
	if result := s.db.First(&board, boardID); result.Error != nil {
		return nil, nil, utils.NewNotFoundError("Board not found").
			WithField("board_id", boardID)
	}

	// Check if user is the creator or a contributor
	if board.CreatorID != userID {
		var contributor models.BoardContributor
		result := s.db.Where("board_id = ? AND user_id = ?", boardID, userID).First(&contributor)
		if result.Error != nil {
			return nil, nil, utils.NewForbiddenError("You don't have permission to view this board's contributors").
				WithField("board_id", boardID).
				WithField("user_id", userID)
		}
	}

	// Get contributors
	var contributors []models.BoardContributor
	if err := s.db.Where("board_id = ?", boardID).Find(&contributors).Error; err != nil {
		return nil, nil, utils.NewInternalError("Failed to fetch contributors", err).
			WithField("board_id", boardID)
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

// CanAccessBoard checks if a user has access to a board
func (s *BoardService) CanAccessBoard(boardID, userID uint) (bool, error) {
	// Find board
	var board models.Board
	if result := s.db.First(&board, boardID); result.Error != nil {
		return false, utils.NewNotFoundError("Board not found").
			WithField("board_id", boardID)
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
