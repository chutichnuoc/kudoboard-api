package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BackgroundType defines the type of background for a board
type BackgroundType string

const (
	BackgroundTypeImage BackgroundType = "image"
	BackgroundTypeColor BackgroundType = "color"
	BackgroundTypeTheme BackgroundType = "theme"
)

// Board represents a kudoboard where users can post messages
type Board struct {
	gorm.Model
	Title              string `gorm:"not null"`
	Description        string
	Slug               string         `gorm:"uniqueIndex;not null"`
	CreatorID          uint           `gorm:"not null"`
	BackgroundType     BackgroundType `gorm:"type:varchar(10);default:'color'"`
	BackgroundImageURL string
	BackgroundColor    string `gorm:"default:'#ffffff'"`
	ThemeID            *uint
	IsPrivate          bool `gorm:"default:false"`
	AllowAnonymous     bool `gorm:"default:true"`
	ExpiresAt          *time.Time
}

// BeforeCreate hook to generate a unique slug for new boards
func (b *Board) BeforeCreate(tx *gorm.DB) error {
	if b.Slug == "" {
		// Generate a URL-friendly slug from a UUID
		b.Slug = uuid.New().String()[0:10]
	}
	return nil
}
