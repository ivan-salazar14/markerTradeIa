package database

import (
	"sync"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	instance *gorm.DB
	once     sync.Once
)

func GetInstance() *gorm.DB {
	if instance == nil {
		once.Do(func() {
			instance, _ = NewConnectionDB()
			logrus.Info("Conexión a la base de datos establecida.")
		})
	} else {
		logrus.Printf("Reutilizando la conexión a la base de datos existente.\n")
	}
	return instance
}

func NewConnectionDB() (*gorm.DB, error) {
	// Aquí se debe inicializar la conexión a la base de datos PostgreSQL
	db, err := gorm.Open(postgres.Open("host=localhost user=postgres dbname=market_data password=postgres port=5432 sslmode=disable"), &gorm.Config{})
	if err != nil {
		logrus.Errorf("Error al conectar a la base de datos: %v", err)
		return nil, err
	}
	return db, nil
}
