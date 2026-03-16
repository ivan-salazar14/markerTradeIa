// Este archivo representa la estructura completa del servicio de trading en Go

// go.mod
// Módulo para la gestión de dependencias.
// En un proyecto real, se agregarían las dependencias de Kafka, PostgreSQL, etc.
//
// go.mod
// module MarkerTradeia
// go 1.25

// go.sum
// Archivo de suma de verificación de dependencias.
//

// file: cmd/main.go
// Punto de entrada de la aplicación. Aquí se "cablean" todas las dependencias.
package main

import (
	"context"
	"log"

	"github.com/ivan-salazar14/markerTradeIa/config"
	"github.com/ivan-salazar14/markerTradeIa/internal/application/services/auth"
	"github.com/ivan-salazar14/markerTradeIa/internal/application/services/monitoring"
	"github.com/ivan-salazar14/markerTradeIa/internal/domain"
	"github.com/ivan-salazar14/markerTradeIa/internal/infrastructure/adapters/api"
	"github.com/ivan-salazar14/markerTradeIa/internal/infrastructure/adapters/api/controllers"
	"github.com/ivan-salazar14/markerTradeIa/internal/infrastructure/adapters/monitoring/revert"
	"github.com/ivan-salazar14/markerTradeIa/internal/infrastructure/adapters/repository/database"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Error cargando configuración: %v", err)
	}

	log.Println("Iniciando servicio de trading con adaptador HyperLiquid...")
	migrator := database.NewMigrator()
	migrator.CreateStructures()

	// Inicializar los adaptadores de salida
	//tradeRepository := tradeAdapter.NewTradeRepository()
	//userAdapter := user.NewHttpUserService("http://localhost:8080/users")

	// Usar HyperLiquid Trader
	//trader := hyperliquid.NewHyperLiquidTrader("0xYOUR_ADDRESS", "0xYOUR_PRIVATE_KEY")
	//tt := out.Trader(trader)

	//tradingService := order.NewTradingService(userAdapter, tt, tradeRepository)

	// Inicializar Auth Service
	authSvc := auth.NewAuthService(domain.AuthConfig{
		JWTSecret:      cfg.JWTSecret,
		AccessExpiry:   cfg.AccessExpiry,
		RefreshExpiry:  cfg.RefreshExpiry,
		ServiceAPIKeys: cfg.ServiceAPIKeys,
	})

	// Inicializar y arrancar el monitoreo de pools de Uniswap (Revert Finance)
	revertAdapter := revert.NewRevertAdapter(cfg.RevertBaseURL)
	poolMonitoringService := monitoring.NewMonitoringService(
		revertAdapter,
		cfg.RevertNetworks,
		cfg.MonitoringInterval,
	)

	// Setup API
	monController := controllers.NewMonitoringController(poolMonitoringService)
	hedgeController := controllers.NewHedgeController()
	router := api.NewRouter(authSvc, monController, hedgeController)
	handler := router.Init()

	// Inicializar el adaptador de entrada de Kafka, inyectando el servicio de aplicación
	//	kafkaConsumer := kafka.NewConsumerAdapter(tradingService)

	// Contexto para manejar la cancelación del servicio
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Iniciar el monitoreo en segundo plano
	go poolMonitoringService.Start(ctx)

	// Iniciar Servidor HTTP (este comando bloquea el hilo principal para mantener la API viva)
	log.Println("Servidor HTTP iniciado en puerto 8081. Presiona Ctrl+C para salir.")
	if err := api.StartServer(8081, handler); err != nil {
		log.Fatalf("HTTP Server error: %v", err)
	}

	// Iniciar el consumidor de Kafka y esperar a que termine
	/*	log.Println("Servicio de trading iniciado. Esperando señales en Kafka...")
		if err := kafkaConsumer.StartConsuming(ctx); err != nil {
			log.Fatalf("Fallo al iniciar el consumidor de Kafka: %v", err)
		}
		log.Println("Servicio de trading finalizado.")*/
}
