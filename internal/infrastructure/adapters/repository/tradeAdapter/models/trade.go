package models

import (
	"time"
)

type Trade struct {
	ExecutionID string    `gorm:"primaryKey"`
	SignalID    string    `gorm:"not null"`
	Status      string    `gorm:"not null"`
	ExecutedAt  time.Time `gorm:"default:current_timestamp"`
	Details     string    `gorm:"type:text"`
	UserID      string    `gorm:"not null"`
}
