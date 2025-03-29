package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Board represents a kudoboard where users can post messages
type Board struct {
	gorm.Model
	Title                string `gorm:"not null"`
	ReceiverName         string `gorm:"not null"`
	Slug                 string `gorm:"uniqueIndex;not null"`
	MaxPost              uint   `gorm:"default:10"`
	CreatorID            uint   `gorm:"not null"`
	FontName             string `gorm:"not null"`
	FontSize             uint   `gorm:"not null;default:14"`
	HeaderColor          string `gorm:"default:'#ffffff'"`
	ShowHeaderColor      bool   `gorm:"default:true"`
	ThemeID              *uint
	Effect               string `gorm:"type:jsonb"`
	EnableIntroAnimation bool   `gorm:"default:false"`
	IsPrivate            bool   `gorm:"default:false"`
	IsLocked             bool   `gorm:"default:false"`
	AllowAnonymous       bool   `gorm:"default:true"`
}

// BeforeCreate hook to generate a unique slug for new boards
func (b *Board) BeforeCreate(tx *gorm.DB) error {
	if b.Slug == "" {
		// Generate a URL-friendly slug from a UUID
		b.Slug = uuid.New().String()[0:18]
	}
	return nil
}
