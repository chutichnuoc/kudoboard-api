package models

import "gorm.io/gorm"

// Post represents a message on a kudoboard
type Post struct {
	gorm.Model
	BoardID         uint `gorm:"not null"`
	AuthorID        *uint
	AuthorName      string `gorm:"not null"`
	Content         string `gorm:"not null"`
	MediaPath       string
	MediaType       string
	MediaSource     string
	BackgroundColor string `gorm:"default:'#ffffff'"`
	TextColor       string `gorm:"default:'#000000'"`
	Position        int    `gorm:"default:0"`
}
