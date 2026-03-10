package controllers

import (
	"net/http"

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
		network = "ethereum"
	}

	// Assuming GetTopPools can be called from service (added for HTTP export)
	// In the real implementation we would get from the last cache update
	// or call the adapter directly through the service.
	// For now let's just return a placeholder or implement in service.
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status": "pool monitoring is active for ` + network + `"}`))
}
