package strategies

import (
	"context"
	"math"

	"github.com/ivan-salazar14/markerTradeIa/internal/domain"
)

// BasicDeltaNeutralStrategy implements a simple 1:1 hedging strategy
type BasicDeltaNeutralStrategy struct {
	Tolerance float64 // Drift difference tolerance before triggering a rebalance
}

func NewBasicDeltaNeutralStrategy(tolerance float64) *BasicDeltaNeutralStrategy {
	return &BasicDeltaNeutralStrategy{Tolerance: tolerance}
}

func (s *BasicDeltaNeutralStrategy) Name() string {
	return "Basic Delta Neutral (1:1)"
}

func (s *BasicDeltaNeutralStrategy) Evaluate(ctx context.Context, lpExposure float64, currentShort float64) (domain.HedgeAction, error) {
	// Let's assume lpExposure is strictly our 'WETH' risk exposure. We want to short EXACTLY that amount.
	targetShort := lpExposure 
	
	// Difference between the ideal target short and our current short
	diff := targetShort - currentShort
	
	// If the absolute drift is larger than the tolerance, we readjust
	if math.Abs(diff) > s.Tolerance {
		return domain.HedgeAction{
			ActionType: "ADJUST_SHORT",
			Asset:      "WETH",
			Size:       targetShort,
			Reason:     "Drift exceeded tolerance",
		}, nil
	}
	
	return domain.HedgeAction{
		ActionType: "DO_NOTHING",
	}, nil
}
