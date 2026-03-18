package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/ivan-salazar14/markerTradeIa/internal/application/services/monitoring"
)

type MonitoringController struct {
	service *monitoring.MonitoringService
}

func NewMonitoringController(s *monitoring.MonitoringService) *MonitoringController {
	return &MonitoringController{service: s}
}

func (c *MonitoringController) GetPools(w http.ResponseWriter, r *http.Request) {
	network := r.URL.Query().Get("network")
	if network == "" {
		network = "mainnet"
	}

	limit := 5
	if rawLimit := r.URL.Query().Get("limit"); rawLimit != "" {
		if parsedLimit, err := strconv.Atoi(rawLimit); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	pools, updatedAt, err := c.service.GetPools(r.Context(), network, limit)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to load pools: %v", err), http.StatusBadGateway)
		return
	}

	response := struct {
		Network   string      `json:"network"`
		Count     int         `json:"count"`
		UpdatedAt time.Time   `json:"updated_at"`
		Pools     interface{} `json:"pools"`
	}{
		Network:   network,
		Count:     len(pools),
		UpdatedAt: updatedAt,
		Pools:     pools,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}
