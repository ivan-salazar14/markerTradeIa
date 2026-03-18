package monitoring

import (
	"context"
	"testing"
	"time"

	"github.com/ivan-salazar14/markerTradeIa/internal/domain"
)

type poolMonitorMock struct {
	pools []domain.LiquidityPool
	calls int
}

func (m *poolMonitorMock) GetTopPools(ctx context.Context, network string, limit int) ([]domain.LiquidityPool, error) {
	m.calls++
	if limit > 0 && limit < len(m.pools) {
		return m.pools[:limit], nil
	}
	return m.pools, nil
}

func (m *poolMonitorMock) GetPositionStats(ctx context.Context, network string, positionID string) (domain.PositionStats, error) {
	return domain.PositionStats{}, nil
}

func TestGetPoolsCachesLatestSnapshot(t *testing.T) {
	monitor := &poolMonitorMock{
		pools: []domain.LiquidityPool{
			{ID: "1", Network: "mainnet", Symbol0: "ETH", Symbol1: "USDC"},
			{ID: "2", Network: "mainnet", Symbol0: "WBTC", Symbol1: "USDC"},
		},
	}
	service := NewMonitoringService(monitor, []string{"mainnet"}, time.Minute)

	pools, updatedAt, err := service.GetPools(context.Background(), "mainnet", 5)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(pools) != 2 {
		t.Fatalf("expected 2 pools, got %d", len(pools))
	}
	if updatedAt.IsZero() {
		t.Fatalf("expected updatedAt to be set")
	}

	_, _, err = service.GetPools(context.Background(), "mainnet", 1)
	if err != nil {
		t.Fatalf("expected no error on cached call, got %v", err)
	}
	if monitor.calls != 1 {
		t.Fatalf("expected pool monitor to be called once, got %d", monitor.calls)
	}
}
