package models

import (
	"gorm.io/gorm"
	"time"
)

// BirthDetail stores the immutable birth information for a user.
type BirthDetail struct {
	gorm.Model
	UserID    uint      `gorm:"uniqueIndex;not null"`
	DOB       time.Time `gorm:"not null"`
	TOB       string    `gorm:"type:varchar(8);not null"` // "HH:MM:SS"
	Latitude  float64   `gorm:"type:decimal(10,8);not null"`
	Longitude float64   `gorm:"type:decimal(11,8);not null"`
}

// User represents the main user profile.
type User struct {
	gorm.Model
	Email       string `gorm:"type:varchar(100);uniqueIndex;not null"`
	Gender      string `gorm:"type:varchar(10)"`
	Location    string `gorm:"type:varchar(100)"`
	PhotoURL    string
	BirthDetail BirthDetail `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}
