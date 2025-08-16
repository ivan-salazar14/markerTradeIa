package tradeAdapter

import (
	"context"
	"testing"

	"github.com/ivan-salazar14/markerTradeIa/internal/domain"
	"gorm.io/gorm"
)

type mockDB struct {
	gorm.DB
	saveCalled bool
}

func (m *mockDB) Save(value interface{}) *gorm.DB {
	m.saveCalled = true
	return &gorm.DB{}
}

func TestSaveTradeExecution(t *testing.T) {
	repo := &TradeRepository{db: &gorm.DB{}} // You may want to use a mock or stub here
	trade := domain.TradeExecution{
		ExecutionID: "exec-1",
		SignalID:    "sig-1",
		Status:      "SUCCESS",
		Details:     "details",
	}
	err := repo.SaveTradeExecution(context.Background(), trade)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}
