package out

import (
	"context"
)

// HyperliquidPort defines the interface to interact with Hyperliquid exchange
type HyperliquidPort interface {
	// Connect initialized the connection to the exchange
	Connect(ctx context.Context, privateKey string) error 
	// GetBalances returns the mapped balances (asset symbol to amount)
	GetBalances(ctx context.Context, address string) (map[string]float64, error)
	// GetShortPosition returns the absolute size of the short
	GetShortPosition(ctx context.Context, address string, asset string) (float64, error)
	// PlaceMarketOrder places an order for hedging
	PlaceMarketOrder(ctx context.Context, asset string, isBuy bool, size float64) error
}
