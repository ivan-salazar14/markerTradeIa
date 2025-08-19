package services

import (
	"context"
	"log"

	"github.com/ivan-salazar14/markerTradeIa/internal/domain"
	"github.com/ivan-salazar14/markerTradeIa/internal/domain/ports/in"
	"github.com/ivan-salazar14/markerTradeIa/internal/domain/ports/out"
)

// BatchProcessor es un servicio de la capa de aplicación que orquesta el procesamiento concurrente
type BatchProcessor struct {
	userFinder    in.UserServicePort  // Puerto para obtener usuarios (adaptador de DB)
	tradingTrader out.Trader          // Puerto para ejecutar trades (BinanceTrader)
	repo          out.TradeRepository // Puerto para guardar ejecuciones de trades
}

// NewBatchProcessor es el constructor del servicio.
func NewBatchProcessor(uf in.UserServicePort, tt out.Trader, r out.TradeRepository) *BatchProcessor {
	return &BatchProcessor{userFinder: uf, tradingTrader: tt, repo: r}
}

// Process es el método principal que usará el usecase.
func (s *BatchProcessor) Process(ctx context.Context, signals []domain.TradingSignal) ([]domain.TradeExecution, error) {
	users, err := s.userFinder.GetUsers()
	if err != nil {
		log.Printf("Error al obtener usuarios: %v", err)
		return nil, err
	}
	log.Printf("Procesando %d señales de trading para %d usuarios...", len(signals), len(users))
	// Llamamos al método del trader para procesar todas las señales en paralelo
	executedTrades := s.tradingTrader.ProcessBatch(ctx, users, signals)

	// Lógica para guardar todos los resultados
	for _, trade := range executedTrades {
		if err := s.repo.SaveTradeExecution(ctx, trade); err != nil {
			log.Printf("Error al guardar ejecución de orden: %v", err)
			// Decide aquí si quieres retornar un error o continuar
		}
	}
	log.Printf("Procesamiento de señales completado. Total de ejecuciones: %d", len(executedTrades))

	return executedTrades, nil
}
