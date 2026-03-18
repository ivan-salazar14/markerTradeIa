package database

import (
	hedgeModels "github.com/ivan-salazar14/markerTradeIa/internal/infrastructure/adapters/repository/hedgeAdapter/models"
	tradeModels "github.com/ivan-salazar14/markerTradeIa/internal/infrastructure/adapters/repository/tradeAdapter/models"
	"gorm.io/gorm"
)

type Migrator struct {
	DB *gorm.DB
}

func NewMigrator() *Migrator {
	connectionDB := GetInstance()
	return &Migrator{
		DB: connectionDB,
	}
}

func (m *Migrator) CreateStructures() error {
	return m.DB.AutoMigrate(
		&tradeModels.Trade{},
		&hedgeModels.WalletConnection{},
		&hedgeModels.HedgeState{},
		&hedgeModels.HedgeAction{},
		&hedgeModels.SyncEvent{},
	)
}
