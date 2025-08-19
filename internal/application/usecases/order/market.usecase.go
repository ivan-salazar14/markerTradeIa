package order

import (
	"context"
	"log"

	"github.com/ivan-salazar14/markerTradeIa/internal/application/services"
	"github.com/ivan-salazar14/markerTradeIa/internal/domain"
	"github.com/ivan-salazar14/markerTradeIa/internal/domain/ports/in"
	"github.com/ivan-salazar14/markerTradeIa/internal/domain/ports/out"
)

// TradingService implementa el puerto de la aplicación.
type TradingService struct {
	orchestrator *services.BatchProcessor
}

// NewTradingService es el constructor del servicio de trading.
func NewTradingService(uf in.UserServicePort, tt out.Trader, r out.TradeRepository) *TradingService {
	return &TradingService{orchestrator: services.NewBatchProcessor(uf, tt, r)}
}

// Este es un fragmento de tu usecase.
func (s *TradingService) ProcessSignalsInBatch(ctx context.Context, signals []domain.TradingSignal) error {
	log.Printf("Procesando %d señales de trading en batch...", len(signals))
	s.orchestrator.Process(ctx, signals)

	return nil
}
