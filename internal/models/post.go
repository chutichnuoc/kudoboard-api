package models

import "gorm.io/gorm"

// Post represents a message on a kudoboard
type Post struct {
	gorm.Model
	BoardID         uint `gorm:"not null"`
	AuthorID        *uint
	AuthorName      string `gorm:"not null"`
	AuthorEmail     string
	Content         string `gorm:"not null"`
	BackgroundColor string `gorm:"default:'#ffffff'"`
	TextColor       string `gorm:"default:'#000000'"`
	PositionX       int    `gorm:"default:0"`
	PositionY       int    `gorm:"default:0"`
	PositionOrder   int    `gorm:"default:0"`
	IsAnonymous     bool   `gorm:"default:false"`
}

// AfterFind hook to count likes
func (p *Post) AfterFind(tx *gorm.DB) error {
	// Count the number of likes for this post
	var count int64
	tx.Model(&PostLike{}).Where("post_id = ?", p.ID).Count(&count)

	// We can't store this directly in the struct since it's not a DB field,
	// but services can use this method to get the count
	return nil
}

// CountLikes returns the number of likes for this post
func (p *Post) CountLikes(db *gorm.DB) int64 {
	var count int64
	db.Model(&PostLike{}).Where("post_id = ?", p.ID).Count(&count)
	return count
}
