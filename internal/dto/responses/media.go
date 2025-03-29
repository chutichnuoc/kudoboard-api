package responses

import (
	"kudoboard-api/internal/models"
	"time"
)

// MediaResponse represents a media attachment in API responses
type MediaResponse struct {
	ID           uint              `json:"id"`
	PostID       uint              `json:"post_id"`
	Type         models.MediaType  `json:"type"`
	SourceType   models.SourceType `json:"source_type"`
	SourceURL    string            `json:"source_url"`
	ExternalID   string            `json:"external_id,omitempty"`
	ThumbnailURL string            `json:"thumbnail_url,omitempty"`
	CreatedAt    time.Time         `json:"created_at"`
}

// NewMediaResponse creates a new media response from a media model
func NewMediaResponse(media *models.Media) MediaResponse {
	return MediaResponse{
		ID:           media.ID,
		PostID:       media.PostID,
		Type:         media.Type,
		SourceType:   media.SourceType,
		SourceURL:    media.SourceURL,
		ExternalID:   media.ExternalID,
		ThumbnailURL: media.ThumbnailURL,
		CreatedAt:    media.CreatedAt,
	}
}
