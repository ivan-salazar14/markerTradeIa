// Adaptador de salida que implementa el puerto Trader para Binance.
package binance

import (
	"context"
	"log"
	"time"

	"github.com/ivan-salazar14/markerTradeIa/internal/domain"
	"github.com/ivan-salazar14/markerTradeIa/internal/infrastructure/ports/out"
)

// BinanceTrader es un adaptador que implementa el puerto Trader
// para interactuar con el exchange de Binance.
type BinanceTrader struct {
	// client *binance.Client // Cliente de la API de Binance
}

// NewBinanceTrader crea un nuevo adaptador de trader para Binance.
func NewBinanceTrader() out.Trader {
	// Inicialización del cliente de Binance aquí
	return &BinanceTrader{}
}

// ExecuteTrade implementa la lógica para ejecutar una orden en Binance.
func (a *BinanceTrader) ExecuteTrade(ctx context.Context, signal domain.TradingSignal) (domain.TradeExecution, error) {
	log.Printf("Ejecutando orden '%s' en Binance para el símbolo '%s' a precio %.2f...",
		signal.Type, signal.Symbol, signal.Price)

	// Lógica para interactuar con la API de Binance
	// En un caso real, esto llamaría a `binanceClient.NewOrder(...)`

	// get account information for users with active plan to trade

	// Simulación de ejecución exitosa
	execution := domain.TradeExecution{
		ExecutionID: "EXEC-" + signal.ID,
		SignalID:    signal.ID,
		Status:      domain.Success,
		ExecutedAt:  time.Now(),
		Details:     "Orden ejecutada en el mercado.",
	}
	return execution, nil
}
