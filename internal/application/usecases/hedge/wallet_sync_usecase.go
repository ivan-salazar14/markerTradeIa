package hedge

import (
	"context"
	"fmt"
	"time"

	"github.com/ivan-salazar14/markerTradeIa/internal/domain"
	"github.com/ivan-salazar14/markerTradeIa/internal/domain/ports/out"
)

type WalletSyncUseCase struct {
	walletPort      out.WalletPort
	hyperliquidPort out.HyperliquidPort
	strategy        domain.IHedgeStrategy
	hedgeRepository out.HedgeRepository
	safeMode        bool
	dryRun          bool
}

func NewWalletSyncUseCase(
	walletPort out.WalletPort,
	hyperliquidPort out.HyperliquidPort,
	strategy domain.IHedgeStrategy,
	hedgeRepository out.HedgeRepository,
	safeMode bool,
	dryRun bool,
) *WalletSyncUseCase {
	return &WalletSyncUseCase{
		walletPort:      walletPort,
		hyperliquidPort: hyperliquidPort,
		strategy:        strategy,
		hedgeRepository: hedgeRepository,
		safeMode:        safeMode,
		dryRun:          dryRun,
	}
}

// ConnectAndFetchWallet connects an EVM wallet and returns its balances and active Uniswap V3 pools.
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

func (u *WalletSyncUseCase) RegisterWallet(ctx context.Context, walletType string, address string) error {
	if u.hedgeRepository == nil {
		return nil
	}
	return u.hedgeRepository.SaveWalletConnection(ctx, walletType, address, "connected")
}

func (u *WalletSyncUseCase) GetLatestDelta(ctx context.Context, asset string) (*domain.SyncHedgeResult, error) {
	if u.hedgeRepository == nil {
		return nil, nil
	}
	return u.hedgeRepository.GetLatestHedgeState(ctx, asset)
}

func (u *WalletSyncUseCase) GetWalletConnections(ctx context.Context) ([]domain.WalletInfo, error) {
	if u.hedgeRepository == nil {
		return nil, nil
	}
	return u.hedgeRepository.GetWalletConnections(ctx)
}

// SyncHedge retrieves current LP exposure and current short, then uses strategy to generate an action.
func (u *WalletSyncUseCase) SyncHedge(ctx context.Context, addressA string, hlAddressB string, asset string) (domain.SyncHedgeResult, error) {
	result := domain.SyncHedgeResult{
		Asset:              asset,
		WalletAddress:      addressA,
		HyperliquidAddress: hlAddressB,
		SafeMode:           u.safeMode,
		DryRun:             u.dryRun,
		LastSync:           time.Now().UTC(),
		Status:             "pending",
	}

	pools, err := u.walletPort.GetActivePoolPositions(ctx, addressA)
	if err != nil {
		result.Status = "error"
		result.Message = fmt.Sprintf("failed fetching active pools: %v", err)
		u.persistSync(ctx, "manual", result)
		return result, err
	}

	for _, p := range pools {
		if p.Symbol == asset {
			result.PoolExposure += p.Size
		}
	}

	currentShort, err := u.hyperliquidPort.GetShortPosition(ctx, hlAddressB, asset)
	if err != nil {
		result.Status = "error"
		result.Message = fmt.Sprintf("failed getting current short: %v", err)
		u.persistSync(ctx, "manual", result)
		return result, err
	}
	result.ShortExposure = currentShort
	result.NetExposure = result.PoolExposure - result.ShortExposure

	action, err := u.strategy.Evaluate(ctx, result.PoolExposure, result.ShortExposure)
	if err != nil {
		result.Status = "error"
		result.Message = fmt.Sprintf("strategy evaluation failed: %v", err)
		u.persistSync(ctx, "manual", result)
		return result, err
	}
	result.Action = action

	if action.ActionType == "ADJUST_SHORT" {
		if u.safeMode || u.dryRun {
			result.Status = "simulated"
			result.Executed = false
			result.Message = "hedge adjustment required but execution skipped because safe mode or dry run is enabled"
		} else {
			err = u.hyperliquidPort.PlaceMarketOrder(ctx, action.Asset, false, action.Size)
			if err != nil {
				result.Status = "error"
				result.Message = fmt.Sprintf("failed to place hyperliquid order: %v", err)
				u.persistSync(ctx, "manual", result)
				return result, err
			}
			result.Status = "adjusted"
			result.Executed = true
			result.Message = "hedge adjustment executed successfully"
		}
	} else {
		result.Status = "synced"
		result.Message = "hedge is already synchronized"
	}

	u.persistSync(ctx, "manual", result)
	return result, nil
}

func (u *WalletSyncUseCase) persistSync(ctx context.Context, triggerType string, result domain.SyncHedgeResult) {
	if u.hedgeRepository == nil {
		return
	}
	_ = u.hedgeRepository.SaveHedgeState(ctx, result)
	_ = u.hedgeRepository.SaveHedgeAction(ctx, result)
	_ = u.hedgeRepository.SaveSyncEvent(ctx, triggerType, result)
}
