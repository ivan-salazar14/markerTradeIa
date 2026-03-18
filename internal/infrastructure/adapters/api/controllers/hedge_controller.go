package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ivan-salazar14/markerTradeIa/internal/application/usecases/hedge"
	"github.com/ivan-salazar14/markerTradeIa/internal/domain"
)

type HedgeController struct {
	walletSyncUseCase *hedge.WalletSyncUseCase
	defaultAsset      string
	defaultWallet     string
	defaultHLWallet   string
	safeMode          bool
}

func NewHedgeController(walletSyncUseCase *hedge.WalletSyncUseCase, defaultAsset string, defaultWallet string, defaultHLWallet string, safeMode bool) *HedgeController {
	return &HedgeController{
		walletSyncUseCase: walletSyncUseCase,
		defaultAsset:      defaultAsset,
		defaultWallet:     defaultWallet,
		defaultHLWallet:   defaultHLWallet,
		safeMode:          safeMode,
	}
}

func (c *HedgeController) GetStrategy(w http.ResponseWriter, r *http.Request) {
	strategy := domain.HedgeStrategy{
		StrategyID:  "delta-neutral-mvp",
		Name:        "Estrategia Delta Neutral",
		Description: "Sincroniza la exposicion LP con una cobertura short en Hyperliquid.",
		Status:      "active",
		IsSynced:    true,
		CreatedAt:   time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
	}
	c.json(w, strategy)
}

func (c *HedgeController) GetStats(w http.ResponseWriter, r *http.Request) {
	result, err := c.walletSyncUseCase.GetLatestDelta(r.Context(), c.assetFromRequest(r))
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to load stats: %v", err), http.StatusInternalServerError)
		return
	}

	stats := domain.HedgeStats{
		APR:             domain.TrendStats{Value: 0, Trend: 0, TrendDirection: "flat"},
		FeesAccumulated: domain.CurrencyStats{Value: 0, Currency: "USD", Trend: 0, TrendDirection: "flat"},
		Delta:           domain.UnitStats{Value: 0, Unit: c.assetFromRequest(r)},
		HedgeEfficiency: domain.TrendStats{Value: 0, Trend: 0, TrendDirection: "flat"},
	}
	if result != nil {
		stats.Delta.Value = result.NetExposure
		if result.Status == "synced" {
			stats.HedgeEfficiency.Value = 100
		}
	}
	c.json(w, stats)
}

func (c *HedgeController) GetWallets(w http.ResponseWriter, r *http.Request) {
	wallets, err := c.walletSyncUseCase.GetWalletConnections(r.Context())
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to load wallets: %v", err), http.StatusInternalServerError)
		return
	}
	c.json(w, wallets)
}

func (c *HedgeController) ConnectWallet(w http.ResponseWriter, r *http.Request) {
	var req domain.ConnectWalletRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	walletData, err := c.walletSyncUseCase.ConnectAndFetchWallet(r.Context(), req.Address)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to connect wallet: %v", err), http.StatusInternalServerError)
		return
	}

	if err := c.walletSyncUseCase.RegisterWallet(r.Context(), req.WalletType, req.Address); err != nil {
		http.Error(w, fmt.Sprintf("failed to register wallet: %v", err), http.StatusInternalServerError)
		return
	}

	resp := struct {
		domain.WalletActionResponse
		Data domain.WalletData `json:"data"`
	}{
		WalletActionResponse: domain.WalletActionResponse{
			Success:    true,
			WalletType: req.WalletType,
			Address:    req.Address,
			Message:    "Wallet conectada y sincronizada exitosamente",
		},
		Data: walletData,
	}

	c.json(w, resp)
}

func (c *HedgeController) DisconnectWallet(w http.ResponseWriter, r *http.Request) {
	var req domain.DisconnectWalletRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	resp := domain.WalletActionResponse{
		Success:    true,
		WalletType: req.WalletType,
		Message:    "Wallet desconectada",
	}
	c.json(w, resp)
}

func (c *HedgeController) SyncNow(w http.ResponseWriter, r *http.Request) {
	var req domain.ManualSyncRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	asset := strings.TrimSpace(req.Asset)
	if asset == "" {
		asset = c.defaultAsset
	}
	walletAddress := strings.TrimSpace(req.WalletAddress)
	if walletAddress == "" {
		walletAddress = c.defaultWallet
	}
	hlAddress := strings.TrimSpace(req.HyperliquidAddress)
	if hlAddress == "" {
		hlAddress = c.defaultHLWallet
	}

	result, err := c.walletSyncUseCase.SyncHedge(r.Context(), walletAddress, hlAddress, asset)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		c.json(w, result)
		return
	}
	c.json(w, result)
}

func (c *HedgeController) GetDelta(w http.ResponseWriter, r *http.Request) {
	result, err := c.walletSyncUseCase.GetLatestDelta(r.Context(), c.assetFromRequest(r))
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to load delta: %v", err), http.StatusInternalServerError)
		return
	}
	if result == nil {
		w.WriteHeader(http.StatusNotFound)
		c.json(w, map[string]string{"error": "no hedge state available yet"})
		return
	}

	status := "neutral"
	if result.Status == "error" {
		status = "error"
	}
	if result.NetExposure > 0 {
		status = "long_bias"
	}
	if result.NetExposure < 0 {
		status = "short_bias"
	}
	if result.Status == "synced" {
		status = "neutral"
	}

	delta := domain.DeltaSync{
		PoolExposure: domain.ExposureInfo{
			Value: result.PoolExposure,
			Unit:  result.Asset,
		},
		HedgeExposure: domain.ExposureInfo{
			Value: -result.ShortExposure,
			Unit:  result.Asset,
		},
		NetExposure: result.NetExposure,
		Status:      status,
		IsLive:      result.Status != "error",
		LastSync:    result.LastSync,
	}
	c.json(w, delta)
}

func (c *HedgeController) GetPermissions(w http.ResponseWriter, r *http.Request) {
	resp := domain.PermissionsResponse{
		Permissions: []domain.HedgePermission{
			{
				Action:  "Ajustar Short",
				WalletA: domain.PermissionDetail{Required: "No aplica", Type: "not_applicable"},
				WalletB: domain.PermissionDetail{Required: "Automatizado", Type: "agent_authorized"},
			},
		},
	}
	c.json(w, resp)
}

func (c *HedgeController) GetSafeMode(w http.ResponseWriter, r *http.Request) {
	status := domain.SafeModeStatus{
		IsActive:      c.safeMode,
		TriggerReason: nil,
		ActivatedAt:   nil,
		Message:       "Safe mode configurado desde variables de entorno",
	}
	c.json(w, status)
}

func (c *HedgeController) GetSyncFlow(w http.ResponseWriter, r *http.Request) {
	flow := domain.SyncFlow{
		CurrentStep: 4,
		Steps: []domain.SyncFlowStep{
			{Step: 1, Title: "Lectura LP", Description: "Lee exposicion LP de la wallet conectada.", Icon: "1"},
			{Step: 2, Title: "Lectura Short", Description: "Consulta la posicion short actual en Hyperliquid.", Icon: "2"},
			{Step: 3, Title: "Evaluacion", Description: "La estrategia decide si hace falta ajustar la cobertura.", Icon: "3"},
			{Step: 4, Title: "Persistencia", Description: "Guarda el resultado del sync y lo expone por API.", Icon: "4"},
		},
		Price:   0,
		EthUnit: "USD",
	}
	c.json(w, flow)
}

func (c *HedgeController) assetFromRequest(r *http.Request) string {
	asset := strings.TrimSpace(r.URL.Query().Get("asset"))
	if asset == "" {
		return c.defaultAsset
	}
	return asset
}

func (c *HedgeController) json(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(data)
}
