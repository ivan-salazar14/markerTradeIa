// Puerto de entrada para recibir eventos del mundo exterior.
package in

import (
	"context"

	"github.com/ivan-salazar14/markerTradeIa/internal/domain"
)

// EventReceiver es el puerto de entrada para recibir eventos.
// Un adaptador de entrada (ej. Kafka) implementará esta interfaz.
type EventReceiver interface {
	StartConsuming(ctx context.Context) error
}

// TradingServicePort es el puerto que el adaptador de entrada usa para llamar al servicio.
type TradingServicePort interface {
	ProcessSignal(ctx context.Context, signal domain.TradingSignal) error
}
