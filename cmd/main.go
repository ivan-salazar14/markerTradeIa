package main

import (
	"context"
	"log"

	"github.com/ivan-salazar14/markerTradeIa/config"
	"github.com/ivan-salazar14/markerTradeIa/internal/application/services/auth"
	"github.com/ivan-salazar14/markerTradeIa/internal/application/services/monitoring"
	"github.com/ivan-salazar14/markerTradeIa/internal/application/usecases/hedge"
	"github.com/ivan-salazar14/markerTradeIa/internal/application/usecases/hedge/strategies"
	"github.com/ivan-salazar14/markerTradeIa/internal/domain"
	"github.com/ivan-salazar14/markerTradeIa/internal/infrastructure/adapters/api"
	"github.com/ivan-salazar14/markerTradeIa/internal/infrastructure/adapters/api/controllers"
	"github.com/ivan-salazar14/markerTradeIa/internal/infrastructure/adapters/dex/uniswap"
	"github.com/ivan-salazar14/markerTradeIa/internal/infrastructure/adapters/monitoring/revert"
	"github.com/ivan-salazar14/markerTradeIa/internal/infrastructure/adapters/perps"
	"github.com/ivan-salazar14/markerTradeIa/internal/infrastructure/adapters/repository/database"
	hedgeAdapter "github.com/ivan-salazar14/markerTradeIa/internal/infrastructure/adapters/repository/hedgeAdapter"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Error cargando configuracion: %v", err)
	}

	log.Println("Iniciando servicio hedge delta-neutral...")

	migrator := database.NewMigrator()
	if err := migrator.CreateStructures(); err != nil {
		log.Fatalf("Error creando estructuras de base de datos: %v", err)
	}

	authSvc := auth.NewAuthService(domain.AuthConfig{
		JWTSecret:      cfg.JWTSecret,
		AccessExpiry:   cfg.AccessExpiry,
		RefreshExpiry:  cfg.RefreshExpiry,
		ServiceAPIKeys: cfg.ServiceAPIKeys,
	})

	revertAdapter := revert.NewRevertAdapter(cfg.RevertBaseURL)
	poolMonitoringService := monitoring.NewMonitoringService(
		revertAdapter,
		cfg.RevertNetworks,
		cfg.MonitoringInterval,
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	walletAdapter, err := uniswap.NewUniswapV3WalletAdapter(
		cfg.EVMRPCURL,
		cfg.UniswapPositionManager,
	)
	if err != nil {
		log.Fatalf("Error inicializando Uniswap adapter: %v", err)
	}

	hyperliquidAdapter := perps.NewHyperliquidAdapter()
	if err := hyperliquidAdapter.Connect(ctx, cfg.HyperliquidKey); err != nil {
		log.Fatalf("Error inicializando Hyperliquid adapter: %v", err)
	}

	hedgeRepository := hedgeAdapter.NewHedgeRepository()
	deltaStrategy := strategies.NewBasicDeltaNeutralStrategy(0.01)
	walletSyncUseCase := hedge.NewWalletSyncUseCase(
		walletAdapter,
		hyperliquidAdapter,
		deltaStrategy,
		hedgeRepository,
		cfg.SafeMode,
		cfg.DryRun,
	)

	defaultAsset := cfg.DefaultHedgeAsset
	defaultWallet := cfg.DefaultLPWalletAddress
	defaultHLWallet := cfg.HyperliquidAddress

	hedgeMonitorSvc := monitoring.NewHedgeMonitorService(
		hyperliquidAdapter,
		walletSyncUseCase,
		defaultAsset,
		defaultWallet,
		defaultHLWallet,
	)

	monController := controllers.NewMonitoringController(poolMonitoringService)
	hedgeController := controllers.NewHedgeController(
		walletSyncUseCase,
		defaultAsset,
		defaultWallet,
		defaultHLWallet,
		cfg.SafeMode,
	)
	router := api.NewRouter(authSvc, monController, hedgeController)
	handler := router.Init()

	go poolMonitoringService.Start(ctx)
	go hedgeMonitorSvc.Start(ctx)

	log.Printf("Servidor HTTP iniciado en puerto %d. Presiona Ctrl+C para salir.", cfg.Port)
	if err := api.StartServer(cfg.Port, handler); err != nil {
		log.Fatalf("HTTP server error: %v", err)
	}
}
