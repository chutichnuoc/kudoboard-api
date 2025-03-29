package requests

import (
	"kudoboard-api/internal/models"
)

// CreateBoardRequest represents the request to create a new board
type CreateBoardRequest struct {
	Title                string `json:"title" binding:"required"`
	ReceiverName         string `json:"receiver_name" binding:"required"`
	FontName             string `json:"font_name" binding:"required"`
	FontSize             uint   `json:"font_size"`
	HeaderColor          string `json:"header_color"`
	ThemeID              *uint  `json:"theme_id"`
	Effect               string `json:"effect"`
	EnableIntroAnimation bool   `json:"enable_intro_animation"`
	IsPrivate            bool   `json:"is_private"`
	AllowAnonymous       bool   `json:"allow_anonymous"`
}

// UpdateBoardRequest represents the request to update a board
type UpdateBoardRequest struct {
	Title                *string `json:"title"`
	ReceiverName         *string `json:"receiver_name" `
	FontName             *string `json:"font_name"`
	FontSize             *uint   `json:"font_size"`
	HeaderColor          *string `json:"header_color"`
	ShowHeaderColor      *bool   `json:"show_header_color"`
	ThemeID              *uint   `json:"theme_id"`
	Effect               *string `json:"effect"`
	EnableIntroAnimation *bool   `json:"enable_intro_animation"`
	IsPrivate            *bool   `json:"is_private"`
	AllowAnonymous       *bool   `json:"allow_anonymous"`
}

// LockBoardRequest represents a request to lock or unlock a board
type LockBoardRequest struct {
	IsLocked bool `json:"is_locked"`
}

// UpdateBoardPreferencesRequest represents a request to update a user's board preferences
type UpdateBoardPreferencesRequest struct {
	IsFavorite *bool `json:"is_favorite,omitempty"`
	IsArchived *bool `json:"is_archived,omitempty"`
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
