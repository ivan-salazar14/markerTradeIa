package hedgeAdapter

import (
	"context"
	"errors"

	"github.com/ivan-salazar14/markerTradeIa/internal/domain"
	"github.com/ivan-salazar14/markerTradeIa/internal/domain/ports/out"
	"github.com/ivan-salazar14/markerTradeIa/internal/infrastructure/adapters/repository/database"
	hedgeModels "github.com/ivan-salazar14/markerTradeIa/internal/infrastructure/adapters/repository/hedgeAdapter/models"
	"gorm.io/gorm"
)

type HedgeRepository struct {
	db *gorm.DB
}

func NewHedgeRepository() out.HedgeRepository {
	return &HedgeRepository{db: database.GetInstance()}
}

func (r *HedgeRepository) SaveWalletConnection(ctx context.Context, walletType string, address string, status string) error {
	model := hedgeModels.WalletConnection{
		WalletType: walletType,
		Address:    address,
		Status:     status,
	}
	return r.db.WithContext(ctx).Create(&model).Error
}

func (r *HedgeRepository) SaveHedgeState(ctx context.Context, result domain.SyncHedgeResult) error {
	model := hedgeModels.HedgeState{
		Asset:              result.Asset,
		WalletAddress:      result.WalletAddress,
		HyperliquidAddress: result.HyperliquidAddress,
		PoolExposure:       result.PoolExposure,
		ShortExposure:      result.ShortExposure,
		NetExposure:        result.NetExposure,
		Status:             result.Status,
		Message:            result.Message,
		SafeMode:           result.SafeMode,
		DryRun:             result.DryRun,
		LastSync:           result.LastSync,
	}
	return r.db.WithContext(ctx).Create(&model).Error
}

func (r *HedgeRepository) SaveHedgeAction(ctx context.Context, result domain.SyncHedgeResult) error {
	model := hedgeModels.HedgeAction{
		Asset:      result.Asset,
		ActionType: result.Action.ActionType,
		Size:       result.Action.Size,
		Status:     result.Status,
		Reason:     result.Action.Reason,
		Executed:   result.Executed,
	}
	return r.db.WithContext(ctx).Create(&model).Error
}

func (r *HedgeRepository) SaveSyncEvent(ctx context.Context, triggerType string, result domain.SyncHedgeResult) error {
	model := hedgeModels.SyncEvent{
		TriggerType: triggerType,
		Asset:       result.Asset,
		Success:     result.Status != "error",
		Message:     result.Message,
	}
	return r.db.WithContext(ctx).Create(&model).Error
}

func (r *HedgeRepository) GetLatestHedgeState(ctx context.Context, asset string) (*domain.SyncHedgeResult, error) {
	var model hedgeModels.HedgeState
	query := r.db.WithContext(ctx).Order("last_sync desc")
	if asset != "" {
		query = query.Where("asset = ?", asset)
	}
	err := query.First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	result := &domain.SyncHedgeResult{
		Asset:              model.Asset,
		WalletAddress:      model.WalletAddress,
		HyperliquidAddress: model.HyperliquidAddress,
		PoolExposure:       model.PoolExposure,
		ShortExposure:      model.ShortExposure,
		NetExposure:        model.NetExposure,
		Status:             model.Status,
		Message:            model.Message,
		SafeMode:           model.SafeMode,
		DryRun:             model.DryRun,
		LastSync:           model.LastSync,
	}
	return result, nil
}

func (r *HedgeRepository) GetWalletConnections(ctx context.Context) ([]domain.WalletInfo, error) {
	var models []hedgeModels.WalletConnection
	if err := r.db.WithContext(ctx).Order("created_at desc").Find(&models).Error; err != nil {
		return nil, err
	}

	wallets := make([]domain.WalletInfo, 0, len(models))
	for _, model := range models {
		address := model.Address
		wallets = append(wallets, domain.WalletInfo{
			Type:        model.WalletType,
			Name:        model.WalletType,
			Connected:   model.Status == "connected",
			Address:     &address,
			FullAddress: &address,
		})
	}

	return wallets, nil
}
