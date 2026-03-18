package tradeAdapter

import (
	"context"
	"testing"

	"github.com/ivan-salazar14/markerTradeIa/internal/domain"
)

func TestSaveTradeExecutionReturnsErrorWithoutDatabase(t *testing.T) {
	repo := &TradeRepository{}
	trade := domain.TradeExecution{
		ExecutionID: "exec-1",
		SignalID:    "sig-1",
		Status:      "SUCCESS",
		Details:     "details",
	}

	err := repo.SaveTradeExecution(context.Background(), trade)
	if err == nil {
		t.Fatalf("expected error when database is not initialized")
	}
}
