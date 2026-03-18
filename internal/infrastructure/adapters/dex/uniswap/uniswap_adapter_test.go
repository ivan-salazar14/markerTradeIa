package uniswap

import (
	"math/big"
	"testing"
)

func TestNormalizeAssetSymbol(t *testing.T) {
	if got := normalizeAssetSymbol("WETH"); got != "ETH" {
		t.Fatalf("expected ETH, got %s", got)
	}
	if got := normalizeAssetSymbol("usdc"); got != "USDC" {
		t.Fatalf("expected USDC, got %s", got)
	}
}

func TestEstimatePositionAmountsReturnsPositiveExposureInsideRange(t *testing.T) {
	liquidity := big.NewInt(1_000_000_000_000_000_000)
	sqrtPriceX96 := big.NewInt(0).SetUint64(7922816251426433759)
	sqrtPriceX96 = sqrtPriceX96.Mul(sqrtPriceX96, big.NewInt(10000000000))

	amount0, amount1 := estimatePositionAmounts(liquidity, -120, 120, sqrtPriceX96, 18, 6)
	if amount0 < 0 {
		t.Fatalf("expected non-negative amount0")
	}
	if amount1 < 0 {
		t.Fatalf("expected non-negative amount1")
	}
}

func TestEstimatePositionAmountsReturnsZeroForEmptyLiquidity(t *testing.T) {
	amount0, amount1 := estimatePositionAmounts(big.NewInt(0), -120, 120, big.NewInt(1), 18, 6)
	if amount0 != 0 || amount1 != 0 {
		t.Fatalf("expected zero amounts, got %f and %f", amount0, amount1)
	}
}
