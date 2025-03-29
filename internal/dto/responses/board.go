package responses

import (
	"kudoboard-api/internal/models"
	"time"
)

// BoardResponse represents a board in API responses
type BoardResponse struct {
	ID                   uint           `json:"id"`
	Title                string         `json:"title"`
	ReceiverName         string         `json:"receiver_name"`
	Slug                 string         `json:"slug"`
	MaxPost              uint           `json:"max_post"`
	Creator              UserResponse   `json:"creator"`
	FontName             string         `json:"font_name" `
	FontSize             uint           `json:"font_size"`
	HeaderColor          string         `json:"header_color"`
	ShowHeaderColor      bool           `json:"show_header_color"`
	Theme                *ThemeResponse `json:"theme,omitempty"`
	Effect               string         `json:"effect"`
	EnableIntroAnimation bool           `json:"enable_intro_animation"`
	IsPrivate            bool           `json:"is_private"`
	IsLocked             bool           `json:"is_locked"`
	AllowAnonymous       bool           `json:"allow_anonymous"`
	CreatedAt            time.Time      `json:"created_at"`
	UpdatedAt            time.Time      `json:"updated_at"`
	PostCount            int            `json:"post_count"`
}

// BoardResponseWithRelation extends BoardResponse with user relationship info
type BoardResponseWithRelation struct {
	BoardResponse
	IsOwner    bool `json:"is_owner"`
	IsFavorite bool `json:"is_favorite"`
	IsArchived bool `json:"is_archived"`
}

// ThemeResponse represents a theme in API responses
type ThemeResponse struct {
	ID                 uint   `json:"id"`
	Category           string `json:"category"`
	Name               string `json:"name"`
	IconUrl            string `json:"icon_url"`
	BackgroundImageURL string `json:"background_image_url"`
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
		ID:                   board.ID,
		Title:                board.Title,
		ReceiverName:         board.ReceiverName,
		Slug:                 board.Slug,
		MaxPost:              board.MaxPost,
		FontName:             board.FontName,
		FontSize:             board.FontSize,
		HeaderColor:          board.HeaderColor,
		ShowHeaderColor:      board.ShowHeaderColor,
		Effect:               board.Effect,
		EnableIntroAnimation: board.EnableIntroAnimation,
		IsPrivate:            board.IsPrivate,
		IsLocked:             board.IsLocked,
		AllowAnonymous:       board.AllowAnonymous,
		CreatedAt:            board.CreatedAt,
		UpdatedAt:            board.UpdatedAt,
		PostCount:            postCount,
	}

	if creator != nil {
		response.Creator = NewUserResponse(creator)
	}

	return response
}

// NewBoardResponseWithRelation creates a new board response with relation info
func NewBoardResponseWithRelation(board *models.Board, creator *models.User, postCount int, isOwner, isFavorite, isArchived bool) BoardResponseWithRelation {
	return BoardResponseWithRelation{
		BoardResponse: NewBoardResponse(board, creator, postCount),
		IsOwner:       isOwner,
		IsFavorite:    isFavorite,
		IsArchived:    isArchived,
	}
}

// NewThemeResponse creates a new theme response from a theme model
func NewThemeResponse(theme *models.Theme) ThemeResponse {
	return ThemeResponse{
		ID:                 theme.ID,
		Category:           theme.Category,
		Name:               theme.Name,
		IconUrl:            theme.IconUrl,
		BackgroundImageURL: theme.BackgroundImageURL,
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
