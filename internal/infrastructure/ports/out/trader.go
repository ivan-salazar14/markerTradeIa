package out

import (
	"context"

	"github.com/ivan-salazar14/markerTradeIa/internal/domain"
)

// Trader es el puerto para ejecutar operaciones de trading.
// Un adaptador de salida (ej. Binance) implementará esta interfaz.
type Trader interface {
	ExecuteTrade(ctx context.Context, signal domain.TradingSignal) (domain.TradeExecution, error)
}
