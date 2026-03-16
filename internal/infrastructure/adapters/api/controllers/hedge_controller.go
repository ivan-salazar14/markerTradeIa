package controllers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/ivan-salazar14/markerTradeIa/internal/domain"
)

type HedgeController struct {
	// En el futuro inyectaremos el servicio aquí
}

func NewHedgeController() *HedgeController {
	return &HedgeController{}
}

func (c *HedgeController) GetStrategy(w http.ResponseWriter, r *http.Request) {
	strategy := domain.HedgeStrategy{
		StrategyID:  "TW-8892",
		Name:        "Estrategia Dual-Wallet (LP + Hedge)",
		Description: "Separe su capital de inversión del capital de cobertura con seguridad aislada.",
		Status:      "active",
		IsSynced:    true,
		CreatedAt:   time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
	}
	c.json(w, strategy)
}

func (c *HedgeController) GetStats(w http.ResponseWriter, r *http.Request) {
	stats := domain.HedgeStats{
		APR: domain.TrendStats{
			Value:          113.82,
			Trend:          12.4,
			TrendDirection: "up",
		},
		FeesAccumulated: domain.CurrencyStats{
			Value:          10.71,
			Currency:       "USD",
			Trend:          5.2,
			TrendDirection: "up",
		},
		Delta: domain.UnitStats{
			Value: 0.0000,
			Unit:  "ETH",
		},
		HedgeEfficiency: domain.TrendStats{
			Value:          99.8,
			Trend:          0.2,
			TrendDirection: "up",
		},
	}
	c.json(w, stats)
}

func (c *HedgeController) GetWallets(w http.ResponseWriter, r *http.Request) {
	addrA := "0x71C7656ec7ab88b098defB751B7401B5f6d4921"
	addrShort := "0x71C...4921"
	
	resp := domain.WalletsResponse{
		WalletA: domain.WalletInfo{
			Type:            "uniswap_lp",
			Name:            "Wallet A (Uniswap LP)",
			Description:     "Gestiona su posición en Uniswap V3 Arbitrum via Permit2.",
			Connected:       true,
			Address:         &addrShort,
			FullAddress:     &addrA,
			Permissions:     []string{"add_liquidity", "collect_fees"},
			PermissionsNote: "Permisos de solo gestión de liquidez aprobados",
		},
		WalletB: domain.WalletInfo{
			Type:            "hyperliquid_trade",
			Name:            "Wallet B (Hyperliquid Trade)",
			Description:     "Dedicada a la cobertura (Short). Utiliza un Signing Agent.",
			Connected:       false,
			Address:         nil,
			FullAddress:     nil,
			Permissions:     []string{"adjust_short"},
			PermissionsNote: "El bot no tiene permisos de retiro",
		},
	}
	c.json(w, resp)
}

func (c *HedgeController) ConnectWallet(w http.ResponseWriter, r *http.Request) {
	var req domain.ConnectWalletRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	resp := domain.WalletActionResponse{
		Success:    true,
		WalletType: req.WalletType,
		Address:    req.Address,
		Message:    "Wallet conectada exitosamente",
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

func (c *HedgeController) GetDelta(w http.ResponseWriter, r *http.Request) {
	delta := domain.DeltaSync{
		PoolExposure: domain.ExposureInfo{
			Value:      0.2522,
			Unit:       "WETH",
			Percentage: 65,
		},
		HedgeExposure: domain.ExposureInfo{
			Value:      -0.2522,
			Unit:       "WETH",
			Percentage: 65,
		},
		NetExposure: 0.0000,
		Status:      "neutral",
		IsLive:      true,
		LastSync:    time.Now(),
	}
	c.json(w, delta)
}

func (c *HedgeController) GetPermissions(w http.ResponseWriter, r *http.Request) {
	resp := domain.PermissionsResponse{
		Permissions: []domain.HedgePermission{
			{
				Action: "Añadir Liquidez",
				WalletA: domain.PermissionDetail{
					Required: "Firma Requerida",
					Type:     "user_signature",
				},
				WalletB: domain.PermissionDetail{
					Required: "No aplica",
					Type:     "not_applicable",
				},
			},
			{
				Action: "Ajustar Short",
				WalletA: domain.PermissionDetail{
					Required: "No aplica",
					Type:     "not_applicable",
				},
				WalletB: domain.PermissionDetail{
					Required: "Automático (Agente)",
					Type:     "agent_authorized",
				},
			},
			{
				Action: "Retirar Capital",
				WalletA: domain.PermissionDetail{
					Required: "Solo Usuario",
					Type:     "user_only",
				},
				WalletB: domain.PermissionDetail{
					Required: "Solo Usuario",
					Type:     "user_only",
				},
			},
			{
				Action: "Cobrar Fees",
				WalletA: domain.PermissionDetail{
					Required: "Bot (Autorizado)",
					Type:     "bot_authorized",
				},
				WalletB: domain.PermissionDetail{
					Required: "No aplica",
					Type:     "not_applicable",
				},
			},
		},
	}
	c.json(w, resp)
}

func (c *HedgeController) GetSafeMode(w http.ResponseWriter, r *http.Request) {
	status := domain.SafeModeStatus{
		IsActive:      false,
		TriggerReason: nil,
		ActivatedAt:   nil,
		Message:       "Sistema operando normalmente",
	}
	c.json(w, status)
}

func (c *HedgeController) GetSyncFlow(w http.ResponseWriter, r *http.Request) {
	flow := domain.SyncFlow{
		CurrentStep: 6,
		Steps: []domain.SyncFlowStep{
			{Step: 1, Title: "Monitoreo Constante", Description: "WebSockets detectan evento de nuevo precio ETH ($2,055).", Icon: "🌐"},
			{Step: 2, Title: "Cálculo de Delta", Description: "Watcher service calcula WETH en Wallet A (0.2522).", Icon: "🧮"},
			{Step: 3, Title: "Verificación de Cobertura", Description: "Consulta posición Short actual en Wallet B (-0.2450).", Icon: "🛡️"},
			{Step: 4, Title: "Detección de Drift", Description: "Diferencia detectada de 0.0072 ETH.", Icon: "⚠️"},
			{Step: 5, Title: "Ajuste de Cobertura", Description: "Ejecución Market Short en Wallet B por 0.0072 ETH.", Icon: "⚡"},
			{Step: 6, Title: "Notificación", Description: "Hedge Sincronizado (Delta 0.0) actualizado en UI.", Icon: "✅"},
		},
		Price:   2055.00,
		EthUnit: "USD",
	}
	c.json(w, flow)
}

func (c *HedgeController) json(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
