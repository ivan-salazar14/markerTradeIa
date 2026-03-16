package perps

import (
	"context"
	"fmt"
	"log"

	"github.com/ivan-salazar14/markerTradeIa/internal/domain/ports/out"
)

// HyperliquidAdapter implements out.HyperliquidPort using mock data for development
type HyperliquidAdapter struct {
	clientConnected bool
	apiSecret       string
}

func NewHyperliquidAdapter() out.HyperliquidPort {
	return &HyperliquidAdapter{
		clientConnected: false,
	}
}

func (a *HyperliquidAdapter) Connect(ctx context.Context, privateKey string) error {
	a.clientConnected = true
	a.apiSecret = "MOCKED_" + privateKey
	log.Println("[Hyperliquid] Connected using private key")
	return nil
}

func (a *HyperliquidAdapter) GetBalances(ctx context.Context, address string) (map[string]float64, error) {
	if !a.clientConnected {
		return nil, fmt.Errorf("hyperliquid client not connected")
	}
	
	return map[string]float64{
		"USDC": 12500.0,
		"WETH": 0.0,
	}, nil
}

func (a *HyperliquidAdapter) GetShortPosition(ctx context.Context, address string, asset string) (float64, error) {
	if !a.clientConnected {
		return 0, fmt.Errorf("hyperliquid client not connected")
	}

	// Returning a mock active short equivalent to 0.5 Asset size.
	if asset == "WETH" {
		return 0.5, nil
	}
	return 0.0, nil
}

func (a *HyperliquidAdapter) PlaceMarketOrder(ctx context.Context, asset string, isBuy bool, size float64) error {
	if !a.clientConnected {
		return fmt.Errorf("hyperliquid client not connected")
	}

	side := "SELL"
	if isBuy {
		side = "BUY"
	}
	log.Printf("[Hyperliquid] Placing %s Market Order for %f of %s", side, size, asset)
	return nil
}
