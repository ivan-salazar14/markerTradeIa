package monitoring

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/ivan-salazar14/markerTradeIa/internal/domain"
	"github.com/ivan-salazar14/markerTradeIa/internal/domain/ports/out"
)

type MonitoringService struct {
	poolMonitor out.PoolMonitor
	networks    []string
	interval    time.Duration

	mu          sync.RWMutex
	lastPools   map[string][]domain.LiquidityPool
	lastUpdated map[string]time.Time
}

func NewMonitoringService(poolMonitor out.PoolMonitor, networks []string, interval time.Duration) *MonitoringService {
	return &MonitoringService{
		poolMonitor: poolMonitor,
		networks:    networks,
		interval:    interval,
		lastPools:   make(map[string][]domain.LiquidityPool),
		lastUpdated: make(map[string]time.Time),
	}
}

func (s *MonitoringService) Start(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	log.Printf("[Monitoring] Iniciando servicio de monitoreo para redes: %v", s.networks)
	s.pollTopPools(ctx)

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

func (s *MonitoringService) GetPools(ctx context.Context, network string, limit int) ([]domain.LiquidityPool, time.Time, error) {
	s.mu.RLock()
	pools, ok := s.lastPools[network]
	lastUpdated := s.lastUpdated[network]
	s.mu.RUnlock()

	if ok && len(pools) > 0 {
		return limitPools(pools, limit), lastUpdated, nil
	}

	fetchedPools, err := s.poolMonitor.GetTopPools(ctx, network, limit)
	if err != nil {
		return nil, time.Time{}, err
	}

	s.mu.Lock()
	s.lastPools[network] = fetchedPools
	s.lastUpdated[network] = time.Now().UTC()
	updated := s.lastUpdated[network]
	s.mu.Unlock()

	return fetchedPools, updated, nil
}

func (s *MonitoringService) pollTopPools(ctx context.Context) {
	for _, network := range s.networks {
		pools, err := s.poolMonitor.GetTopPools(ctx, network, 5)
		if err != nil {
			log.Printf("[Monitoring] Error obteniendo pools para %s: %v", network, err)
			continue
		}

		now := time.Now().UTC()
		s.mu.Lock()
		s.lastPools[network] = pools
		s.lastUpdated[network] = now
		s.mu.Unlock()

		log.Printf("[Monitoring] Top 5 pools en %s:", network)
		for _, p := range pools {
			log.Printf("  - %s/%s (%d bps): TVL $%.2f, Vol $%.2f",
				p.Symbol0, p.Symbol1, p.FeeTier, p.TVLUSD, p.VolumeUSD)
		}
	}
}

func limitPools(pools []domain.LiquidityPool, limit int) []domain.LiquidityPool {
	if limit <= 0 || limit >= len(pools) {
		return pools
	}
	return pools[:limit]
}
