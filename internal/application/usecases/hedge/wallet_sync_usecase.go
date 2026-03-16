package hedge

import (
	"context"
	"fmt"

	"github.com/ivan-salazar14/markerTradeIa/internal/domain"
	"github.com/ivan-salazar14/markerTradeIa/internal/domain/ports/out"
)

type WalletSyncUseCase struct {
	walletPort      out.WalletPort
	hyperliquidPort out.HyperliquidPort
	strategy        domain.IHedgeStrategy
}

func NewWalletSyncUseCase(
	walletPort out.WalletPort,
	hyperliquidPort out.HyperliquidPort,
	strategy domain.IHedgeStrategy,
) *WalletSyncUseCase {
	return &WalletSyncUseCase{
		walletPort:      walletPort,
		hyperliquidPort: hyperliquidPort,
		strategy:        strategy,
	}
}

// ConnectAndFetchWallet connects an EVM wallet and returns its balances and active Uniswap V3 pools
func (u *WalletSyncUseCase) ConnectAndFetchWallet(ctx context.Context, address string) (domain.WalletData, error) {
	balances, err := u.walletPort.GetBalances(ctx, address)
	if err != nil {
		return domain.WalletData{}, fmt.Errorf("failed getting EVM balances: %w", err)
	}

	pools, err := u.walletPort.GetActivePoolPositions(ctx, address)
	if err != nil {
		return domain.WalletData{}, fmt.Errorf("failed fetching active pools: %w", err)
	}

	var parsedBalances []domain.WalletBalance
	for asset, amt := range balances {
		parsedBalances = append(parsedBalances, domain.WalletBalance{
			Asset:  asset,
			Amount: amt,
		})
	}

	return domain.WalletData{
		Address:     address,
		Balances:    parsedBalances,
		ActivePools: pools,
	}, nil
}

// SyncHedge retrieves current LP exposure and current short, then uses strategy to generate an action.
// If an adjustment is required, it commands the Hyperliquid adapter to place a market order to hedge.
func (u *WalletSyncUseCase) SyncHedge(ctx context.Context, addressA string, hlAddressB string, asset string) error {
	// 1. Get LP Exposure (using wallet port or a specific pool monitor port; for now mocking via an active pool)
	pools, err := u.walletPort.GetActivePoolPositions(ctx, addressA)
	if err != nil {
		return err
	}

	var lpExposure float64
	for _, p := range pools {
		if p.Symbol == asset {
			lpExposure += p.Size
		}
	}

	// 2. Get Current Short Position on Hyperliquid
	currentShort, err := u.hyperliquidPort.GetShortPosition(ctx, hlAddressB, asset)
	if err != nil {
		return err
	}

	// 3. Evaluate the Strategy
	action, err := u.strategy.Evaluate(ctx, lpExposure, currentShort)
	if err != nil {
		return fmt.Errorf("strategy evaluation failed: %w", err)
	}

	// 4. Take action
	if action.ActionType == "ADJUST_SHORT" {
		fmt.Printf("[Hedge] Adjusting short for %s. New target size: %f. Reason: %s\n", action.Asset, action.Size, action.Reason)
		err = u.hyperliquidPort.PlaceMarketOrder(ctx, action.Asset, false, action.Size)
		if err != nil {
			return fmt.Errorf("failed to place hyperliquid order: %w", err)
		}
	} else {
		fmt.Printf("[Hedge] Hedge is fully synced. No action required.\n")
	}

	return nil
}
