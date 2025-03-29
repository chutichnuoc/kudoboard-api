package responses

import (
	"kudoboard-api/internal/models"
	"time"
)

// BoardResponse represents a board in API responses
type BoardResponse struct {
	ID                 uint                  `json:"id"`
	Title              string                `json:"title"`
	Description        string                `json:"description"`
	Slug               string                `json:"slug"`
	Creator            UserResponse          `json:"creator"`
	BackgroundType     models.BackgroundType `json:"background_type"`
	BackgroundImageURL string                `json:"background_image_url"`
	BackgroundColor    string                `json:"background_color"`
	ThemeID            *uint                 `json:"theme_id,omitempty"`
	Theme              *ThemeResponse        `json:"theme,omitempty"`
	IsPrivate          bool                  `json:"is_private"`
	AllowAnonymous     bool                  `json:"allow_anonymous"`
	ExpiresAt          *time.Time            `json:"expires_at"`
	CreatedAt          time.Time             `json:"created_at"`
	UpdatedAt          time.Time             `json:"updated_at"`
	PostCount          int                   `json:"post_count"`
}

// ThemeResponse represents a theme in API responses
type ThemeResponse struct {
	ID                 uint   `json:"id"`
	Name               string `json:"name"`
	Description        string `json:"description"`
	BackgroundColor    string `json:"background_color"`
	BackgroundImageURL string `json:"background_image_url"`
	AdditionalStyles   string `json:"additional_styles"`
	IsDefault          bool   `json:"is_default"`
}

// BoardContributorResponse represents a board contributor in API responses
type BoardContributorResponse struct {
	BoardID   uint         `json:"board_id"`
	User      UserResponse `json:"user"`
	Role      string       `json:"role"`
	CreatedAt time.Time    `json:"created_at"`
}

// NewBoardResponse creates a new board response from a board model
func NewBoardResponse(board *models.Board, creator *models.User, postCount int) BoardResponse {
	response := BoardResponse{
		ID:                 board.ID,
		Title:              board.Title,
		Description:        board.Description,
		Slug:               board.Slug,
		BackgroundType:     board.BackgroundType,
		BackgroundImageURL: board.BackgroundImageURL,
		BackgroundColor:    board.BackgroundColor,
		ThemeID:            board.ThemeID,
		IsPrivate:          board.IsPrivate,
		AllowAnonymous:     board.AllowAnonymous,
		ExpiresAt:          board.ExpiresAt,
		CreatedAt:          board.CreatedAt,
		UpdatedAt:          board.UpdatedAt,
		PostCount:          postCount,
	}

	if creator != nil {
		response.Creator = NewUserResponse(creator)
	}

	return response
}

// NewThemeResponse creates a new theme response from a theme model
func NewThemeResponse(theme *models.Theme) ThemeResponse {
	return ThemeResponse{
		ID:                 theme.ID,
		Name:               theme.Name,
		Description:        theme.Description,
		BackgroundColor:    theme.BackgroundColor,
		BackgroundImageURL: theme.BackgroundImageURL,
		AdditionalStyles:   theme.AdditionalStyles,
		IsDefault:          theme.IsDefault,
	}
}

// NewBoardContributorResponse creates a new board contributor response
func NewBoardContributorResponse(contributor *models.BoardContributor, user *models.User) BoardContributorResponse {
	return BoardContributorResponse{
		BoardID:   contributor.BoardID,
		User:      NewUserResponse(user),
		Role:      string(contributor.Role),
		CreatedAt: contributor.CreatedAt,
	}
}
