package responses

import (
	"kudoboard-api/internal/models"
	"time"
)

// PostResponse represents a post in API responses
type PostResponse struct {
	ID              uint            `json:"id"`
	BoardID         uint            `json:"board_id"`
	Author          *UserResponse   `json:"author,omitempty"`
	AuthorName      string          `json:"author_name"`
	AuthorEmail     string          `json:"author_email,omitempty"`
	Content         string          `json:"content"`
	BackgroundColor string          `json:"background_color"`
	TextColor       string          `json:"text_color"`
	PositionX       int             `json:"position_x"`
	PositionY       int             `json:"position_y"`
	PositionOrder   int             `json:"position_order"`
	IsAnonymous     bool            `json:"is_anonymous"`
	Media           []MediaResponse `json:"media,omitempty"`
	LikesCount      int             `json:"likes_count"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

// NewPostResponse creates a new post response from a post model
func NewPostResponse(post *models.Post, author *models.User, media []models.Media, likesCount int64) PostResponse {
	response := PostResponse{
		ID:              post.ID,
		BoardID:         post.BoardID,
		AuthorName:      post.AuthorName,
		Content:         post.Content,
		BackgroundColor: post.BackgroundColor,
		TextColor:       post.TextColor,
		PositionX:       post.PositionX,
		PositionY:       post.PositionY,
		PositionOrder:   post.PositionOrder,
		IsAnonymous:     post.IsAnonymous,
		LikesCount:      int(likesCount),
		CreatedAt:       post.CreatedAt,
		UpdatedAt:       post.UpdatedAt,
	}

	// Include author email only for authenticated users
	if !post.IsAnonymous {
		response.AuthorEmail = post.AuthorEmail
	}

	// Include author details if not anonymous
	if !post.IsAnonymous && author != nil {
		authorResponse := NewUserResponse(author)
		response.Author = &authorResponse
	}

	// Include media if available
	if len(media) > 0 {
		response.Media = make([]MediaResponse, len(media))
		for i, m := range media {
			response.Media[i] = NewMediaResponse(&m)
		}
	}

	return response
}
