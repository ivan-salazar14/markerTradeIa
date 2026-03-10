package revert

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetTopPools(t *testing.T) {
	// Mock UI Response
	mockResp := gqlResponse{}
	mockResp.Data.Pools = []struct {
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
	}{
		{
			ID: "0x88e6a0c2ddd26feeb64f039a2c41296fcb3f5640",
			Token0: struct {
				Symbol string `json:"symbol"`
			}{Symbol: "WETH"},
			Token1: struct {
				Symbol string `json:"symbol"`
			}{Symbol: "USDC"},
			FeeTier:             "500",
			TotalValueLockedUSD: "1000000.50",
			VolumeUSD:           "500000.25",
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResp)
	}))
	defer server.Close()

	adapter := &RevertAdapter{
		client:      server.Client(),
		baseURL:     server.URL,
		subgraphURL: server.URL,
	}

	pools, err := adapter.GetTopPools(context.Background(), "ethereum", 1)
	assert.NoError(t, err)
	assert.Len(t, pools, 1)
	assert.Equal(t, "0x88e6a0c2ddd26feeb64f039a2c41296fcb3f5640", pools[0].ID)
	assert.Equal(t, "WETH", pools[0].Symbol0)
	assert.Equal(t, "USDC", pools[0].Symbol1)
	assert.Equal(t, 1000000.50, pools[0].TVLUSD)
}

func TestGetPositionStats(t *testing.T) {
	mockResp := struct {
		Positions []struct {
			ID             string  `json:"id"`
			APR            float64 `json:"apr"`
			ROI            float64 `json:"roi"`
			UncollectedFee float64 `json:"uncollected_fee"`
		} `json:"positions"`
	}{
		Positions: []struct {
			ID             string  `json:"id"`
			APR            float64 `json:"apr"`
			ROI            float64 `json:"roi"`
			UncollectedFee float64 `json:"uncollected_fee"`
		}{
			{
				ID:             "12345",
				APR:            15.5,
				ROI:            2.3,
				UncollectedFee: 10.5,
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/positions", r.URL.Path)
		assert.Equal(t, "ethereum", r.URL.Query().Get("network"))
		assert.Equal(t, "12345", r.URL.Query().Get("tokenId"))

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResp)
	}))
	defer server.Close()

	adapter := NewRevertAdapter(server.URL).(*RevertAdapter)
	adapter.client = server.Client() // Use the test client

	stats, err := adapter.GetPositionStats(context.Background(), "ethereum", "12345")
	assert.NoError(t, err)
	assert.Equal(t, "12345", stats.ID)
	assert.Equal(t, 15.5, stats.APR)
	assert.Equal(t, 2.3, stats.ROI)
	assert.Equal(t, 10.5, stats.UncollectedFee)
}
