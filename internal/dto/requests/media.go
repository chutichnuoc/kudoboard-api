package requests

import "kudoboard-api/internal/models"

// AddMediaToPostRequest represents the request to add media to a post
type AddMediaToPostRequest struct {
	PostID     uint              `json:"post_id" binding:"required"`
	Type       models.MediaType  `json:"type" binding:"required,oneof=image gif video youtube"`
	SourceType models.SourceType `json:"source_type" binding:"required,oneof=upload youtube external"`
}

// AddYoutubeMediaRequest represents the request to add a YouTube video
type AddYoutubeMediaRequest struct {
	PostID     uint   `json:"post_id" binding:"required"`
	YoutubeURL string `json:"youtube_url" binding:"required,url"`
}
