package handlers

import (
	"github.com/gin-gonic/gin"
	"kudoboard-api/internal/config"
	"kudoboard-api/internal/dto/requests"
	"kudoboard-api/internal/dto/responses"
	"kudoboard-api/internal/models"
	"kudoboard-api/internal/services"
	"kudoboard-api/internal/utils"
	"net/http"
	"strconv"
)

// BoardHandler handles board-related requests
type BoardHandler struct {
	boardService *services.BoardService
	postService  *services.PostService
	themeService *services.ThemeService
	authService  *services.AuthService
	cfg          *config.Config
}

// NewBoardHandler creates a new BoardHandler
func NewBoardHandler(boardService *services.BoardService, postService *services.PostService, themeService *services.ThemeService, authService *services.AuthService, cfg *config.Config) *BoardHandler {
	return &BoardHandler{
		boardService: boardService,
		postService:  postService,
		themeService: themeService,
		authService:  authService,
		cfg:          cfg,
	}
}

// CreateBoard handles the creation of a new board
func (h *BoardHandler) CreateBoard(c *gin.Context) {
	// Get user ID from context
	userID := c.GetUint("userID")
	if userID == 0 {
		_ = c.Error(utils.NewUnauthorizedError("User not authenticated"))
		return
	}

	// Parse request
	var req requests.CreateBoardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(utils.NewValidationError(err.Error()))
		return
	}

	// Create board using service
	board, err := h.boardService.CreateBoard(userID, req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	// Get user for response
	user, _ := c.Get("user")

	c.JSON(http.StatusCreated, responses.SuccessResponse(
		responses.NewBoardResponse(board, user.(*models.User), 0),
	))
}

// ListUserBoards lists all boards created by or contributed to by the current user
func (h *BoardHandler) ListUserBoards(c *gin.Context) {
	// Get user ID from context
	userID := c.GetUint("userID")
	if userID == 0 {
		_ = c.Error(utils.NewUnauthorizedError("User not authenticated"))
		return
	}

	// Parse query parameters
	var query requests.BoardQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		_ = c.Error(utils.NewValidationError(err.Error()))
		return
	}

	// Set defaults if not provided
	if query.Page < 1 {
		query.Page = 1
	}
	if query.PerPage < 1 {
		query.PerPage = 10
	}

	// Get boards using service
	boardsWithInfo, total, err := h.boardService.ListUserBoards(userID, query.Page, query.PerPage, query.Search, query.SortBy, query.Order)
	if err != nil {
		_ = c.Error(err)
		return
	}

	// Build response
	boardResponses := make([]responses.BoardResponseWithRelation, len(boardsWithInfo))
	for i, boardInfo := range boardsWithInfo {
		posts, _ := h.postService.GetPostsForBoard(boardInfo.ID, 1, 0, "", "") // 0 for PerPage means all posts

		// Create response
		boardResponses[i] = responses.NewBoardResponseWithRelation(
			&boardInfo.Board,
			&boardInfo.Creator,
			len(posts),
			boardInfo.IsOwner,
			boardInfo.IsFavorite,
			boardInfo.IsArchived,
		)
	}

	// Create pagination info
	pagination := &responses.Pagination{
		Total:      total,
		Page:       query.Page,
		PerPage:    query.PerPage,
		TotalPages: int((total + int64(query.PerPage) - 1) / int64(query.PerPage)),
	}

	c.JSON(http.StatusOK, responses.SuccessResponseWithPagination(boardResponses, pagination))
}

// GetBoardBySlug gets a board by its slug
func (h *BoardHandler) GetBoardBySlug(c *gin.Context) {
	slug := c.Param("slug")

	// Get current user if authenticated
	var userID uint
	user, exists := c.Get("user")
	if exists && user != nil {
		userID = user.(*models.User).ID
	}

	// Get board by slug using service
	board, creator, posts, err := h.boardService.GetBoardBySlug(slug)
	if err != nil {
		_ = c.Error(err)
		return
	}

	// Check if board is private and user is not creator
	if board.IsPrivate && (userID == 0 || userID != board.CreatorID) {
		// Check if user is a contributor
		canAccess, _ := h.boardService.CanAccessBoard(board.ID, userID)
		if !canAccess {
			_ = c.Error(utils.NewForbiddenError("You don't have access to this board"))
			return
		}
	}

	// Create board response
	boardResponse := responses.NewBoardResponse(board, creator, len(posts))

	// If board has a theme, include it
	if board.ThemeID != nil {
		theme, err := h.themeService.GetThemeByID(*board.ThemeID)
		if err == nil {
			themeResponse := responses.NewThemeResponse(theme)
			boardResponse.Theme = &themeResponse
		}
	}

	// Create post responses
	postResponses := make([]responses.PostResponse, len(posts))
	for i, post := range posts {
		// Get post author if not anonymous
		var author *models.User
		if post.AuthorID != nil {
			author, _ = h.authService.GetUserByID(*post.AuthorID)
		}

		// Count likes
		likesCount, _ := h.postService.CountPostLikes(post.ID)

		// Create post response
		postResponses[i] = responses.NewPostResponse(&post, author, likesCount)
	}

	// Add posts to response
	response := gin.H{
		"board": boardResponse,
		"posts": postResponses,
	}

	c.JSON(http.StatusOK, responses.SuccessResponse(response))
}

// UpdateBoard updates a board
func (h *BoardHandler) UpdateBoard(c *gin.Context) {
	// Get user ID from context
	userID := c.GetUint("userID")
	if userID == 0 {
		_ = c.Error(utils.NewUnauthorizedError("User not authenticated"))
		return
	}

	// Get board ID from URL
	boardID, err := strconv.ParseUint(c.Param("boardId"), 10, 32)
	if err != nil {
		_ = c.Error(utils.NewBadRequestError("Invalid board ID"))
		return
	}

	// Parse request
	var req requests.UpdateBoardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(utils.NewValidationError(err.Error()))
		return
	}

	// Update board using service
	board, err := h.boardService.UpdateBoard(uint(boardID), userID, req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	// Get user for response
	user, _ := c.Get("user")

	// Count posts
	posts, _ := h.postService.GetPostsForBoard(uint(boardID), 1, 0, "", "")

	c.JSON(http.StatusOK, responses.SuccessResponse(
		responses.NewBoardResponse(board, user.(*models.User), len(posts)),
	))
}

// DeleteBoard deletes a board
func (h *BoardHandler) DeleteBoard(c *gin.Context) {
	// Get user ID from context
	userID := c.GetUint("userID")
	if userID == 0 {
		_ = c.Error(utils.NewUnauthorizedError("User not authenticated"))
		return
	}

	// Get board ID from URL
	boardID, err := strconv.ParseUint(c.Param("boardId"), 10, 32)
	if err != nil {
		_ = c.Error(utils.NewBadRequestError("Invalid board ID format"))
		return
	}

	// Delete board using service
	err = h.boardService.DeleteBoard(uint(boardID), userID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, responses.SuccessResponse(gin.H{"message": "Board deleted successfully"}))
}

// ToggleBoardLock handles locking or unlocking a board
func (h *BoardHandler) ToggleBoardLock(c *gin.Context) {
	// Get user ID from context
	userID := c.GetUint("userID")
	if userID == 0 {
		_ = c.Error(utils.NewUnauthorizedError("User not authenticated"))
		return
	}

	// Get board ID from URL
	boardID, err := strconv.ParseUint(c.Param("boardId"), 10, 32)
	if err != nil {
		_ = c.Error(utils.NewBadRequestError("Invalid board ID"))
		return
	}

	// Parse request
	var req requests.LockBoardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(utils.NewValidationError(err.Error()))
		return
	}

	// Toggle board lock using service
	board, err := h.boardService.ToggleBoardLock(uint(boardID), userID, req.IsLocked)
	if err != nil {
		_ = c.Error(err)
		return
	}

	// Get user for response
	user, _ := c.Get("user")

	// Count posts
	posts, _ := h.postService.GetPostsForBoard(uint(boardID), 1, 0, "", "")

	c.JSON(http.StatusOK, responses.SuccessResponse(
		responses.NewBoardResponse(board, user.(*models.User), len(posts)),
	))
}

// ListBoardContributors lists all contributors for a board
func (h *BoardHandler) ListBoardContributors(c *gin.Context) {
	// Get user ID from context
	userID := c.GetUint("userID")
	if userID == 0 {
		_ = c.Error(utils.NewUnauthorizedError("User not authenticated"))
		return
	}

	// Get board ID from URL
	boardID, err := strconv.ParseUint(c.Param("boardId"), 10, 32)
	if err != nil {
		_ = c.Error(utils.NewBadRequestError("Invalid board ID"))
		return
	}

	// Get contributors using service
	contributors, users, err := h.boardService.ListBoardContributors(uint(boardID), userID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	// Build response
	contributorResponses := make([]responses.BoardContributorResponse, len(contributors))
	for i, contributor := range contributors {
		// Find matching user
		var user *models.User
		for _, u := range users {
			if u.ID == contributor.UserID {
				user = &u
				break
			}
		}

		if user != nil {
			// Create response
			contributorResponses[i] = responses.NewBoardContributorResponse(&contributor, user)
		}
	}

	c.JSON(http.StatusOK, responses.SuccessResponse(contributorResponses))
}

// UpdateBoardPreferences updates a user's preferences for a board (favorite/archived status)
func (h *BoardHandler) UpdateBoardPreferences(c *gin.Context) {
	// Get user ID from context
	userID := c.GetUint("userID")
	if userID == 0 {
		_ = c.Error(utils.NewUnauthorizedError("User not authenticated"))
		return
	}

	// Get board ID from URL
	boardID, err := strconv.ParseUint(c.Param("boardId"), 10, 32)
	if err != nil {
		_ = c.Error(utils.NewBadRequestError("Invalid board ID"))
		return
	}

	// Parse request
	var req requests.UpdateBoardPreferencesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(utils.NewValidationError(err.Error()))
		return
	}

	// Update preferences using service
	err = h.boardService.UpdateBoardPreferences(uint(boardID), userID, req.IsFavorite, req.IsArchived)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, responses.SuccessResponse(gin.H{"message": "Board preferences updated successfully"}))
}

// AddContributor adds a new contributor to a board
func (h *BoardHandler) AddContributor(c *gin.Context) {
	// Get user ID from context
	userID := c.GetUint("userID")
	if userID == 0 {
		_ = c.Error(utils.NewUnauthorizedError("User not authenticated"))
		return
	}

	// Get board ID from URL
	boardID, err := strconv.ParseUint(c.Param("boardId"), 10, 32)
	if err != nil {
		_ = c.Error(utils.NewBadRequestError("Invalid board ID"))
		return
	}

	// Parse request
	var req requests.AddContributorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(utils.NewValidationError(err.Error()))
		return
	}

	// Add contributor using service
	contributor, user, err := h.boardService.AddContributor(uint(boardID), userID, req.Email, req.Role)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, responses.SuccessResponse(responses.NewBoardContributorResponse(contributor, user)))
}

// UpdateContributor updates a contributor's role
func (h *BoardHandler) UpdateContributor(c *gin.Context) {
	// Get user ID from context
	userID := c.GetUint("userID")
	if userID == 0 {
		_ = c.Error(utils.NewUnauthorizedError("User not authenticated"))
		return
	}

	// Get board ID and user ID from URL
	boardID, err := strconv.ParseUint(c.Param("boardId"), 10, 32)
	if err != nil {
		_ = c.Error(utils.NewBadRequestError("Invalid board ID"))
		return
	}

	contributorID, err := strconv.ParseUint(c.Param("contributorId"), 10, 32)
	if err != nil {
		_ = c.Error(utils.NewBadRequestError("Invalid user ID"))
		return
	}

	// Parse request
	var req requests.UpdateContributorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(utils.NewValidationError(err.Error()))
		return
	}

	// Update contributor using service
	contributor, user, err := h.boardService.UpdateContributor(uint(boardID), userID, uint(contributorID), req.Role)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, responses.SuccessResponse(responses.NewBoardContributorResponse(contributor, user)))
}

// RemoveContributor removes a contributor from a board
func (h *BoardHandler) RemoveContributor(c *gin.Context) {
	// Get user ID from context
	userID := c.GetUint("userID")
	if userID == 0 {
		_ = c.Error(utils.NewUnauthorizedError("User not authenticated"))
		return
	}

	// Get board ID and user ID from URL
	boardID, err := strconv.ParseUint(c.Param("boardId"), 10, 32)
	if err != nil {
		_ = c.Error(utils.NewBadRequestError("Invalid board ID"))
		return
	}

	contributorID, err := strconv.ParseUint(c.Param("contributorId"), 10, 32)
	if err != nil {
		_ = c.Error(utils.NewBadRequestError("Invalid user ID"))
		return
	}

	// Remove contributor using service
	err = h.boardService.RemoveContributor(uint(boardID), userID, uint(contributorID))
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, responses.SuccessResponse(gin.H{"message": "Contributor removed successfully"}))
}
