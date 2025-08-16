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

	"github.com/ivan-salazar14/markerTradeIa/internal/application/usecases/order"
	"github.com/ivan-salazar14/markerTradeIa/internal/infrastructure/adapters/kafka"
	"github.com/ivan-salazar14/markerTradeIa/internal/infrastructure/adapters/repository/database"
	"github.com/ivan-salazar14/markerTradeIa/internal/infrastructure/adapters/repository/tradeAdapter"
	"github.com/ivan-salazar14/markerTradeIa/internal/infrastructure/adapters/trading/binance"
)

func main() {
	log.Println("Iniciando servicio de trading...")
	migrator := database.NewMigrator()
	migrator.CreateStructures()
	// Inicializar los adaptadores de salida
	tradeRepository := tradeAdapter.NewTradeRepository()

	binanceTrader := binance.NewBinanceTrader()

	// Inicializar el servicio de aplicación, inyectando los adaptadores de salida
	tradingService := order.NewTradingService(binanceTrader, tradeRepository)

	// Inicializar el adaptador de entrada de Kafka, inyectando el servicio de aplicación
	kafkaConsumer := kafka.NewConsumerAdapter(tradingService)

	// Contexto para manejar la cancelación del servicio
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Iniciar el consumidor de Kafka en una goroutine para que el main no se bloquee
	go func() {
		if err := kafkaConsumer.StartConsuming(ctx); err != nil {
			log.Fatalf("Fallo al iniciar el consumidor de Kafka: %v", err)
		}
	}()

	log.Println("Servicio de trading iniciado. Esperando señales...")
	// Bloquear el main para mantener el programa en ejecución
	select {}
}
