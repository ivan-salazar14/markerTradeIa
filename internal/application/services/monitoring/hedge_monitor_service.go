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

// Start inits the Event-Driven + Polling hybrid monitoring
func (s *HedgeMonitorService) Start(ctx context.Context) {
	log.Printf("[HedgeMonitor] Iniciando background service de cobertura para %s (Híbrido Event-Driven & Polling)", s.asset)

	// Create channels for WebSocket events
	priceCh := make(chan float64, 100)
	sizeCh := make(chan float64, 100)

	// Subscribirse a Hyperliquid (Push Mode)
	if err := s.hyperliquidPort.SubscribeToMarketUpdates(ctx, s.asset, priceCh); err != nil {
		log.Printf("[HedgeMonitor] Error subscribiéndose a Market Updates: %v", err)
	}

	if err := s.hyperliquidPort.SubscribeToUserEvents(ctx, s.hlAddress, sizeCh); err != nil {
		log.Printf("[HedgeMonitor] Error subscribiéndose a User Events: %v", err)
	}

	// Ticker para Polling de Reconciliación (cada 30 seg)
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("[HedgeMonitor] Deteniendo servicio (Cancelado)...")
			return
			
		case <-ticker.C:
			// Polling (Pull): Forzamos sincronización absóluta por seguridad
			log.Println("[HedgeMonitor/Polling] Ejecutando Job de reconciliación...")
			if err := s.walletSync.SyncHedge(ctx, s.walletAddress, s.hlAddress, s.asset); err != nil {
				log.Printf("[HedgeMonitor/Polling] Error durante syncHedge: %v", err)
			}

		case price := <-priceCh:
			// Real-time Event (Push): Orderbook update detectado
			log.Printf("[HedgeMonitor/Events] Push Event -> Variación L2 Detectada (Precio Simul/Mark: %f). Reevaluando Delta...", price)
			if err := s.walletSync.SyncHedge(ctx, s.walletAddress, s.hlAddress, s.asset); err != nil {
				log.Printf("[HedgeMonitor] Error durante syncHedge por evento: %v", err)
			}

		case newSize := <-sizeCh:
			// Real-time Event (Push): Short Position update detectado
			log.Printf("[HedgeMonitor/Events] Push Event -> Reflejo en cuenta detectado. Nuevo Short: %f. Reevaluando Delta...", newSize)
			if err := s.walletSync.SyncHedge(ctx, s.walletAddress, s.hlAddress, s.asset); err != nil {
				log.Printf("[HedgeMonitor] Error durante syncHedge: %v", err)
			}
		}
	}
}
