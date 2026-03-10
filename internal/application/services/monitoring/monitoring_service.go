package monitoring

import (
	"context"
	"log"
	"time"

	"github.com/ivan-salazar14/markerTradeIa/internal/domain/ports/out"
)

type MonitoringService struct {
	poolMonitor out.PoolMonitor
	networks    []string
	interval    time.Duration
}

func NewMonitoringService(poolMonitor out.PoolMonitor, networks []string, interval time.Duration) *MonitoringService {
	return &MonitoringService{
		poolMonitor: poolMonitor,
		networks:    networks,
		interval:    interval,
	}
}

func (s *MonitoringService) Start(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	log.Printf("[Monitoring] Iniciando servicio de monitoreo para redes: %v", s.networks)

	for {
		select {
		case <-ctx.Done():
			log.Println("[Monitoring] Deteniendo servicio de monitoreo...")
			return
		case <-ticker.C:
			s.pollTopPools(ctx)
		}
	}
}

func (s *MonitoringService) pollTopPools(ctx context.Context) {
	for _, network := range s.networks {
		pools, err := s.poolMonitor.GetTopPools(ctx, network, 5)
		if err != nil {
			log.Printf("[Monitoring] Error obteniendo pools para %s: %v", network, err)
			continue
		}

		log.Printf("[Monitoring] Top 5 pools en %s:", network)
		for _, p := range pools {
			log.Printf("  - %s/%s (%d bps): TVL $%.2f, Vol $%.2f",
				p.Symbol0, p.Symbol1, p.FeeTier, p.TVLUSD, p.VolumeUSD)
		}
	}
}
