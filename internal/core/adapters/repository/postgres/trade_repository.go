// Adaptador de salida que implementa el puerto TradeRepository.
package postgres

import (
	"context"
	"log"

	"github.com/ivan-salazar14/markerTradeIa/internal/core/domain"
	"github.com/ivan-salazar14/markerTradeIa/internal/core/ports/out"
)

// TradeRepository implementa la interfaz out.TradeRepository para PostgreSQL.
type TradeRepository struct {
	db *connectionDB // Cliente de la base de datos, se debe inicializar en NewTradeRepository
}

// NewTradeRepository crea un nuevo adaptador de repositorio.
func NewTradeRepository() out.TradeRepository {
	// Inicialización del cliente de PostgreSQL
	db := getInstance()
	return &TradeRepository{db: db}
}

// SaveTradeExecution guarda el resultado de una ejecución en la base de datos.
func (r *TradeRepository) SaveTradeExecution(ctx context.Context, trade domain.TradeExecution) error {
	log.Printf("Guardando ejecución de orden %s en PostgreSQL...", trade.ExecutionID)
	// Lógica de SQL para guardar el `trade`
	return nil
}
