package domain

// WalletBalance details an individual asset balance
type WalletBalance struct {
	Asset  string  `json:"asset"`
	Amount float64 `json:"amount"`
}

// ActivePool details an active LP position
type ActivePool struct {
	PoolID   string  `json:"pool_id"`
	Symbol   string  `json:"symbol"`
	Size     float64 `json:"size"`
	ValueUsd float64 `json:"value_usd"`
}

// WalletData contains balance and pool information for a wallet
type WalletData struct {
	Address     string          `json:"address"`
	Balances    []WalletBalance `json:"balances"`
	ActivePools []ActivePool    `json:"active_pools"`
}
