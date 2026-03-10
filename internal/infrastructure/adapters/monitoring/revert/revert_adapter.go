package revert

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ivan-salazar14/markerTradeIa/internal/domain"
	"github.com/ivan-salazar14/markerTradeIa/internal/domain/ports/out"
)

type RevertAdapter struct {
	client      *http.Client
	baseURL     string // https://api.revert.finance/v1
	subgraphURL string // https://api.thegraph.com/subgraphs/name/revert-finance/uniswap-v3
}

func NewRevertAdapter(baseURL string) out.PoolMonitor {
	return &RevertAdapter{
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
		baseURL:     baseURL,
		subgraphURL: "https://api.thegraph.com/subgraphs/name/revert-finance/uniswap-v3",
	}
}

type gqlRequest struct {
	Query string `json:"query"`
}

type gqlResponse struct {
	Data struct {
		Pools []struct {
			ID     string `json:"id"`
			Token0 struct {
				Symbol string `json:"symbol"`
			} `json:"token0"`
			Token1 struct {
				Symbol string `json:"symbol"`
			} `json:"token1"`
			FeeTier             string `json:"feeTier"`
			TotalValueLockedUSD string `json:"totalValueLockedUSD"`
			VolumeUSD           string `json:"volumeUSD"`
		} `json:"pools"`
	} `json:"data"`
}

func (a *RevertAdapter) GetTopPools(ctx context.Context, network string, limit int) ([]domain.LiquidityPool, error) {
	subgraphURL := a.subgraphURL
	if !strings.Contains(subgraphURL, "localhost") && !strings.Contains(subgraphURL, "127.0.0.1") {
		subgraphURL = fmt.Sprintf("%s-%s", a.subgraphURL, network)
	}
	query := fmt.Sprintf(`{
		pools(first: %d, orderBy: totalValueLockedUSD, orderDirection: desc) {
			id
			token0 { symbol }
			token1 { symbol }
			feeTier
			totalValueLockedUSD
			volumeUSD
		}
	}`, limit)

	reqBody, _ := json.Marshal(gqlRequest{Query: query})
	req, err := http.NewRequestWithContext(ctx, "POST", subgraphURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("subgraph error: %s", resp.Status)
	}

	var gqlResp gqlResponse
	if err := json.NewDecoder(resp.Body).Decode(&gqlResp); err != nil {
		return nil, err
	}

	var result []domain.LiquidityPool
	for _, p := range gqlResp.Data.Pools {
		tvl, _ := strconv.ParseFloat(p.TotalValueLockedUSD, 64)
		vol, _ := strconv.ParseFloat(p.VolumeUSD, 64)
		fee, _ := strconv.Atoi(p.FeeTier)

		result = append(result, domain.LiquidityPool{
			ID:        p.ID,
			Network:   network,
			Protocol:  "uniswap_v3",
			Symbol0:   p.Token0.Symbol,
			Symbol1:   p.Token1.Symbol,
			FeeTier:   fee,
			TVLUSD:    tvl,
			VolumeUSD: vol,
			UpdatedAt: time.Now(),
		})
	}

	return result, nil
}

func (a *RevertAdapter) GetPositionStats(ctx context.Context, network string, positionID string) (domain.PositionStats, error) {
	// Usando la API interna de Revert que vimos en el descubrimiento
	url := fmt.Sprintf("%s/positions?network=%s&tokenId=%s", a.baseURL, network, positionID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return domain.PositionStats{}, err
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return domain.PositionStats{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return domain.PositionStats{}, fmt.Errorf("revert api error: %s", resp.Status)
	}

	// Simulación de respuesta ya que no tenemos el esquema exacto pero sabemos que trae APR y ROI
	body, _ := io.ReadAll(resp.Body)
	var data struct {
		Positions []struct {
			ID             string  `json:"id"`
			APR            float64 `json:"apr"`
			ROI            float64 `json:"roi"`
			UncollectedFee float64 `json:"uncollected_fee"`
		} `json:"positions"`
	}

	if err := json.Unmarshal(body, &data); err != nil {
		return domain.PositionStats{}, err
	}

	if len(data.Positions) == 0 {
		return domain.PositionStats{}, fmt.Errorf("position not found")
	}

	pos := data.Positions[0]
	return domain.PositionStats{
		ID:             pos.ID,
		Network:        network,
		UncollectedFee: pos.UncollectedFee,
		APR:            pos.APR,
		ROI:            pos.ROI,
		UpdatedAt:      time.Now(),
	}, nil
}
