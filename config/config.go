// Gestión de la configuración de la aplicación.
package config

import "time"

// Config contiene todos los parámetros de configuración de la aplicación.
type Config struct {
	KafkaBrokers       []string
	KafkaTopic         string
	PostgresDSN        string
	BinanceAPIKey      string
	BinanceSecret      string
	HyperliquidAddress string
	HyperliquidKey     string
	ProcessTimeout     time.Duration
	RevertBaseURL      string
	RevertNetworks     []string
	MonitoringInterval time.Duration
	JWTSecret          string
	AccessExpiry       time.Duration
	RefreshExpiry      time.Duration
	ServiceAPIKeys     map[string]string
}

// Load carga la configuración desde variables de entorno, archivos, etc.
func Load() (*Config, error) {
	// Lógica para cargar la configuración
	// (Ej. usando viper, os.Getenv, etc.)
	return &Config{
		KafkaBrokers:       []string{"localhost:9092"},
		KafkaTopic:         "trading-signals",
		ProcessTimeout:     10 * time.Second,
		RevertBaseURL:      "https://api.revert.finance/v1",
		RevertNetworks:     []string{"mainnet", "polygon", "arbitrum", "optimism"},
		MonitoringInterval: 1 * time.Minute,
		JWTSecret:          "your-shhh-secret",
		AccessExpiry:       15 * time.Minute,
		RefreshExpiry:      24 * 7 * time.Hour,
		ServiceAPIKeys:     map[string]string{"monitoring-service": "m2m-super-secret-key"},
	}, nil
}

// file: internal/core/domain
