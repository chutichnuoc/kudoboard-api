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

// CreatePost creates a new post
func (h *PostHandler) CreatePost(c *gin.Context) {
	// Get board ID from URL
	boardID, err := strconv.ParseUint(c.Param("boardId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, responses.ErrorResponse("INVALID_ID", "Invalid board ID"))
		return
	}

	// Check if user is authenticated
	userID := c.GetUint("userID")
	isAuthenticated := userID != 0

	// Parse request
	var req requests.CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, responses.ErrorResponse("VALIDATION_ERROR", err.Error()))
		return
	}

	// Check if board exists
	board, err := h.boardService.GetBoardByID(uint(boardID))
	if err != nil {
		c.JSON(http.StatusNotFound, responses.ErrorResponse("BOARD_NOT_FOUND", "Board not found"))
		return
	}

	// For anonymous users, check if board allows anonymous posts
	if !isAuthenticated {
		if !board.AllowAnonymous {
			c.JSON(http.StatusForbidden, responses.ErrorResponse("ANONYMOUS_NOT_ALLOWED", "This board does not allow anonymous posts"))
			return
		}

		// Validate author name for anonymous posts
		if req.AuthorName == "" {
			c.JSON(http.StatusBadRequest, responses.ErrorResponse("VALIDATION_ERROR", "Author name is required for anonymous posts"))
			return
		}
	}

	// Create post using service
	post, err := h.postService.CreatePost(uint(boardID), userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, responses.ErrorResponse("POST_CREATION_ERROR", err.Error()))
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
		c.JSON(http.StatusUnauthorized, responses.ErrorResponse("UNAUTHORIZED", "User not authenticated"))
		return
	}

	// Get post ID from URL
	postID, err := strconv.ParseUint(c.Param("postId"), 10, 32)
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
		c.JSON(http.StatusUnauthorized, responses.ErrorResponse("UNAUTHORIZED", "User not authenticated"))
		return
	}

	// Get post ID from URL
	postID, err := strconv.ParseUint(c.Param("postId"), 10, 32)
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
	postID, err := strconv.ParseUint(c.Param("postId"), 10, 32)
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
	postID, err := strconv.ParseUint(c.Param("postId"), 10, 32)
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
	err = h.postService.ReorderPosts(uint(boardID), userID, req.PostPositions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, responses.ErrorResponse("REORDER_ERROR", err.Error()))
		return
	}

	c.JSON(http.StatusOK, responses.SuccessResponse(gin.H{"message": "Posts reordered successfully"}))
}
