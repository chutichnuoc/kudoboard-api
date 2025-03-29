package responses

import (
	"kudoboard-api/internal/models"
	"time"
)

// PostResponse represents a post in API responses
type PostResponse struct {
	ID              uint          `json:"id"`
	BoardID         uint          `json:"board_id"`
	Author          *UserResponse `json:"author,omitempty"`
	AuthorName      string        `json:"author_name"`
	Content         string        `json:"content"`
	BackgroundColor string        `json:"background_color"`
	TextColor       string        `json:"text_color"`
	Position        int           `json:"position"`
	MediaPath       string        `json:"media_path"`
	MediaType       string        `json:"media_type"`
	MediaSource     string        `json:"media_source"`
	LikesCount      int           `json:"likes_count"`
	CreatedAt       time.Time     `json:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at"`
}

// NewPostResponse creates a new post response from a post model
func NewPostResponse(post *models.Post, author *models.User, likesCount int64) PostResponse {
	response := PostResponse{
		ID:              post.ID,
		BoardID:         post.BoardID,
		AuthorName:      post.AuthorName,
		Content:         post.Content,
		BackgroundColor: post.BackgroundColor,
		TextColor:       post.TextColor,
		Position:        post.Position,
		MediaPath:       post.MediaPath,
		MediaType:       post.MediaType,
		MediaSource:     post.MediaSource,
		LikesCount:      int(likesCount),
		CreatedAt:       post.CreatedAt,
		UpdatedAt:       post.UpdatedAt,
	}

	// Include author details if not anonymous
	if author != nil {
		authorResponse := NewUserResponse(author)
		response.Author = &authorResponse
	}

	return response
}
