package tradeAdapter

import (
	"context"
	"errors"
	"log"

	"github.com/ivan-salazar14/markerTradeIa/internal/domain"
	"github.com/ivan-salazar14/markerTradeIa/internal/domain/ports/out"
	"github.com/ivan-salazar14/markerTradeIa/internal/infrastructure/adapters/repository/database"
	"github.com/ivan-salazar14/markerTradeIa/internal/infrastructure/adapters/repository/tradeAdapter/mappers"
	"gorm.io/gorm"
)

// TradeRepository implementa la interfaz out.TradeRepository para PostgreSQL.
type TradeRepository struct {
	db *gorm.DB
}

// NewTradeRepository crea un nuevo adaptador de repositorio.
func NewTradeRepository() out.TradeRepository {
	return &TradeRepository{db: database.GetInstance()}
}

// SaveTradeExecution guarda el resultado de una ejecucion en la base de datos.
func (r *TradeRepository) SaveTradeExecution(ctx context.Context, trade domain.TradeExecution) error {
	log.Printf("Guardando ejecucion de orden %s en PostgreSQL y estatus %s", trade.ExecutionID, trade.Status)
	if r.db == nil || r.db.Dialector == nil {
		return errors.New("trade repository database is not initialized")
	}

	model := mappers.TradeDomainToModel(trade, "UserID")
	tx := r.db.WithContext(ctx).Save(&model)
	return tx.Error
}
