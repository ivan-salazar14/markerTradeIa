package models

import "time"

type HedgeAction struct {
	ID         uint    `gorm:"primaryKey"`
	Asset      string  `gorm:"size:32;not null;index"`
	ActionType string  `gorm:"size:50;not null"`
	Size       float64 `gorm:"not null"`
	Status     string  `gorm:"size:50;not null"`
	Reason     string  `gorm:"type:text"`
	Executed   bool    `gorm:"not null"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
