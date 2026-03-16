package domain

import "time"

type HedgeStrategy struct {
	StrategyID  string    `json:"strategy_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	IsSynced    bool      `json:"is_synced"`
	CreatedAt   time.Time `json:"created_at"`
}

type TrendStats struct {
	Value          float64 `json:"value"`
	Trend          float64 `json:"trend"`
	TrendDirection string  `json:"trend_direction"` // "up", "down"
}

type CurrencyStats struct {
	Value          float64 `json:"value"`
	Currency       string  `json:"currency"`
	Trend          float64 `json:"trend"`
	TrendDirection string  `json:"trend_direction"`
}

type UnitStats struct {
	Value float64 `json:"value"`
	Unit  string  `json:"unit"`
}

type HedgeStats struct {
	APR              TrendStats    `json:"apr"`
	FeesAccumulated  CurrencyStats `json:"fees_accumulated"`
	Delta            UnitStats     `json:"delta"`
	HedgeEfficiency TrendStats    `json:"hedge_efficiency"`
}

type WalletInfo struct {
	Type            string   `json:"type"`
	Name            string   `json:"name"`
	Description     string   `json:"description"`
	Connected       bool     `json:"connected"`
	Address         *string  `json:"address"`
	FullAddress     *string  `json:"full_address"`
	Permissions     []string `json:"permissions"`
	PermissionsNote string   `json:"permissions_note"`
}

type WalletsResponse struct {
	WalletA WalletInfo `json:"wallet_a"`
	WalletB WalletInfo `json:"wallet_b"`
}

type ConnectWalletRequest struct {
	WalletType string `json:"wallet_type"` // "wallet_a" | "wallet_b"
	Address    string `json:"address"`
}

type DisconnectWalletRequest struct {
	WalletType string `json:"wallet_type"`
}

type WalletActionResponse struct {
	Success    bool   `json:"success"`
	WalletType string `json:"wallet_type"`
	Address    string `json:"address,omitempty"`
	Message    string `json:"message"`
}

type ExposureInfo struct {
	Value      float64 `json:"value"`
	Unit       string  `json:"unit"`
	Percentage int     `json:"percentage"`
}

type DeltaSync struct {
	PoolExposure  ExposureInfo `json:"pool_exposure"`
	HedgeExposure ExposureInfo `json:"hedge_exposure"`
	NetExposure   float64      `json:"net_exposure"`
	Status        string       `json:"status"` // "neutral", etc.
	IsLive        bool         `json:"is_live"`
	LastSync      time.Time    `json:"last_sync"`
}

type PermissionDetail struct {
	Required string `json:"required"`
	Type     string `json:"type"`
}

type HedgePermission struct {
	Action  string           `json:"action"`
	WalletA PermissionDetail `json:"wallet_a"`
	WalletB PermissionDetail `json:"wallet_b"`
}

type PermissionsResponse struct {
	Permissions []HedgePermission `json:"permissions"`
}

type SafeModeStatus struct {
	IsActive      bool       `json:"is_active"`
	TriggerReason *string    `json:"trigger_reason"`
	ActivatedAt   *time.Time `json:"activated_at"`
	Message       string     `json:"message"`
}

type SyncFlowStep struct {
	Step        int    `json:"step"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
}

type SyncFlow struct {
	CurrentStep int            `json:"current_step"`
	Steps       []SyncFlowStep `json:"steps"`
	Price       float64        `json:"price"`
	EthUnit     string         `json:"eth_unit"`
}
