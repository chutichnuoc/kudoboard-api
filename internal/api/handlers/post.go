package handlers

import (
	"github.com/gin-gonic/gin"
	"kudoboard-api/internal/config"
	"kudoboard-api/internal/dto/requests"
	"kudoboard-api/internal/dto/responses"
	"kudoboard-api/internal/models"
	"kudoboard-api/internal/services"
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

// CreatePost creates a new post for an authenticated user
func (h *PostHandler) CreatePost(c *gin.Context) {
	// Get user ID from context
	userID := c.GetUint("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, responses.ErrorResponse("UNAUTHORIZED", "User not authenticated"))
		return
	}

	// Get board ID from URL
	boardID, err := strconv.ParseUint(c.Param("boardId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, responses.ErrorResponse("INVALID_ID", "Invalid board ID"))
		return
	}

	// Parse request
	var req requests.CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, responses.ErrorResponse("VALIDATION_ERROR", err.Error()))
		return
	}

	// Create post using service
	post, err := h.postService.CreatePost(uint(boardID), userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, responses.ErrorResponse("POST_CREATION_ERROR", err.Error()))
		return
	}

	// Get current user for response
	user, _ := h.authService.GetUserByID(userID)

	// Return response
	c.JSON(http.StatusCreated, responses.SuccessResponse(
		responses.NewPostResponse(post, user, nil, 0),
	))
}

// CreateAnonymousPost creates a new post for an anonymous user
func (h *PostHandler) CreateAnonymousPost(c *gin.Context) {
	// Get board ID from URL
	boardID, err := strconv.ParseUint(c.Param("boardId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, responses.ErrorResponse("INVALID_ID", "Invalid board ID"))
		return
	}

	// Parse request
	var req requests.CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, responses.ErrorResponse("VALIDATION_ERROR", err.Error()))
		return
	}

	// Check if board exists and allows anonymous posts
	board, err := h.boardService.GetBoardByID(uint(boardID))
	if err != nil {
		c.JSON(http.StatusNotFound, responses.ErrorResponse("BOARD_NOT_FOUND", "Board not found"))
		return
	}

	if !board.AllowAnonymous {
		c.JSON(http.StatusForbidden, responses.ErrorResponse("ANONYMOUS_NOT_ALLOWED", "This board does not allow anonymous posts"))
		return
	}

	// Force request to be anonymous
	req.IsAnonymous = true

	// Validate author name for anonymous posts
	if req.AuthorName == "" {
		c.JSON(http.StatusBadRequest, responses.ErrorResponse("VALIDATION_ERROR", "Author name is required for anonymous posts"))
		return
	}

	// Create post using service
	post, err := h.postService.CreatePost(uint(boardID), 0, req) // Pass 0 as userID for anonymous
	if err != nil {
		c.JSON(http.StatusInternalServerError, responses.ErrorResponse("POST_CREATION_ERROR", err.Error()))
		return
	}

	// Return response
	c.JSON(http.StatusCreated, responses.SuccessResponse(
		responses.NewPostResponse(post, nil, nil, 0),
	))
}

// UpdatePost updates an existing post
func (h *PostHandler) UpdatePost(c *gin.Context) {
	// Get user ID from context
	userID := c.GetUint("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, responses.ErrorResponse("UNAUTHORIZED", "User not authenticated"))
		return
	}

	// Get post ID from URL
	postID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, responses.ErrorResponse("INVALID_ID", "Invalid post ID"))
		return
	}

	// Parse request
	var req requests.UpdatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, responses.ErrorResponse("VALIDATION_ERROR", err.Error()))
		return
	}

	// Update post using service
	post, err := h.postService.UpdatePost(uint(postID), userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, responses.ErrorResponse("UPDATE_ERROR", err.Error()))
		return
	}

	// Get author if not anonymous
	var author *models.User
	if !post.IsAnonymous && post.AuthorID != nil {
		author, _ = h.authService.GetUserByID(*post.AuthorID)
	}

	// Get media for this post
	media, _ := h.postService.GetMediaForPost(post.ID)

	// Count likes
	likesCount, _ := h.postService.CountPostLikes(post.ID)

	// Return response
	c.JSON(http.StatusOK, responses.SuccessResponse(
		responses.NewPostResponse(post, author, media, likesCount),
	))
}

// DeletePost deletes a post
func (h *PostHandler) DeletePost(c *gin.Context) {
	// Get user ID from context
	userID := c.GetUint("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, responses.ErrorResponse("UNAUTHORIZED", "User not authenticated"))
		return
	}

	// Get post ID from URL
	postID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, responses.ErrorResponse("INVALID_ID", "Invalid post ID"))
		return
	}

	// Delete post using service
	err = h.postService.DeletePost(uint(postID), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, responses.ErrorResponse("DELETE_ERROR", err.Error()))
		return
	}

	c.JSON(http.StatusOK, responses.SuccessResponse(gin.H{"message": "Post deleted successfully"}))
}

// LikePost adds a like to a post
func (h *PostHandler) LikePost(c *gin.Context) {
	// Get user ID from context
	userID := c.GetUint("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, responses.ErrorResponse("UNAUTHORIZED", "User not authenticated"))
		return
	}

	// Get post ID from URL
	postID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, responses.ErrorResponse("INVALID_ID", "Invalid post ID"))
		return
	}

	// Like post using service
	likesCount, err := h.postService.LikePost(uint(postID), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, responses.ErrorResponse("LIKE_ERROR", err.Error()))
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
		c.JSON(http.StatusUnauthorized, responses.ErrorResponse("UNAUTHORIZED", "User not authenticated"))
		return
	}

	// Get post ID from URL
	postID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, responses.ErrorResponse("INVALID_ID", "Invalid post ID"))
		return
	}

	// Unlike post using service
	likesCount, err := h.postService.UnlikePost(uint(postID), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, responses.ErrorResponse("UNLIKE_ERROR", err.Error()))
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
		c.JSON(http.StatusUnauthorized, responses.ErrorResponse("UNAUTHORIZED", "User not authenticated"))
		return
	}

	// Get board ID from URL
	boardID, err := strconv.ParseUint(c.Param("boardId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, responses.ErrorResponse("INVALID_ID", "Invalid board ID"))
		return
	}

	// Parse request
	var req requests.ReorderPostsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, responses.ErrorResponse("VALIDATION_ERROR", err.Error()))
		return
	}

	// Reorder posts using service
	err = h.postService.ReorderPosts(uint(boardID), userID, req.PostOrders)
	if err != nil {
		c.JSON(http.StatusInternalServerError, responses.ErrorResponse("REORDER_ERROR", err.Error()))
		return
	}

	c.JSON(http.StatusOK, responses.SuccessResponse(gin.H{"message": "Posts reordered successfully"}))
}
