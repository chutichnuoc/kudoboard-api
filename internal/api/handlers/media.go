package handlers

import (
	"github.com/gin-gonic/gin"
	"kudoboard-api/internal/config"
	"kudoboard-api/internal/dto/requests"
	"kudoboard-api/internal/dto/responses"
	"kudoboard-api/internal/services"
	"net/http"
	"strconv"
)

// MediaHandler handles media-related requests
type MediaHandler struct {
	mediaService *services.MediaService
	boardService *services.BoardService
	postService  *services.PostService
	cfg          *config.Config
}

// NewMediaHandler creates a new MediaHandler
func NewMediaHandler(mediaService *services.MediaService, boardService *services.BoardService, postService *services.PostService, cfg *config.Config) *MediaHandler {
	return &MediaHandler{
		mediaService: mediaService,
		boardService: boardService,
		postService:  postService,
		cfg:          cfg,
	}
}

// UploadMedia handles file uploads for posts
func (h *MediaHandler) UploadMedia(c *gin.Context) {
	// Get user ID from context
	userID := c.GetUint("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, responses.ErrorResponse("UNAUTHORIZED", "User not authenticated"))
		return
	}

	// Get the post ID from form
	postIDStr := c.PostForm("post_id")
	if postIDStr == "" {
		c.JSON(http.StatusBadRequest, responses.ErrorResponse("MISSING_POST_ID", "Post ID is required"))
		return
	}

	postID, err := strconv.ParseUint(postIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, responses.ErrorResponse("INVALID_POST_ID", "Invalid post ID"))
		return
	}

	// Get file from form
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, responses.ErrorResponse("INVALID_FILE", "Failed to read file"))
		return
	}

	// Upload media using service
	media, err := h.mediaService.UploadMedia(file, uint(postID), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, responses.ErrorResponse("UPLOAD_ERROR", err.Error()))
		return
	}

	c.JSON(http.StatusCreated, responses.SuccessResponse(responses.NewMediaResponse(media)))
}

// UploadAnonymousMedia handles file uploads for anonymous posts
func (h *MediaHandler) UploadAnonymousMedia(c *gin.Context) {
	// Get board ID from URL
	boardID, err := strconv.ParseUint(c.Param("boardId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, responses.ErrorResponse("INVALID_ID", "Invalid board ID"))
		return
	}

	// Get the post ID from form
	postIDStr := c.PostForm("post_id")
	if postIDStr == "" {
		c.JSON(http.StatusBadRequest, responses.ErrorResponse("MISSING_POST_ID", "Post ID is required"))
		return
	}

	postID, err := strconv.ParseUint(postIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, responses.ErrorResponse("INVALID_POST_ID", "Invalid post ID"))
		return
	}

	// Check if board allows anonymous posts
	board, err := h.boardService.GetBoardByID(uint(boardID))
	if err != nil {
		c.JSON(http.StatusNotFound, responses.ErrorResponse("BOARD_NOT_FOUND", "Board not found"))
		return
	}

	if !board.AllowAnonymous {
		c.JSON(http.StatusForbidden, responses.ErrorResponse("ANONYMOUS_NOT_ALLOWED", "This board does not allow anonymous posts"))
		return
	}

	// Get file from form
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, responses.ErrorResponse("INVALID_FILE", "Failed to read file"))
		return
	}

	// Verify post belongs to the specified board and is anonymous
	post, err := h.postService.GetPostByID(uint(postID))
	if err != nil {
		c.JSON(http.StatusNotFound, responses.ErrorResponse("POST_NOT_FOUND", "Post not found"))
		return
	}

	if post.BoardID != uint(boardID) || !post.IsAnonymous {
		c.JSON(http.StatusForbidden, responses.ErrorResponse("INVALID_POST", "Invalid post for anonymous media upload"))
		return
	}

	// Upload media using service (pass 0 as userID for anonymous)
	media, err := h.mediaService.UploadMedia(file, uint(postID), 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, responses.ErrorResponse("UPLOAD_ERROR", err.Error()))
		return
	}

	c.JSON(http.StatusCreated, responses.SuccessResponse(responses.NewMediaResponse(media)))
}

// AddYoutube adds a YouTube video to a post
func (h *MediaHandler) AddYoutube(c *gin.Context) {
	// Get user ID from context
	userID := c.GetUint("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, responses.ErrorResponse("UNAUTHORIZED", "User not authenticated"))
		return
	}

	// Parse request
	var req requests.AddYoutubeMediaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, responses.ErrorResponse("VALIDATION_ERROR", err.Error()))
		return
	}

	// Add YouTube video using service
	media, err := h.mediaService.AddYoutubeVideo(req.PostID, userID, req.YoutubeURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, responses.ErrorResponse("YOUTUBE_ERROR", err.Error()))
		return
	}

	c.JSON(http.StatusCreated, responses.SuccessResponse(responses.NewMediaResponse(media)))
}

// AddYoutubeAnonymous adds a YouTube video to an anonymous post
func (h *MediaHandler) AddYoutubeAnonymous(c *gin.Context) {
	// Get board ID from URL
	boardID, err := strconv.ParseUint(c.Param("boardId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, responses.ErrorResponse("INVALID_ID", "Invalid board ID"))
		return
	}

	// Parse request
	var req requests.AddYoutubeMediaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, responses.ErrorResponse("VALIDATION_ERROR", err.Error()))
		return
	}

	// Check if board allows anonymous posts
	board, err := h.boardService.GetBoardByID(uint(boardID))
	if err != nil {
		c.JSON(http.StatusNotFound, responses.ErrorResponse("BOARD_NOT_FOUND", "Board not found"))
		return
	}

	if !board.AllowAnonymous {
		c.JSON(http.StatusForbidden, responses.ErrorResponse("ANONYMOUS_NOT_ALLOWED", "This board does not allow anonymous posts"))
		return
	}

	// Verify post belongs to the specified board and is anonymous
	post, err := h.postService.GetPostByID(req.PostID)
	if err != nil {
		c.JSON(http.StatusNotFound, responses.ErrorResponse("POST_NOT_FOUND", "Post not found"))
		return
	}

	if post.BoardID != uint(boardID) || !post.IsAnonymous {
		c.JSON(http.StatusForbidden, responses.ErrorResponse("INVALID_POST", "Invalid post for anonymous media upload"))
		return
	}

	// Add YouTube video using service (pass 0 as userID for anonymous)
	media, err := h.mediaService.AddYoutubeVideo(req.PostID, 0, req.YoutubeURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, responses.ErrorResponse("YOUTUBE_ERROR", err.Error()))
		return
	}

	c.JSON(http.StatusCreated, responses.SuccessResponse(responses.NewMediaResponse(media)))
}

// DeleteMedia removes a media item
func (h *MediaHandler) DeleteMedia(c *gin.Context) {
	// Get user ID from context
	userID := c.GetUint("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, responses.ErrorResponse("UNAUTHORIZED", "User not authenticated"))
		return
	}

	// Get media ID from URL
	mediaID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, responses.ErrorResponse("INVALID_ID", "Invalid media ID"))
		return
	}

	// Delete media using service
	err = h.mediaService.DeleteMedia(uint(mediaID), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, responses.ErrorResponse("DELETE_ERROR", err.Error()))
		return
	}

	c.JSON(http.StatusOK, responses.SuccessResponse(gin.H{"message": "Media deleted successfully"}))
}
