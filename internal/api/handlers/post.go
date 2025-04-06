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

// PostHandler handles post-related requests
type PostHandler struct {
	postService  *services.PostService
	boardService *services.BoardService
	authService  *services.AuthService
	cfg          *config.Config
}

// NewPostHandler creates a new PostHandler
func NewPostHandler(postService *services.PostService, boardService *services.BoardService, authService *services.AuthService, cfg *config.Config) *PostHandler {
	return &PostHandler{
		postService:  postService,
		boardService: boardService,
		authService:  authService,
		cfg:          cfg,
	}
}

// CreatePost creates a new post
func (h *PostHandler) CreatePost(c *gin.Context) {
	// Get board ID from URL
	boardID, err := strconv.ParseUint(c.Param("boardId"), 10, 32)
	if err != nil {
		_ = c.Error(utils.NewBadRequestError("Invalid board ID"))
		return
	}

	// Check if user is authenticated
	userID := c.GetUint("userID")
	isAuthenticated := userID != 0

	// Parse request
	var req requests.CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(utils.NewValidationError(err.Error()))
		return
	}

	// Check if board exists
	board, err := h.boardService.GetBoardByID(uint(boardID))
	if err != nil {
		_ = c.Error(utils.NewNotFoundError("Board not found"))
		return
	}

	// For anonymous users, check if board allows anonymous posts
	if !isAuthenticated {
		if !board.AllowAnonymous {
			_ = c.Error(utils.NewForbiddenError("This board does not allow anonymous posts"))
			return
		}

		// Validate author name for anonymous posts
		if req.AuthorName == "" {
			_ = c.Error(utils.NewBadRequestError("Author name is required for anonymous posts"))
			return
		}
	}

	// Create post using service
	post, err := h.postService.CreatePost(uint(boardID), userID, req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	// Get author if authenticated
	var author *models.User
	if isAuthenticated {
		author, _ = h.authService.GetUserByID(userID)
	}

	// Return response
	c.JSON(http.StatusCreated, responses.SuccessResponse(
		responses.NewPostResponse(post, author, 0),
	))
}

// UpdatePost updates an existing post
func (h *PostHandler) UpdatePost(c *gin.Context) {
	// Get user ID from context
	userID := c.GetUint("userID")
	if userID == 0 {
		_ = c.Error(utils.NewUnauthorizedError("User not authenticated"))
		return
	}

	// Get post ID from URL
	postID, err := strconv.ParseUint(c.Param("postId"), 10, 32)
	if err != nil {
		_ = c.Error(utils.NewBadRequestError("Invalid post ID"))
		return
	}

	// Parse request
	var req requests.UpdatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(utils.NewValidationError(err.Error()))
		return
	}

	// Update post using service
	post, err := h.postService.UpdatePost(uint(postID), userID, req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	// Get author if not anonymous
	var author *models.User
	if post.AuthorID != nil {
		author, _ = h.authService.GetUserByID(*post.AuthorID)
	}

	// Count likes
	likesCount, _ := h.postService.CountPostLikes(post.ID)

	// Return response
	c.JSON(http.StatusOK, responses.SuccessResponse(
		responses.NewPostResponse(post, author, likesCount),
	))
}

// DeletePost deletes a post
func (h *PostHandler) DeletePost(c *gin.Context) {
	// Get user ID from context
	userID := c.GetUint("userID")
	if userID == 0 {
		_ = c.Error(utils.NewUnauthorizedError("User not authenticated"))
		return
	}

	// Get post ID from URL
	postID, err := strconv.ParseUint(c.Param("postId"), 10, 32)
	if err != nil {
		_ = c.Error(utils.NewBadRequestError("Invalid post ID format"))
		return
	}

	// Delete post using service
	err = h.postService.DeletePost(uint(postID), userID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, responses.SuccessResponse(gin.H{"message": "Post deleted successfully"}))
}

// LikePost adds a like to a post
func (h *PostHandler) LikePost(c *gin.Context) {
	// Get user ID from context
	userID := c.GetUint("userID")
	if userID == 0 {
		_ = c.Error(utils.NewUnauthorizedError("User not authenticated"))
		return
	}

	// Get post ID from URL
	postID, err := strconv.ParseUint(c.Param("postId"), 10, 32)
	if err != nil {
		_ = c.Error(utils.NewBadRequestError("Invalid post ID"))
		return
	}

	// Like post using service
	likesCount, err := h.postService.LikePost(uint(postID), userID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, responses.SuccessResponse(gin.H{
		"message":     "Post liked successfully",
		"likes_count": likesCount,
	}))
}

// UnlikePost removes a like from a post
func (h *PostHandler) UnlikePost(c *gin.Context) {
	// Get user ID from context
	userID := c.GetUint("userID")
	if userID == 0 {
		_ = c.Error(utils.NewUnauthorizedError("User not authenticated"))
		return
	}

	// Get post ID from URL
	postID, err := strconv.ParseUint(c.Param("postId"), 10, 32)
	if err != nil {
		_ = c.Error(utils.NewBadRequestError("Invalid post ID"))
		return
	}

	// Unlike post using service
	likesCount, err := h.postService.UnlikePost(uint(postID), userID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, responses.SuccessResponse(gin.H{
		"message":     "Post unliked successfully",
		"likes_count": likesCount,
	}))
}

// ReorderPosts updates the order of posts on a board
func (h *PostHandler) ReorderPosts(c *gin.Context) {
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
	var req requests.ReorderPostsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(utils.NewValidationError(err.Error()))
		return
	}

	// Reorder posts using service
	err = h.postService.ReorderPosts(uint(boardID), userID, req.PostPositions)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, responses.SuccessResponse(gin.H{"message": "Posts reordered successfully"}))
}
