package order

import (
	"context"

	"github.com/ivan-salazar14/markerTradeIa/internal/application/services"
	"github.com/ivan-salazar14/markerTradeIa/internal/domain"
	"github.com/ivan-salazar14/markerTradeIa/internal/domain/ports/out"
)

// TradingService implementa el puerto de la aplicación.
type TradingService struct {
	orchestrator *services.BatchProcessor
}

func NewTradingService(uf out.UserServicePort, tt out.Trader, r out.TradeRepository) *TradingService {
	return &TradingService{orchestrator: services.NewBatchProcessor(uf, tt, r)}
}

func (s *TradingService) ProcessSignalsInBatch(ctx context.Context, signals []domain.TradingSignal) error {
	s.orchestrator.Process(ctx, signals)

	return nil
}
