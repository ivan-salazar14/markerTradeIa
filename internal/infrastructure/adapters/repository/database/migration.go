package database

import (
	"github.com/ivan-salazar14/markerTradeIa/internal/infrastructure/adapters/repository/tradeAdapter/models"
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
	// Aquí se pueden agregar más modelos según sea necesario

	return m.DB.AutoMigrate(&models.Trade{})
}
