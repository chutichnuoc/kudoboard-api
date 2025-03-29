package models

import (
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// User represents a user in the system
type User struct {
	gorm.Model
	Name           string `gorm:"not null"`
	Email          string `gorm:"uniqueIndex;not null"`
	Password       string `gorm:"not null"`
	ProfilePicture string
	IsVerified     bool    `gorm:"default:false"`
	GoogleID       *string `gorm:"uniqueIndex;default:null"`
	FacebookID     *string `gorm:"uniqueIndex;default:null"`
	AuthProvider   string  `gorm:"default:'local'"`
}

// BeforeSave hook is called before saving a User to hash the password
func (u *User) BeforeSave(tx *gorm.DB) error {
	if u.Password != "" {
		newRecord := tx.Statement.RowsAffected == 0
		passwordChanged := tx.Statement.Changed("Password")

		if newRecord || passwordChanged {
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
			if err != nil {
				return err
			}
			u.Password = string(hashedPassword)
		}
	}
	return nil
}

// CheckPassword verifies if the provided password matches the stored hash
func (u *User) CheckPassword(password string) error {
	return bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
}
