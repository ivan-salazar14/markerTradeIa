package database

import (
	"fmt"
	"os"
	"strings"
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
			db, err := NewConnectionDB()
			if err != nil {
				logrus.Fatalf("Error al conectar a la base de datos: %v", err)
			}
			instance = db
			logrus.Info("Conexion a la base de datos establecida.")
		})
	} else {
		logrus.Println("Reutilizando la conexion a la base de datos existente.")
	}
	return instance
}

func NewConnectionDB() (*gorm.DB, error) {
	dsn := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if dsn == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
}
