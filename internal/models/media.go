package models

import "gorm.io/gorm"

// MediaType defines the type of media
type MediaType string

const (
	MediaTypeImage   MediaType = "image"
	MediaTypeGif     MediaType = "gif"
	MediaTypeVideo   MediaType = "video"
	MediaTypeYoutube MediaType = "youtube"
)

// SourceType defines where the media comes from
type SourceType string

const (
	SourceTypeUpload   SourceType = "upload"
	SourceTypeYoutube  SourceType = "youtube"
	SourceTypeExternal SourceType = "external"
)

// Media represents a media attachment (image, gif, video)
type Media struct {
	gorm.Model
	PostID       uint       `gorm:"not null"`
	Type         MediaType  `gorm:"type:varchar(20);not null"`
	SourceType   SourceType `gorm:"type:varchar(20);not null"`
	SourceURL    string     `gorm:"not null"`
	ExternalID   string     // For YouTube video IDs, etc.
	ThumbnailURL string
}
