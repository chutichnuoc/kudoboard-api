package requests

import (
	"time"

	"kudoboard-api/internal/models"
)

// CreateBoardRequest represents the request to create a new board
type CreateBoardRequest struct {
	Title              string                `json:"title" binding:"required"`
	Description        string                `json:"description"`
	BackgroundType     models.BackgroundType `json:"background_type"`
	BackgroundImageURL string                `json:"background_image_url"`
	BackgroundColor    string                `json:"background_color"`
	ThemeID            *uint                 `json:"theme_id"`
	IsPrivate          bool                  `json:"is_private"`
	AllowAnonymous     bool                  `json:"allow_anonymous"`
	ExpiresAt          *time.Time            `json:"expires_at"`
}

// UpdateBoardRequest represents the request to update a board
type UpdateBoardRequest struct {
	Title              *string                `json:"title"`
	Description        *string                `json:"description"`
	BackgroundType     *models.BackgroundType `json:"background_type"`
	BackgroundImageURL *string                `json:"background_image_url"`
	BackgroundColor    *string                `json:"background_color"`
	ThemeID            *uint                  `json:"theme_id"`
	IsPrivate          *bool                  `json:"is_private"`
	AllowAnonymous     *bool                  `json:"allow_anonymous"`
	ExpiresAt          *time.Time             `json:"expires_at"`
}

// BoardQuery represents query parameters for board listing
type BoardQuery struct {
	Page    int    `form:"page" binding:"min=1"`
	PerPage int    `form:"per_page" binding:"min=1,max=100"`
	Search  string `form:"search"`
	SortBy  string `form:"sort_by" binding:"omitempty,oneof=created_at title"`
	Order   string `form:"order" binding:"omitempty,oneof=asc desc"`
}

// AddContributorRequest represents a request to add a contributor to a board
type AddContributorRequest struct {
	Email string      `json:"email" binding:"required,email"`
	Role  models.Role `json:"role" binding:"required,oneof=viewer contributor admin"`
}

// UpdateContributorRequest represents a request to update a contributor's role
type UpdateContributorRequest struct {
	Role models.Role `json:"role" binding:"required,oneof=viewer contributor admin"`
}
