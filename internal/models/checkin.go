package models

import "time"

type Checkin struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	UserID          uint      `gorm:"not null;index" json:"userId"`
	MoodScore       uint      `gorm:"not null;check:mood_score >= 1 AND mood_score <= 10" json:"moodScore"`
	SleepHours      float32   `gorm:"not null" json:"sleepHours"`
	EnergyLevel     uint      `gorm:"not null;check:energy_level >= 1 AND energy_level <= 10" json:"energyLevel"`
	MedicationTaken bool      `gorm:"not null;default:false" json:"medicationTaken"`
	Note            string    `gorm:"size:1000" json:"note"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
	User            User      `gorm:"foreignKey:UserID"`
}
