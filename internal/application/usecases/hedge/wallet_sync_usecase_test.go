package hedge

import (
	"context"
	"errors"
	"testing"

	"github.com/ivan-salazar14/markerTradeIa/internal/application/usecases/hedge/strategies"
	"github.com/ivan-salazar14/markerTradeIa/internal/domain"
)

type walletPortMock struct {
	pools []domain.ActivePool
	err   error
}

func (m walletPortMock) GetBalances(ctx context.Context, address string) (map[string]float64, error) {
	return map[string]float64{"ETH": 1}, nil
}

func (m walletPortMock) GetActivePoolPositions(ctx context.Context, address string) ([]domain.ActivePool, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.pools, nil
}

type hyperliquidPortMock struct {
	short      float64
	shortErr   error
	orderErr   error
	orderCount int
}

func (m *hyperliquidPortMock) Connect(ctx context.Context, privateKey string) error { return nil }
func (m *hyperliquidPortMock) GetBalances(ctx context.Context, address string) (map[string]float64, error) {
	return map[string]float64{"USDC": 1000}, nil
}
func (m *hyperliquidPortMock) GetShortPosition(ctx context.Context, address string, asset string) (float64, error) {
	if m.shortErr != nil {
		return 0, m.shortErr
	}
	return m.short, nil
}
func (m *hyperliquidPortMock) PlaceMarketOrder(ctx context.Context, asset string, isBuy bool, size float64) error {
	m.orderCount++
	return m.orderErr
}
func (m *hyperliquidPortMock) SubscribeToMarketUpdates(ctx context.Context, asset string, priceCh chan<- float64) error {
	return nil
}
func (m *hyperliquidPortMock) SubscribeToUserEvents(ctx context.Context, address string, sizeCh chan<- float64) error {
	return nil
}

type hedgeRepositoryMock struct {
	lastResult *domain.SyncHedgeResult
}

func (m *hedgeRepositoryMock) SaveWalletConnection(ctx context.Context, walletType string, address string, status string) error {
	return nil
}
func (m *hedgeRepositoryMock) SaveHedgeState(ctx context.Context, result domain.SyncHedgeResult) error {
	copy := result
	m.lastResult = &copy
	return nil
}
func (m *hedgeRepositoryMock) SaveHedgeAction(ctx context.Context, result domain.SyncHedgeResult) error {
	return nil
}
func (m *hedgeRepositoryMock) SaveSyncEvent(ctx context.Context, triggerType string, result domain.SyncHedgeResult) error {
	return nil
}
func (m *hedgeRepositoryMock) GetLatestHedgeState(ctx context.Context, asset string) (*domain.SyncHedgeResult, error) {
	return m.lastResult, nil
}
func (m *hedgeRepositoryMock) GetWalletConnections(ctx context.Context) ([]domain.WalletInfo, error) {
	return nil, nil
}

func TestSyncHedgeReturnsSyncedWhenAlreadyBalanced(t *testing.T) {
	repo := &hedgeRepositoryMock{}
	hl := &hyperliquidPortMock{short: 0.25}
	uc := NewWalletSyncUseCase(
		walletPortMock{pools: []domain.ActivePool{{Symbol: "ETH", Size: 0.25}}},
		hl,
		strategies.NewBasicDeltaNeutralStrategy(0.01),
		repo,
		false,
		false,
	)

	result, err := uc.SyncHedge(context.Background(), "wallet-a", "wallet-b", "ETH")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Status != "synced" {
		t.Fatalf("expected synced status, got %s", result.Status)
	}
	if result.Executed {
		t.Fatalf("expected no execution")
	}
}

func TestSyncHedgeReturnsSimulatedInSafeMode(t *testing.T) {
	repo := &hedgeRepositoryMock{}
	hl := &hyperliquidPortMock{short: 0}
	uc := NewWalletSyncUseCase(
		walletPortMock{pools: []domain.ActivePool{{Symbol: "ETH", Size: 0.25}}},
		hl,
		strategies.NewBasicDeltaNeutralStrategy(0.01),
		repo,
		true,
		true,
	)

	result, err := uc.SyncHedge(context.Background(), "wallet-a", "wallet-b", "ETH")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Status != "simulated" {
		t.Fatalf("expected simulated status, got %s", result.Status)
	}
	if hl.orderCount != 0 {
		t.Fatalf("expected no order placement in safe mode")
	}
}

func TestSyncHedgeReturnsErrorWhenWalletFails(t *testing.T) {
	uc := NewWalletSyncUseCase(
		walletPortMock{err: errors.New("wallet failed")},
		&hyperliquidPortMock{},
		strategies.NewBasicDeltaNeutralStrategy(0.01),
		&hedgeRepositoryMock{},
		false,
		false,
	)

	result, err := uc.SyncHedge(context.Background(), "wallet-a", "wallet-b", "ETH")
	if err == nil {
		t.Fatalf("expected error")
	}
	if result.Status != "error" {
		t.Fatalf("expected error status, got %s", result.Status)
	}
}

func TestSyncHedgeReturnsErrorWhenShortLookupFails(t *testing.T) {
	uc := NewWalletSyncUseCase(
		walletPortMock{pools: []domain.ActivePool{{Symbol: "ETH", Size: 0.25}}},
		&hyperliquidPortMock{shortErr: errors.New("short failed")},
		strategies.NewBasicDeltaNeutralStrategy(0.01),
		&hedgeRepositoryMock{},
		false,
		false,
	)

	result, err := uc.SyncHedge(context.Background(), "wallet-a", "wallet-b", "ETH")
	if err == nil {
		t.Fatalf("expected error")
	}
	if result.Status != "error" {
		t.Fatalf("expected error status, got %s", result.Status)
	}
}
