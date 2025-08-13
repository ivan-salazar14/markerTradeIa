package out

import (
	"context"

	"github.com/ivan-salazar14/markerTradeIa/internal/core/domain"
)

// TradeRepository es el puerto para guardar datos de la aplicación.
// Un adaptador de salida (ej. PostgreSQL) implementará esta interfaz.
type TradeRepository interface {
	SaveTradeExecution(ctx context.Context, trade domain.TradeExecution) error
}
