// Gestión de la configuración de la aplicación.
package config

import "time"

// Config contiene todos los parámetros de configuración de la aplicación.
type Config struct {
	KafkaBrokers   []string
	KafkaTopic     string
	PostgresDSN    string
	BinanceAPIKey  string
	BinanceSecret  string
	ProcessTimeout time.Duration
}

// Load carga la configuración desde variables de entorno, archivos, etc.
func Load() (*Config, error) {
	// Lógica para cargar la configuración
	// (Ej. usando viper, os.Getenv, etc.)
	return &Config{
		KafkaBrokers:   []string{"localhost:9092"},
		KafkaTopic:     "trading-signals",
		ProcessTimeout: 10 * time.Second,
	}, nil
}

// file: internal/core/domain
