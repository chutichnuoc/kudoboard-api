package requests

// CreatePostRequest represents the request to create a new post
type CreatePostRequest struct {
	Content         string `json:"content" binding:"required"`
	AuthorName      string `json:"author_name"`
	BackgroundColor string `json:"background_color"`
	TextColor       string `json:"text_color"`
	MediaPath       string `json:"media_path,omitempty"`
	MediaType       string `json:"media_type,omitempty" binding:"required_with=MediaPath"`
	MediaSource     string `json:"media_source,omitempty" binding:"required_with=MediaPath,omitempty,oneof=internal external"`
}

// UpdatePostRequest represents the request to update a post
type UpdatePostRequest struct {
	Content         *string `json:"content"`
	AuthorName      *string `json:"author_name"`
	BackgroundColor *string `json:"background_color"`
	TextColor       *string `json:"text_color"`
	MediaPath       *string `json:"media_path"`
	MediaType       *string `json:"media_type" binding:"required_with=MediaPath"`
	MediaSource     *string `json:"media_source" binding:"required_with=MediaPath,omitempty,oneof=internal external"`
}

// ReorderPostsRequest represents the request to reorder posts on a board
type ReorderPostsRequest struct {
	PostPositions []PostPosition `json:"post_positions" binding:"required"`
}

// PostPosition represents the new order for a post
type PostPosition struct {
	ID       uint `json:"id" binding:"required"`
	Position int  `json:"position" binding:"required"`
}
