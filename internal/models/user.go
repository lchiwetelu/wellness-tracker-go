package models

import "time"

type User struct {
	ID           uint   `gorm:"primaryKey" json:"id"`
	Email        string `gorm:"uniqueIndex;not null"`
	Name         string
	Picture      string
	GoogleID     string    `gorm:"uniqueIndex"`
	AuthProvider string    `gorm:"not null"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}
