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
	"github.com/ivan-salazar14/markerTradeIa/internal/domain/ports/out"
	"github.com/ivan-salazar14/markerTradeIa/internal/infrastructure/adapters/kafka"
	"github.com/ivan-salazar14/markerTradeIa/internal/infrastructure/adapters/repository/database"
	"github.com/ivan-salazar14/markerTradeIa/internal/infrastructure/adapters/repository/tradeAdapter"
	"github.com/ivan-salazar14/markerTradeIa/internal/infrastructure/adapters/trading/binance"
	"github.com/ivan-salazar14/markerTradeIa/internal/infrastructure/adapters/user"
)

func main() {
	log.Println("Iniciando servicio de trading...")
	migrator := database.NewMigrator()
	migrator.CreateStructures()
	// Inicializar los adaptadores de salida
	tradeRepository := tradeAdapter.NewTradeRepository()
	binanceTrader := binance.NewBinanceTrader()
	userAdapter := user.NewHttpUserService("http://localhost:8080/users")
	tt := out.Trader(binanceTrader)

	tradingService := order.NewTradingService(userAdapter, tt, tradeRepository)

	// Inicializar el adaptador de entrada de Kafka, inyectando el servicio de aplicación
	kafkaConsumer := kafka.NewConsumerAdapter(tradingService)

	// Contexto para manejar la cancelación del servicio
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Iniciar el consumidor de Kafka y esperar a que termine
	log.Println("Servicio de trading iniciado. Esperando señales...")
	if err := kafkaConsumer.StartConsuming(ctx); err != nil {
		log.Fatalf("Fallo al iniciar el consumidor de Kafka: %v", err)
	}
	log.Println("Servicio de trading finalizado.")
}
