package out

import (
	"context"

	"github.com/ivan-salazar14/markerTradeIa/internal/domain"
)

// WalletPort defines operations for monitoring on-chain EVM wallet balances and active LP positions (Uniswap V3, etc)
type WalletPort interface {
	GetBalances(ctx context.Context, address string) (map[string]float64, error)
	GetActivePoolPositions(ctx context.Context, address string) ([]domain.ActivePool, error)
}
