package out

import (
	"context"

	"github.com/ivan-salazar14/markerTradeIa/internal/domain"
)

type HedgeRepository interface {
	SaveWalletConnection(ctx context.Context, walletType string, address string, status string) error
	SaveHedgeState(ctx context.Context, result domain.SyncHedgeResult) error
	SaveHedgeAction(ctx context.Context, result domain.SyncHedgeResult) error
	SaveSyncEvent(ctx context.Context, triggerType string, result domain.SyncHedgeResult) error
	GetLatestHedgeState(ctx context.Context, asset string) (*domain.SyncHedgeResult, error)
	GetWalletConnections(ctx context.Context) ([]domain.WalletInfo, error)
}
