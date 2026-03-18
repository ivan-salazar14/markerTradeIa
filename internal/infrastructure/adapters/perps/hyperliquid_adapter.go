package perps

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"

	"github.com/gorilla/websocket"
	"github.com/ivan-salazar14/markerTradeIa/internal/domain/ports/out"
)

// HyperliquidAdapter implements out.HyperliquidPort using real WebSockets 
// for market data and mocked order execution
type HyperliquidAdapter struct {
	clientConnected bool
	apiSecret       string
}

func NewHyperliquidAdapter() out.HyperliquidPort {
	return &HyperliquidAdapter{
		clientConnected: false,
	}
}

func (a *HyperliquidAdapter) Connect(ctx context.Context, privateKey string) error {
	a.clientConnected = true
	a.apiSecret = "MOCKED_" + privateKey
	log.Println("[Hyperliquid] Connected using private key")
	return nil
}

func (a *HyperliquidAdapter) GetBalances(ctx context.Context, address string) (map[string]float64, error) {
	if !a.clientConnected {
		return nil, fmt.Errorf("hyperliquid client not connected")
	}
	return map[string]float64{
		"USDC": 12500.0,
		"WETH": 0.0,
	}, nil
}

func (a *HyperliquidAdapter) GetShortPosition(ctx context.Context, address string, asset string) (float64, error) {
	if !a.clientConnected {
		return 0, fmt.Errorf("hyperliquid client not connected")
	}
	// Returning a mock active short equivalent to 0.5 Asset size.
	if asset == "WETH" {
		return 0.5, nil
	}
	return 0.0, nil
}

func (a *HyperliquidAdapter) PlaceMarketOrder(ctx context.Context, asset string, isBuy bool, size float64) error {
	if !a.clientConnected {
		return fmt.Errorf("hyperliquid client not connected")
	}
	side := "SELL"
	if isBuy {
		side = "BUY"
	}
	log.Printf("[Hyperliquid REST API] Placing %s Market Order for %f of %s", side, size, asset)
	return nil
}

// SubscribeToMarketUpdates connects via Gorilla WebSocket to get real L2 orderbook updates for pricing
func (a *HyperliquidAdapter) SubscribeToMarketUpdates(ctx context.Context, asset string, priceCh chan<- float64) error {
	u := url.URL{Scheme: "wss", Host: "api.hyperliquid.xyz", Path: "/ws"}
	
	log.Printf("[Hyperliquid WS] Connecting to %s for Market Updates (%s)...", u.String(), asset)
	c, _, err := websocket.DefaultDialer.DialContext(ctx, u.String(), nil)
	if err != nil {
		return fmt.Errorf("dial error: %w", err)
	}

	subscribeMsg := map[string]interface{}{
		"method": "subscribe",
		"subscription": map[string]interface{}{
			"type": "l2Book",
			"coin": asset,
		},
	}

	if err := c.WriteJSON(subscribeMsg); err != nil {
		c.Close()
		return fmt.Errorf("subscribe error: %w", err)
	}

	go func() {
		defer c.Close()
		for {
			select {
			case <-ctx.Done():
				log.Println("[Hyperliquid WS] Stopping Market Updates subscription")
				return
			default:
				_, message, err := c.ReadMessage()
				if err != nil {
					log.Printf("[Hyperliquid WS] Read Error: %v", err)
					return // End loop on close/read error
				}
				
				var result map[string]interface{}
				if err := json.Unmarshal(message, &result); err != nil {
					continue
				}

				if channel, ok := result["channel"].(string); ok && channel == "l2Book" {
					// We've received a real update!
					// In a full implementation, we'd parse `data.levels` to calculate the Mark price properly.
					// For demonstration, we simply push a signal/price to the channel to trigger standard polling/logic.
					log.Printf("[Hyperliquid WS] Real-time orderbook update received for %s", asset)
					select {
					case priceCh <- 2055.00: // Mock parsed price
					default:
						// Skip if channel is full to prevent block
					}
				} else if channel, ok := result["channel"].(string); ok && channel == "subscriptionResponse" {
					log.Printf("[Hyperliquid WS] Successfully subscribed to L2Book for %s", asset)
				}
			}
		}
	}()

	return nil
}

// SubscribeToUserEvents connects via WebSocket to listen for trade executions, fills, or position changes
func (a *HyperliquidAdapter) SubscribeToUserEvents(ctx context.Context, address string, sizeCh chan<- float64) error {
	u := url.URL{Scheme: "wss", Host: "api.hyperliquid.xyz", Path: "/ws"}
	
	log.Printf("[Hyperliquid WS] Connecting for User Events (Wallet: %s)...", address)
	c, _, err := websocket.DefaultDialer.DialContext(ctx, u.String(), nil)
	if err != nil {
		return fmt.Errorf("dial error: %w", err)
	}

	subscribeMsg := map[string]interface{}{
		"method": "subscribe",
		"subscription": map[string]interface{}{
			"type": "userEvents",
			"user": address,
		},
	}

	if err := c.WriteJSON(subscribeMsg); err != nil {
		c.Close()
		return fmt.Errorf("subscribe userEvents error: %w", err)
	}

	go func() {
		defer c.Close()
		for {
			select {
			case <-ctx.Done():
				log.Println("[Hyperliquid WS] Stopping User Events subscription")
				return
			default:
				_, message, err := c.ReadMessage()
				if err != nil {
					log.Printf("[Hyperliquid WS] UserEvents Read Error: %v", err)
					return
				}
				
				var result map[string]interface{}
				if err := json.Unmarshal(message, &result); err != nil {
					continue
				}

				if channel, ok := result["channel"].(string); ok && channel == "userEvents" {
					log.Printf("[Hyperliquid WS] User event state change detected for %s!", address)
					
					// Typically we would parse `data.fills` to figure out position delta, 
					// or `data.clearinghouseState` to just read the current short absolute size.
					// We're pushing a size signaling a change required.
					select {
					case sizeCh <- 0.5: // Mock short size from event
					default:
					}
				} else if channel, ok := result["channel"].(string); ok && channel == "subscriptionResponse" {
					log.Printf("[Hyperliquid WS] Successfully subscribed to User Events for wallet %s", address)
				}
			}
		}
	}()

	return nil
}
