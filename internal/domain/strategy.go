package domain

import (
	"context"
)

// HedgeAction represents the action to take after evaluating a hedging strategy
type HedgeAction struct {
	ActionType string  // e.g., "ADJUST_SHORT", "DO_NOTHING"
	Asset      string  // e.g., "WETH"
	Size       float64 // The new size for the short or the size to adjust
	Reason     string
}

// IHedgeStrategy defines the Strategy pattern interface for hedging
type IHedgeStrategy interface {
	// Evaluate analyzes the LP's exposure and current short, returning an action to take
	Evaluate(ctx context.Context, lpExposure float64, currentShort float64) (HedgeAction, error)
	Name() string
}
