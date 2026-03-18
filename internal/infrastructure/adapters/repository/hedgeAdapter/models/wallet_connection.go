package models

import "time"

type WalletConnection struct {
	ID         uint   `gorm:"primaryKey"`
	WalletType string `gorm:"size:50;not null;index"`
	Address    string `gorm:"size:128;not null"`
	Status     string `gorm:"size:50;not null"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
