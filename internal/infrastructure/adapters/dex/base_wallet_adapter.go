package dex

import (
	"context"

	"github.com/ivan-salazar14/markerTradeIa/internal/domain"
	"github.com/ivan-salazar14/markerTradeIa/internal/domain/ports/out"
)

type EVMWalletAdapter struct {
	// in future include ethclient here
}

func NewEVMWalletAdapter() out.WalletPort {
	return &EVMWalletAdapter{}
}

func (a *EVMWalletAdapter) GetBalances(ctx context.Context, address string) (map[string]float64, error) {
	// Mock balances
	return map[string]float64{
		"USDC": 100.0,
		"WETH": 0.5,
	}, nil
}

func (a *EVMWalletAdapter) GetActivePoolPositions(ctx context.Context, address string) ([]domain.ActivePool, error) {
	// Mock active pools
	return []domain.ActivePool{
		{
			PoolID:   "0x123...456",
			Symbol:   "WETH-USDC",
			Size:     1.2, // Let's say we have 1.2 WETH of exposure in our Uniswap pool
			ValueUsd: 3500.0,
		},
	}, nil
}
