// Entidad del dominio que representa una señal de trading.
package domain

import "time"

// SignalType representa el tipo de orden a ejecutar.
type SignalType string

const (
	Buy  SignalType = "BUY"
	Sell SignalType = "SELL"
)

// TradingSignal es el objeto de dominio que encapsula la información de una señal de trading.
// No contiene lógica externa ni referencias a la infraestructura.
type TradingSignal struct {
	ID        string
	Symbol    string
	Price     float64
	Timestamp time.Time
	Type      SignalType // Cambiado de Signal a Type para mayor claridad
	Strategy  string     // Estrategia asociada a la señal
}
