package mappers

import (
	"github.com/ivan-salazar14/markerTradeIa/internal/domain"
	"github.com/ivan-salazar14/markerTradeIa/internal/infrastructure/adapters/repository/tradeAdapter/models"
)

func TradeDomainToModel(trade domain.TradeExecution, userId string) models.Trade {
	return models.Trade{
		ExecutionID: trade.ExecutionID,
		SignalID:    trade.SignalID,
		Status:      string(trade.Status),
		ExecutedAt:  trade.ExecutedAt,
		Details:     trade.Details,
		UserID:      userId,
	}
}
