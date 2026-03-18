package models

import "time"

type HedgeState struct {
	ID                 uint      `gorm:"primaryKey"`
	Asset              string    `gorm:"size:32;not null;index"`
	WalletAddress      string    `gorm:"size:128;not null"`
	HyperliquidAddress string    `gorm:"size:128;not null"`
	PoolExposure       float64   `gorm:"not null"`
	ShortExposure      float64   `gorm:"not null"`
	NetExposure        float64   `gorm:"not null"`
	Status             string    `gorm:"size:50;not null"`
	Message            string    `gorm:"type:text"`
	SafeMode           bool      `gorm:"not null"`
	DryRun             bool      `gorm:"not null"`
	LastSync           time.Time `gorm:"not null;index"`
	CreatedAt          time.Time
	UpdatedAt          time.Time
}
