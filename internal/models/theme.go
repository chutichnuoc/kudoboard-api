package models

import "gorm.io/gorm"

// Theme represents a predefined board theme
type Theme struct {
	gorm.Model
	Name               string `gorm:"not null"`
	Description        string
	BackgroundColor    string `gorm:"default:'#ffffff'"`
	BackgroundImageURL string
	AdditionalStyles   string `gorm:"type:json"` // JSON string with additional style settings
	IsDefault          bool   `gorm:"default:false"`
}
