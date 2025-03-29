package requests

// CreatePostRequest represents the request to create a new post
type CreatePostRequest struct {
	Content         string `json:"content" binding:"required"`
	AuthorName      string `json:"author_name" binding:"required_if=IsAnonymous true"`
	AuthorEmail     string `json:"author_email"`
	BackgroundColor string `json:"background_color"`
	TextColor       string `json:"text_color"`
	PositionX       int    `json:"position_x"`
	PositionY       int    `json:"position_y"`
	IsAnonymous     bool   `json:"is_anonymous"`
}

// UpdatePostRequest represents the request to update a post
type UpdatePostRequest struct {
	Content         *string `json:"content"`
	BackgroundColor *string `json:"background_color"`
	TextColor       *string `json:"text_color"`
	PositionX       *int    `json:"position_x"`
	PositionY       *int    `json:"position_y"`
}

// ReorderPostsRequest represents the request to reorder posts on a board
type ReorderPostsRequest struct {
	PostOrders []PostOrder `json:"post_orders" binding:"required"`
}

// PostOrder represents the new order for a post
type PostOrder struct {
	ID            uint `json:"id" binding:"required"`
	PositionOrder int  `json:"position_order" binding:"required"`
}

// PostQuery represents query parameters for post listing
type PostQuery struct {
	Page    int    `form:"page" binding:"min=1"`
	PerPage int    `form:"per_page" binding:"min=1,max=100"`
	SortBy  string `form:"sort_by" binding:"omitempty,oneof=created_at position_order"`
	Order   string `form:"order" binding:"omitempty,oneof=asc desc"`
}
