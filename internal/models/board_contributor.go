package models

import "time"

// Role defines access levels for board contributors
type Role string

const (
	RoleViewer      Role = "viewer"
	RoleContributor Role = "contributor"
	RoleAdmin       Role = "admin"
)

// BoardContributor represents a user who has access to a board
type BoardContributor struct {
	BoardID    uint `gorm:"primaryKey"`
	UserID     uint `gorm:"primaryKey"`
	Role       Role `gorm:"type:varchar(20);default:'viewer'"`
	IsFavorite bool `gorm:"default:false"`
	IsArchived bool `gorm:"default:false"`
	CreatedAt  time.Time
}
