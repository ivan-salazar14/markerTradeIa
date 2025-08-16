package order

import (
	"context"
	"log"

	"github.com/ivan-salazar14/markerTradeIa/internal/domain"
	"github.com/ivan-salazar14/markerTradeIa/internal/infrastructure/ports/out"
)

// TradingService implementa el puerto de la aplicación.
type TradingService struct {
	trader out.Trader
	repo   out.TradeRepository
}

// NewTradingService es el constructor del servicio de trading.
func NewTradingService(t out.Trader, r out.TradeRepository) *TradingService {
	return &TradingService{trader: t, repo: r}
}

func (s *TradingService) ProcessSignal(ctx context.Context, signal domain.TradingSignal) error {
	log.Printf("Procesando señal de trading para símbolo: %s, precio: %f", signal.Symbol, signal.Price)

	// Aquí iría la lógica de validación de la señal, si fuera necesaria.

	// Usar el puerto Trader para ejecutar la orden
	execution, err := s.trader.ExecuteTrade(ctx, signal)
	if err != nil {
		log.Printf("Error al ejecutar orden para señal %s: %v", signal.ID, err)
		return domain.ErrExecutionFailed
	}

	// Usar el puerto de persistencia para guardar el resultado
	if err := s.repo.SaveTradeExecution(ctx, execution); err != nil {
		log.Printf("Error al guardar ejecución de orden para señal %s: %v", signal.ID, err)
		return err
	}

	log.Printf("Señal %s procesada y ejecutada exitosamente. Estado: %s", signal.ID, execution.Status)
	return nil
}
