package models

import "time"

type SyncEvent struct {
	ID          uint   `gorm:"primaryKey"`
	TriggerType string `gorm:"size:50;not null"`
	Asset       string `gorm:"size:32;not null;index"`
	Success     bool   `gorm:"not null"`
	Message     string `gorm:"type:text"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
