package models

import "gorm.io/gorm"

// Theme represents a predefined board theme
type Theme struct {
	gorm.Model
	Category           string `gorm:"not null"`
	Name               string `gorm:"not null"`
	IconUrl            string `gorm:"not null"`
	BackgroundImageURL string `gorm:"not null"`
}
