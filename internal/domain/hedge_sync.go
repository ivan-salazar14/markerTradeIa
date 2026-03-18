package domain

import "time"

type SyncHedgeResult struct {
	Asset              string      `json:"asset"`
	WalletAddress      string      `json:"wallet_address"`
	HyperliquidAddress string      `json:"hyperliquid_address"`
	PoolExposure       float64     `json:"pool_exposure"`
	ShortExposure      float64     `json:"short_exposure"`
	NetExposure        float64     `json:"net_exposure"`
	Status             string      `json:"status"`
	Action             HedgeAction `json:"action"`
	Executed           bool        `json:"executed"`
	SafeMode           bool        `json:"safe_mode"`
	DryRun             bool        `json:"dry_run"`
	Message            string      `json:"message"`
	LastSync           time.Time   `json:"last_sync"`
}

type ManualSyncRequest struct {
	Asset              string `json:"asset"`
	WalletAddress      string `json:"wallet_address"`
	HyperliquidAddress string `json:"hyperliquid_address"`
}
