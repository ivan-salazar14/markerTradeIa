package postgres

import (
	"sync"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type connectionDB struct {
	*gorm.DB
}

var (
	instance *connectionDB
	once     sync.Once
)

func getInstance() *connectionDB {
	if instance == nil {
		once.Do(func() {
			instance, _ = NewConnectionDB()
		})
	} else {
		logrus.Printf("Reutilizando la conexión a la base de datos existente.\n")
	}
	return instance
}

func NewConnectionDB() (*connectionDB, error) {
	// Aquí se debe inicializar la conexión a la base de datos PostgreSQL
	db, err := gorm.Open(postgres.Open("host=localhost user=postgres dbname=postgres port=5432 sslmode=disable"), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return &connectionDB{DB: db}, nil
}
