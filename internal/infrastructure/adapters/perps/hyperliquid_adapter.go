package perps

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/ivan-salazar14/markerTradeIa/internal/domain/ports/out"
)

const (
	hyperliquidInfoURL = "https://api.hyperliquid.xyz/info"
	hyperliquidWSURL   = "wss://api.hyperliquid.xyz/ws"
)

// HyperliquidAdapter implements out.HyperliquidPort using Hyperliquid's public info API
// and websocket subscriptions. Order placement remains intentionally blocked until the
// signing flow is implemented.
type HyperliquidAdapter struct {
	clientConnected bool
	apiSecret       string
	httpClient      *http.Client
	wsDialer        *websocket.Dialer
}

func NewHyperliquidAdapter() out.HyperliquidPort {
	return &HyperliquidAdapter{
		httpClient: &http.Client{Timeout: 15 * time.Second},
		wsDialer:   websocket.DefaultDialer,
	}
}

func (a *HyperliquidAdapter) Connect(ctx context.Context, privateKey string) error {
	privateKey = strings.TrimSpace(privateKey)

	// If no private key is provided, we can still connect for reading public data
	// (balances, positions, market prices, user events)
	// Order execution will fail if no key is provided
	if privateKey == "" {
		log.Println("[Hyperliquid] Adapter connected in READ-ONLY mode (no private key)")
	} else {
		a.apiSecret = privateKey
		log.Println("[Hyperliquid] Adapter connected with private key (order execution enabled)")
	}
	a.clientConnected = true
	return nil
}

func (a *HyperliquidAdapter) GetBalances(ctx context.Context, address string) (map[string]float64, error) {
	// Validate address format before making API call
	address = strings.TrimSpace(address)
	if !strings.HasPrefix(address, "0x") || len(address) != 42 {
		return nil, fmt.Errorf("invalid hyperliquid address format: %s (expected 0x... format with 42 characters)", address)
	}

	state, err := a.fetchClearinghouseState(ctx, address)
	if err != nil {
		return nil, err
	}

	balances := map[string]float64{}
	if value, ok := parseFloat(state.MarginSummary.AccountValue); ok {
		balances["ACCOUNT_VALUE"] = value
	}
	if value, ok := parseFloat(state.MarginSummary.TotalMarginUsed); ok {
		balances["TOTAL_MARGIN_USED"] = value
	}
	// Try MarginSummary.Withdrawable first, then fallback to top-level Withdrawable
	if value, ok := parseFloat(state.MarginSummary.Withdrawable); ok {
		balances["WITHDRAWABLE"] = value
	} else if value, ok := parseFloat(state.Withdrawable); ok {
		balances["WITHDRAWABLE"] = value
	}

	for _, balance := range state.AssetPositions {
		coin := strings.TrimSpace(balance.Position.Coin)
		if coin == "" {
			continue
		}
		if size, ok := parseFloat(balance.Position.Szi); ok {
			balances[coin] = size
		}
	}

	return balances, nil
}

func (a *HyperliquidAdapter) GetShortPosition(ctx context.Context, address string, asset string) (float64, error) {
	// Validate address format before making API call
	address = strings.TrimSpace(address)
	if !strings.HasPrefix(address, "0x") || len(address) != 42 {
		return 0, fmt.Errorf("invalid hyperliquid address format: %s (expected 0x... format with 42 characters)", address)
	}

	state, err := a.fetchClearinghouseState(ctx, address)
	if err != nil {
		return 0, err
	}

	for _, position := range state.AssetPositions {
		coin := strings.TrimSpace(position.Position.Coin)
		if !strings.EqualFold(coin, asset) {
			continue
		}
		size, ok := parseFloat(position.Position.Szi)
		if !ok {
			return 0, fmt.Errorf("invalid position size for asset %s", asset)
		}
		if size < 0 {
			return math.Abs(size), nil
		}
		return 0, nil
	}

	return 0, nil
}

func (a *HyperliquidAdapter) PlaceMarketOrder(ctx context.Context, asset string, isBuy bool, size float64) error {
	if !a.clientConnected {
		return fmt.Errorf("hyperliquid client not connected")
	}

	if a.apiSecret == "" {
		return fmt.Errorf("hyperliquid private key required for order execution - currently in read-only mode")
	}

	side := "SELL"
	if isBuy {
		side = "BUY"
	}

	// The signing flow for Hyperliquid exchange actions is not implemented yet.
	return fmt.Errorf("hyperliquid order placement not implemented yet for %s %f %s", side, size, asset)
}

func (a *HyperliquidAdapter) SubscribeToMarketUpdates(ctx context.Context, asset string, priceCh chan<- float64) error {
	c, err := a.openWS(ctx)
	if err != nil {
		return err
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
				log.Println("[Hyperliquid WS] Stopping market subscription")
				return
			default:
				_, message, err := c.ReadMessage()
				if err != nil {
					log.Printf("[Hyperliquid WS] Market read error: %v", err)
					return
				}

				price, ok := parseL2BookPrice(message)
				if !ok {
					continue
				}

				select {
				case priceCh <- price:
				default:
				}
			}
		}
	}()

	return nil
}

func (a *HyperliquidAdapter) SubscribeToUserEvents(ctx context.Context, address string, sizeCh chan<- float64) error {
	c, err := a.openWS(ctx)
	if err != nil {
		return err
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
				log.Println("[Hyperliquid WS] Stopping user events subscription")
				return
			default:
				_, message, err := c.ReadMessage()
				if err != nil {
					log.Printf("[Hyperliquid WS] UserEvents read error: %v", err)
					return
				}

				positionSize, ok := parseUserEventPositionSize(message)
				if !ok {
					continue
				}

				select {
				case sizeCh <- positionSize:
				default:
				}
			}
		}
	}()

	return nil
}

func (a *HyperliquidAdapter) fetchClearinghouseState(ctx context.Context, address string) (*clearinghouseStateResponse, error) {
	if !a.clientConnected {
		return nil, fmt.Errorf("hyperliquid client not connected")
	}

	// Validate address format
	address = strings.TrimSpace(address)
	if !strings.HasPrefix(address, "0x") || len(address) != 42 {
		return nil, fmt.Errorf("invalid hyperliquid address format: %s (expected 0x... format with 42 characters)", address)
	}

	log.Printf("[Hyperliquid] Fetching clearinghouse state for address: %s", address)

	payload := map[string]string{
		"type": "clearinghouseState",
		"user": address,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, hyperliquidInfoURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	log.Printf("[Hyperliquid] Sending request to %s with payload: %s", hyperliquidInfoURL, string(body))

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("hyperliquid request failed: %w", err)
	}
	defer resp.Body.Close()

	responseBody, _ := io.ReadAll(resp.Body)
	log.Printf("[Hyperliquid] Response status: %s, body: %s", resp.Status, string(responseBody))

	if resp.StatusCode == 422 {
		return nil, fmt.Errorf("hyperliquid info error: %s - %s (invalid request format or address)", resp.Status, strings.TrimSpace(string(responseBody)))
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("hyperliquid info error: %s - %s", resp.Status, strings.TrimSpace(string(responseBody)))
	}

	var state clearinghouseStateResponse
	if err := json.Unmarshal(responseBody, &state); err != nil {
		return nil, fmt.Errorf("failed to deserialize hyperliquid response: %w - response: %s", err, string(responseBody))
	}

	return &state, nil
}

func (a *HyperliquidAdapter) openWS(ctx context.Context) (*websocket.Conn, error) {
	u, err := url.Parse(hyperliquidWSURL)
	if err != nil {
		return nil, err
	}
	log.Printf("[Hyperliquid WS] Connecting to %s", u.String())
	conn, _, err := a.wsDialer.DialContext(ctx, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("dial error: %w", err)
	}
	return conn, nil
}

type clearinghouseStateResponse struct {
	MarginSummary struct {
		AccountValue    string `json:"accountValue"`
		TotalMarginUsed string `json:"totalMarginUsed"`
		Withdrawable    string `json:"withdrawable"`
	} `json:"marginSummary"`
	Withdrawable   string `json:"withdrawable,omitempty"`
	AssetPositions []struct {
		Position struct {
			Coin string `json:"coin"`
			Szi  string `json:"szi"`
		} `json:"position"`
	} `json:"assetPositions"`
}

func parseL2BookPrice(message []byte) (float64, bool) {
	var payload map[string]interface{}
	if err := json.Unmarshal(message, &payload); err != nil {
		return 0, false
	}

	channel, _ := payload["channel"].(string)
	if channel != "l2Book" {
		return 0, false
	}

	data, ok := payload["data"].(map[string]interface{})
	if !ok {
		return 0, false
	}

	levels, ok := data["levels"].([]interface{})
	if !ok || len(levels) == 0 {
		return 0, false
	}

	firstSide, ok := levels[0].([]interface{})
	if !ok || len(firstSide) == 0 {
		return 0, false
	}

	firstLevel, ok := firstSide[0].(map[string]interface{})
	if !ok {
		return 0, false
	}

	price, ok := parseUnknownFloat(firstLevel["px"])
	if !ok {
		return 0, false
	}
	return price, true
}

func parseUserEventPositionSize(message []byte) (float64, bool) {
	var payload map[string]interface{}
	if err := json.Unmarshal(message, &payload); err != nil {
		return 0, false
	}

	channel, _ := payload["channel"].(string)
	if channel != "userEvents" {
		return 0, false
	}

	data, ok := payload["data"].(map[string]interface{})
	if !ok {
		return 0, false
	}

	if fills, ok := data["fills"].([]interface{}); ok && len(fills) > 0 {
		lastFill, ok := fills[len(fills)-1].(map[string]interface{})
		if !ok {
			return 0, false
		}
		if size, ok := parseUnknownFloat(lastFill["sz"]); ok {
			return size, true
		}
	}

	if state, ok := data["clearinghouseState"].(map[string]interface{}); ok {
		positions, ok := state["assetPositions"].([]interface{})
		if !ok || len(positions) == 0 {
			return 0, false
		}
		for _, entry := range positions {
			positionEntry, ok := entry.(map[string]interface{})
			if !ok {
				continue
			}
			position, ok := positionEntry["position"].(map[string]interface{})
			if !ok {
				continue
			}
			if size, ok := parseUnknownFloat(position["szi"]); ok {
				return math.Abs(size), true
			}
		}
	}

	return 0, false
}

func parseUnknownFloat(value interface{}) (float64, bool) {
	switch typed := value.(type) {
	case float64:
		return typed, true
	case string:
		return parseFloat(typed)
	case json.Number:
		v, err := typed.Float64()
		return v, err == nil
	default:
		return 0, false
	}
}

func parseFloat(value string) (float64, bool) {
	parsed, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
	if err != nil {
		return 0, false
	}
	return parsed, true
}
