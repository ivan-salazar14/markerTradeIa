// Adaptador de salida que implementa el puerto TradeRepository.
package tradeAdapter

import (
	"context"
	"log"

	"github.com/ivan-salazar14/markerTradeIa/internal/domain"
	"github.com/ivan-salazar14/markerTradeIa/internal/infrastructure/adapters/repository/database"
	"github.com/ivan-salazar14/markerTradeIa/internal/infrastructure/adapters/repository/tradeAdapter/mappers"
	"github.com/ivan-salazar14/markerTradeIa/internal/infrastructure/ports/out"
	"gorm.io/gorm"
)

// TradeRepository implementa la interfaz out.TradeRepository para PostgreSQL.
type TradeRepository struct {
	db *gorm.DB // Cliente de la base de datos, se debe inicializar en NewTradeRepository
}

// NewTradeRepository crea un nuevo adaptador de repositorio.
func NewTradeRepository() out.TradeRepository {
	// Inicialización del cliente de PostgreSQL
	return &TradeRepository{db: database.GetInstance()}
}

// SaveTradeExecution guarda el resultado de una ejecución en la base de datos.
func (r *TradeRepository) SaveTradeExecution(ctx context.Context, trade domain.TradeExecution) error {
	log.Printf("Guardando ejecución de orden %s en PostgreSQL...", trade.ExecutionID)
	// Lógica de SQL para guardar el `trade`
	model := mappers.TradeDomainToModel(trade, "UserID")
	tx := r.db.Save(&model)

	return tx.Error
}
