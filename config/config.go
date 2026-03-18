package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config contiene todos los parametros de configuracion de la aplicacion.
type Config struct {
	Port                   int
	DatabaseURL            string
	KafkaBrokers           []string
	KafkaTopic             string
	PostgresDSN            string
	BinanceAPIKey          string
	BinanceSecret          string
	HyperliquidAddress     string
	HyperliquidKey         string
	ProcessTimeout         time.Duration
	RevertBaseURL          string
	RevertNetworks         []string
	MonitoringInterval     time.Duration
	JWTSecret              string
	AccessExpiry           time.Duration
	RefreshExpiry          time.Duration
	ServiceAPIKeys         map[string]string
	EVMRPCURL              string
	UniswapPositionManager string
	DefaultHedgeAsset      string
	DefaultLPWalletAddress string
	SafeMode               bool
	DryRun                 bool
}

// Load carga la configuracion desde variables de entorno.
func Load() (*Config, error) {
	cfg := &Config{
		Port:                   getEnvAsInt("PORT", 8081),
		DatabaseURL:            getEnv("DATABASE_URL", ""),
		KafkaBrokers:           splitCSV(getEnv("KAFKA_BROKERS", "localhost:9092")),
		KafkaTopic:             getEnv("KAFKA_TOPIC", "trading-signals"),
		BinanceAPIKey:          getEnv("BINANCE_API_KEY", ""),
		BinanceSecret:          getEnv("BINANCE_SECRET", ""),
		HyperliquidAddress:     getEnv("HYPERLIQUID_ADDRESS", ""),
		HyperliquidKey:         getEnv("HYPERLIQUID_PRIVATE_KEY", ""),
		ProcessTimeout:         getEnvAsDuration("PROCESS_TIMEOUT", 10*time.Second),
		RevertBaseURL:          getEnv("REVERT_BASE_URL", "https://api.revert.finance/v1"),
		RevertNetworks:         splitCSV(getEnv("REVERT_NETWORKS", "mainnet,polygon,arbitrum,optimism")),
		MonitoringInterval:     getEnvAsDuration("MONITORING_INTERVAL", time.Minute),
		JWTSecret:              getEnv("JWT_SECRET", ""),
		AccessExpiry:           getEnvAsDuration("ACCESS_EXPIRY", 15*time.Minute),
		RefreshExpiry:          getEnvAsDuration("REFRESH_EXPIRY", 7*24*time.Hour),
		ServiceAPIKeys:         map[string]string{"monitoring-service": getEnv("MONITORING_SERVICE_API_KEY", "m2m-super-secret-key")},
		EVMRPCURL:              getEnv("EVM_RPC_URL", ""),
		UniswapPositionManager: getEnv("UNISWAP_POSITION_MANAGER", ""),
		DefaultHedgeAsset:      getEnv("DEFAULT_HEDGE_ASSET", "ETH"),
		DefaultLPWalletAddress: getEnv("DEFAULT_LP_WALLET_ADDRESS", ""),
		SafeMode:               getEnvAsBool("SAFE_MODE", true),
		DryRun:                 getEnvAsBool("DRY_RUN", true),
	}
	cfg.PostgresDSN = cfg.DatabaseURL

	var missing []string
	for key, value := range map[string]string{
		"DATABASE_URL":              cfg.DatabaseURL,
		"JWT_SECRET":                cfg.JWTSecret,
		"EVM_RPC_URL":               cfg.EVMRPCURL,
		"UNISWAP_POSITION_MANAGER":  cfg.UniswapPositionManager,
		"HYPERLIQUID_PRIVATE_KEY":   cfg.HyperliquidKey,
		"HYPERLIQUID_ADDRESS":       cfg.HyperliquidAddress,
		"DEFAULT_LP_WALLET_ADDRESS": cfg.DefaultLPWalletAddress,
	} {
		if strings.TrimSpace(value) == "" {
			missing = append(missing, key)
		}
	}

	if len(missing) > 0 {
		return nil, fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}

	return cfg, nil
}

func getEnv(key string, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return value
}

func getEnvAsDuration(key string, fallback time.Duration) time.Duration {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	value, err := time.ParseDuration(raw)
	if err != nil {
		return fallback
	}
	return value
}

func getEnvAsBool(key string, fallback bool) bool {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	value, err := strconv.ParseBool(raw)
	if err != nil {
		return fallback
	}
	return value
}

func splitCSV(raw string) []string {
	parts := strings.Split(raw, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		value := strings.TrimSpace(part)
		if value != "" {
			result = append(result, value)
		}
	}
	return result
}
