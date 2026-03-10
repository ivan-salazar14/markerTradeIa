package domain

import "time"

// LiquidityPool representa un pool de liquidez en Uniswap (u otros AMMs).
type LiquidityPool struct {
	ID        string    `json:"id"`
	Network   string    `json:"network"`
	Protocol  string    `json:"protocol"` // e.g., "uniswap_v3"
	Symbol0   string    `json:"symbol0"`
	Symbol1   string    `json:"symbol1"`
	FeeTier   int       `json:"fee_tier"`
	TVLUSD    float64   `json:"tvl_usd"`
	VolumeUSD float64   `json:"volume_usd"`
	APR       float64   `json:"apr"`
	UpdatedAt time.Time `json:"updated_at"`
}

// PositionStats representa las estadísticas de una posición específica de LP.
type PositionStats struct {
	ID             string    `json:"id"`
	PoolID         string    `json:"pool_id"`
	Network        string    `json:"network"`
	Owner          string    `json:"owner"`
	Liquidity      string    `json:"liquidity"`
	UncollectedFee float64   `json:"uncollected_fee"`
	APR            float64   `json:"apr"`
	ROI            float64   `json:"roi"`
	UpdatedAt      time.Time `json:"updated_at"`
}
