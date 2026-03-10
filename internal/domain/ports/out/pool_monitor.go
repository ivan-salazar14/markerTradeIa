package out

import (
	"context"

	"github.com/ivan-salazar14/markerTradeIa/internal/domain"
)

// PoolMonitor es un puerto de salida para monitorear pools de liquidez y posiciones.
type PoolMonitor interface {
	// GetTopPools obtiene los mejores pools según TVL o volumen.
	GetTopPools(ctx context.Context, network string, limit int) ([]domain.LiquidityPool, error)

	// GetPositionStats obtiene estadísticas detalladas de una posición específica.
	GetPositionStats(ctx context.Context, network string, positionID string) (domain.PositionStats, error)
}
