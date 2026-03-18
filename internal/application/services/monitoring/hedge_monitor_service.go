package monitoring

import (
	"context"
	"log"
	"time"

	"github.com/ivan-salazar14/markerTradeIa/internal/application/usecases/hedge"
	"github.com/ivan-salazar14/markerTradeIa/internal/domain/ports/out"
)

type HedgeMonitorService struct {
	hyperliquidPort out.HyperliquidPort
	walletSync      *hedge.WalletSyncUseCase
	asset           string
	walletAddress   string
	hlAddress       string
}

func NewHedgeMonitorService(
	hlPort out.HyperliquidPort,
	walletSync *hedge.WalletSyncUseCase,
	asset string,
	walletAddress string,
	hlAddress string,
) *HedgeMonitorService {
	return &HedgeMonitorService{
		hyperliquidPort: hlPort,
		walletSync:      walletSync,
		asset:           asset,
		walletAddress:   walletAddress,
		hlAddress:       hlAddress,
	}
}

// Start inits the Event-Driven + Polling hybrid monitoring.
func (s *HedgeMonitorService) Start(ctx context.Context) {
	log.Printf("[HedgeMonitor] Iniciando background service de cobertura para %s (hibrido Event-Driven & Polling)", s.asset)

	priceCh := make(chan float64, 100)
	sizeCh := make(chan float64, 100)

	if err := s.hyperliquidPort.SubscribeToMarketUpdates(ctx, s.asset, priceCh); err != nil {
		log.Printf("[HedgeMonitor] Error suscribiendose a Market Updates: %v", err)
	}

	if err := s.hyperliquidPort.SubscribeToUserEvents(ctx, s.hlAddress, sizeCh); err != nil {
		log.Printf("[HedgeMonitor] Error suscribiendose a User Events: %v", err)
	}

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("[HedgeMonitor] Deteniendo servicio...")
			return
		case <-ticker.C:
			log.Println("[HedgeMonitor/Polling] Ejecutando job de reconciliacion...")
			if _, err := s.walletSync.SyncHedge(ctx, s.walletAddress, s.hlAddress, s.asset); err != nil {
				log.Printf("[HedgeMonitor/Polling] Error durante syncHedge: %v", err)
			}
		case price := <-priceCh:
			log.Printf("[HedgeMonitor/Events] Cambio de mercado detectado: %f. Reevaluando delta...", price)
			if _, err := s.walletSync.SyncHedge(ctx, s.walletAddress, s.hlAddress, s.asset); err != nil {
				log.Printf("[HedgeMonitor/Events] Error durante syncHedge por evento de mercado: %v", err)
			}
		case newSize := <-sizeCh:
			log.Printf("[HedgeMonitor/Events] Cambio de posicion detectado: %f. Reevaluando delta...", newSize)
			if _, err := s.walletSync.SyncHedge(ctx, s.walletAddress, s.hlAddress, s.asset); err != nil {
				log.Printf("[HedgeMonitor/Events] Error durante syncHedge por evento de usuario: %v", err)
			}
		}
	}
}
