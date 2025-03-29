package models

import "time"

// PostLike represents a user liking a post
type PostLike struct {
	PostID    uint `gorm:"primaryKey"`
	UserID    uint `gorm:"primaryKey"`
	CreatedAt time.Time
}
