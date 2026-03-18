package uniswap

import (
	"context"
	"fmt"
	"log"
	"math"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ivan-salazar14/markerTradeIa/internal/domain"
	"github.com/ivan-salazar14/markerTradeIa/internal/domain/ports/out"
)

const (
	uniswapV3FactoryAddress = "0x1F98431c8aD98523631AE4a59f267346ea31F984"
	q96Float                = 79228162514264337593543950336.0
)

const positionManagerABI = `[
	{"inputs":[{"internalType":"address","name":"owner","type":"address"}],"name":"balanceOf","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},
	{"inputs":[{"internalType":"address","name":"owner","type":"address"},{"internalType":"uint256","name":"index","type":"uint256"}],"name":"tokenOfOwnerByIndex","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},
	{"inputs":[{"internalType":"uint256","name":"tokenId","type":"uint256"}],"name":"positions","outputs":[{"internalType":"uint96","name":"nonce","type":"uint96"},{"internalType":"address","name":"operator","type":"address"},{"internalType":"address","name":"token0","type":"address"},{"internalType":"address","name":"token1","type":"address"},{"internalType":"uint24","name":"fee","type":"uint24"},{"internalType":"int24","name":"tickLower","type":"int24"},{"internalType":"int24","name":"tickUpper","type":"int24"},{"internalType":"uint128","name":"liquidity","type":"uint128"},{"internalType":"uint256","name":"feeGrowthInside0LastX128","type":"uint256"},{"internalType":"uint256","name":"feeGrowthInside1LastX128","type":"uint256"},{"internalType":"uint128","name":"tokensOwed0","type":"uint128"},{"internalType":"uint128","name":"tokensOwed1","type":"uint128"}],"stateMutability":"view","type":"function"}
]`

const erc20ABI = `[
	{"inputs":[],"name":"symbol","outputs":[{"internalType":"string","name":"","type":"string"}],"stateMutability":"view","type":"function"},
	{"inputs":[],"name":"decimals","outputs":[{"internalType":"uint8","name":"","type":"uint8"}],"stateMutability":"view","type":"function"}
]`

const factoryABI = `[
	{"inputs":[{"internalType":"address","name":"tokenA","type":"address"},{"internalType":"address","name":"tokenB","type":"address"},{"internalType":"uint24","name":"fee","type":"uint24"}],"name":"getPool","outputs":[{"internalType":"address","name":"pool","type":"address"}],"stateMutability":"view","type":"function"}
]`

const poolABI = `[
	{"inputs":[],"name":"slot0","outputs":[{"internalType":"uint160","name":"sqrtPriceX96","type":"uint160"},{"internalType":"int24","name":"tick","type":"int24"},{"internalType":"uint16","name":"observationIndex","type":"uint16"},{"internalType":"uint16","name":"observationCardinality","type":"uint16"},{"internalType":"uint16","name":"observationCardinalityNext","type":"uint16"},{"internalType":"uint8","name":"feeProtocol","type":"uint8"},{"internalType":"bool","name":"unlocked","type":"bool"}],"stateMutability":"view","type":"function"}
]`

type UniswapV3WalletAdapter struct {
	client          *ethclient.Client
	positionManager common.Address
	factory         common.Address
	pmABI           abi.ABI
	erc20ABI        abi.ABI
	factoryABI      abi.ABI
	poolABI         abi.ABI
}

type positionDetails struct {
	TokenID     *big.Int
	Token0      common.Address
	Token1      common.Address
	Fee         uint32
	TickLower   int32
	TickUpper   int32
	Liquidity   *big.Int
	TokensOwed0 *big.Int
	TokensOwed1 *big.Int
}

type tokenMetadata struct {
	Symbol   string
	Decimals int
}

func NewUniswapV3WalletAdapter(rpcURL string, positionManagerAddr string) (out.WalletPort, error) {
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("failed to dial RPC %s: %w", rpcURL, err)
	}

	pmABI, err := abi.JSON(strings.NewReader(positionManagerABI))
	if err != nil {
		return nil, fmt.Errorf("failed to parse position manager ABI: %w", err)
	}
	erc20ABIParsed, err := abi.JSON(strings.NewReader(erc20ABI))
	if err != nil {
		return nil, fmt.Errorf("failed to parse erc20 ABI: %w", err)
	}
	factoryABIParsed, err := abi.JSON(strings.NewReader(factoryABI))
	if err != nil {
		return nil, fmt.Errorf("failed to parse factory ABI: %w", err)
	}
	poolABIParsed, err := abi.JSON(strings.NewReader(poolABI))
	if err != nil {
		return nil, fmt.Errorf("failed to parse pool ABI: %w", err)
	}

	return &UniswapV3WalletAdapter{
		client:          client,
		positionManager: common.HexToAddress(positionManagerAddr),
		factory:         common.HexToAddress(uniswapV3FactoryAddress),
		pmABI:           pmABI,
		erc20ABI:        erc20ABIParsed,
		factoryABI:      factoryABIParsed,
		poolABI:         poolABIParsed,
	}, nil
}

func (a *UniswapV3WalletAdapter) GetBalances(ctx context.Context, address string) (map[string]float64, error) {
	account := common.HexToAddress(address)
	balance, err := a.client.BalanceAt(ctx, account, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get eth balance: %w", err)
	}

	fbal := new(big.Float)
	fbal.SetString(balance.String())
	ethValue := new(big.Float).Quo(fbal, big.NewFloat(1e18))
	val, _ := ethValue.Float64()

	return map[string]float64{
		"ETH": val,
	}, nil
}

func (a *UniswapV3WalletAdapter) GetActivePoolPositions(ctx context.Context, address string) ([]domain.ActivePool, error) {
	log.Printf("[Uniswap] Fetching active pool positions for %s", address)

	tokenIDs, err := a.getOwnedTokenIDs(ctx, common.HexToAddress(address))
	if err != nil {
		return nil, err
	}
	if len(tokenIDs) == 0 {
		return []domain.ActivePool{}, nil
	}

	activePools := make([]domain.ActivePool, 0, len(tokenIDs)*2)
	for _, tokenID := range tokenIDs {
		position, err := a.getPositionDetails(ctx, tokenID)
		if err != nil {
			log.Printf("[Uniswap] failed to load position %s: %v", tokenID.String(), err)
			continue
		}
		if position.Liquidity == nil || position.Liquidity.Sign() == 0 {
			continue
		}

		token0Meta, err := a.getTokenMetadata(ctx, position.Token0)
		if err != nil {
			log.Printf("[Uniswap] failed to load token0 metadata for %s: %v", position.Token0.Hex(), err)
			continue
		}
		token1Meta, err := a.getTokenMetadata(ctx, position.Token1)
		if err != nil {
			log.Printf("[Uniswap] failed to load token1 metadata for %s: %v", position.Token1.Hex(), err)
			continue
		}

		poolAddress, err := a.getPoolAddress(ctx, position.Token0, position.Token1, position.Fee)
		if err != nil {
			log.Printf("[Uniswap] failed to resolve pool for position %s: %v", tokenID.String(), err)
			continue
		}
		if poolAddress == (common.Address{}) {
			log.Printf("[Uniswap] no pool found for position %s", tokenID.String())
			continue
		}

		sqrtPriceX96, err := a.getPoolSqrtPriceX96(ctx, poolAddress)
		if err != nil {
			log.Printf("[Uniswap] failed to read slot0 for pool %s: %v", poolAddress.Hex(), err)
			continue
		}

		amount0, amount1 := estimatePositionAmounts(
			position.Liquidity,
			int(position.TickLower),
			int(position.TickUpper),
			sqrtPriceX96,
			token0Meta.Decimals,
			token1Meta.Decimals,
		)

		basePoolID := fmt.Sprintf("%s:%s", poolAddress.Hex(), tokenID.String())
		if amount0 > 0 {
			activePools = append(activePools, domain.ActivePool{
				PoolID:   basePoolID,
				TokenID:  tokenID.String(),
				Protocol: "uniswap_v3",
				Symbol:   normalizeAssetSymbol(token0Meta.Symbol),
				Size:     amount0,
				ValueUsd: 0,
			})
		}
		if amount1 > 0 {
			activePools = append(activePools, domain.ActivePool{
				PoolID:   basePoolID,
				TokenID:  tokenID.String(),
				Protocol: "uniswap_v3",
				Symbol:   normalizeAssetSymbol(token1Meta.Symbol),
				Size:     amount1,
				ValueUsd: 0,
			})
		}
	}

	return activePools, nil
}

func (a *UniswapV3WalletAdapter) getOwnedTokenIDs(ctx context.Context, owner common.Address) ([]*big.Int, error) {
	output, err := a.callContract(ctx, a.positionManager, a.pmABI, "balanceOf", owner)
	if err != nil {
		return nil, fmt.Errorf("failed to read NFT balance: %w", err)
	}
	results, err := a.pmABI.Unpack("balanceOf", output)
	if err != nil || len(results) == 0 {
		return nil, fmt.Errorf("failed to unpack balanceOf: %w", err)
	}

	balance, ok := results[0].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("unexpected balanceOf type")
	}

	tokenIDs := make([]*big.Int, 0, balance.Int64())
	for i := int64(0); i < balance.Int64(); i++ {
		output, err := a.callContract(ctx, a.positionManager, a.pmABI, "tokenOfOwnerByIndex", owner, big.NewInt(i))
		if err != nil {
			return nil, fmt.Errorf("failed to read tokenOfOwnerByIndex(%d): %w", i, err)
		}
		results, err := a.pmABI.Unpack("tokenOfOwnerByIndex", output)
		if err != nil || len(results) == 0 {
			return nil, fmt.Errorf("failed to unpack tokenOfOwnerByIndex(%d): %w", i, err)
		}
		tokenID, ok := results[0].(*big.Int)
		if !ok {
			return nil, fmt.Errorf("unexpected token id type at index %d", i)
		}
		tokenIDs = append(tokenIDs, tokenID)
	}

	return tokenIDs, nil
}

func (a *UniswapV3WalletAdapter) getPositionDetails(ctx context.Context, tokenID *big.Int) (*positionDetails, error) {
	output, err := a.callContract(ctx, a.positionManager, a.pmABI, "positions", tokenID)
	if err != nil {
		return nil, err
	}
	results, err := a.pmABI.Unpack("positions", output)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack positions(%s): %w", tokenID.String(), err)
	}
	if len(results) < 12 {
		return nil, fmt.Errorf("unexpected positions result length for token %s", tokenID.String())
	}

	position := &positionDetails{
		TokenID:     tokenID,
		Token0:      results[2].(common.Address),
		Token1:      results[3].(common.Address),
		Fee:         results[4].(uint32),
		TickLower:   results[5].(int32),
		TickUpper:   results[6].(int32),
		Liquidity:   results[7].(*big.Int),
		TokensOwed0: results[10].(*big.Int),
		TokensOwed1: results[11].(*big.Int),
	}
	return position, nil
}

func (a *UniswapV3WalletAdapter) getTokenMetadata(ctx context.Context, token common.Address) (tokenMetadata, error) {
	symbolOutput, err := a.callContract(ctx, token, a.erc20ABI, "symbol")
	if err != nil {
		return tokenMetadata{}, err
	}
	symbolResults, err := a.erc20ABI.Unpack("symbol", symbolOutput)
	if err != nil || len(symbolResults) == 0 {
		return tokenMetadata{}, fmt.Errorf("failed to unpack symbol: %w", err)
	}
	symbol, ok := symbolResults[0].(string)
	if !ok {
		return tokenMetadata{}, fmt.Errorf("unexpected symbol type")
	}

	decimalsOutput, err := a.callContract(ctx, token, a.erc20ABI, "decimals")
	if err != nil {
		return tokenMetadata{}, err
	}
	decimalsResults, err := a.erc20ABI.Unpack("decimals", decimalsOutput)
	if err != nil || len(decimalsResults) == 0 {
		return tokenMetadata{}, fmt.Errorf("failed to unpack decimals: %w", err)
	}

	decimalsValue := 18
	switch typed := decimalsResults[0].(type) {
	case uint8:
		decimalsValue = int(typed)
	case *big.Int:
		decimalsValue = int(typed.Int64())
	}

	return tokenMetadata{
		Symbol:   symbol,
		Decimals: decimalsValue,
	}, nil
}

func (a *UniswapV3WalletAdapter) getPoolAddress(ctx context.Context, token0 common.Address, token1 common.Address, fee uint32) (common.Address, error) {
	output, err := a.callContract(ctx, a.factory, a.factoryABI, "getPool", token0, token1, fee)
	if err != nil {
		return common.Address{}, err
	}
	results, err := a.factoryABI.Unpack("getPool", output)
	if err != nil || len(results) == 0 {
		return common.Address{}, fmt.Errorf("failed to unpack getPool: %w", err)
	}
	poolAddress, ok := results[0].(common.Address)
	if !ok {
		return common.Address{}, fmt.Errorf("unexpected pool address type")
	}
	return poolAddress, nil
}

func (a *UniswapV3WalletAdapter) getPoolSqrtPriceX96(ctx context.Context, pool common.Address) (*big.Int, error) {
	output, err := a.callContract(ctx, pool, a.poolABI, "slot0")
	if err != nil {
		return nil, err
	}
	results, err := a.poolABI.Unpack("slot0", output)
	if err != nil || len(results) == 0 {
		return nil, fmt.Errorf("failed to unpack slot0: %w", err)
	}
	switch typed := results[0].(type) {
	case *big.Int:
		return typed, nil
	default:
		return nil, fmt.Errorf("unexpected sqrtPriceX96 type")
	}
}

func (a *UniswapV3WalletAdapter) callContract(ctx context.Context, target common.Address, parsedABI abi.ABI, method string, args ...interface{}) ([]byte, error) {
	callData, err := parsedABI.Pack(method, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to pack %s: %w", method, err)
	}

	msg := ethereum.CallMsg{
		To:   &target,
		Data: callData,
	}
	return a.client.CallContract(ctx, msg, nil)
}

func estimatePositionAmounts(liquidity *big.Int, tickLower int, tickUpper int, sqrtPriceX96 *big.Int, decimals0 int, decimals1 int) (float64, float64) {
	if liquidity == nil || liquidity.Sign() == 0 || sqrtPriceX96 == nil || sqrtPriceX96.Sign() == 0 {
		return 0, 0
	}

	liq, _ := new(big.Float).SetInt(liquidity).Float64()
	currentSqrt := bigIntToFloat64(sqrtPriceX96) / q96Float
	sqrtLower := tickToSqrtPrice(tickLower)
	sqrtUpper := tickToSqrtPrice(tickUpper)

	if currentSqrt <= 0 || sqrtLower <= 0 || sqrtUpper <= 0 {
		return 0, 0
	}

	var amount0Raw float64
	var amount1Raw float64

	switch {
	case currentSqrt <= sqrtLower:
		amount0Raw = liq * (sqrtUpper - sqrtLower) / (sqrtLower * sqrtUpper)
	case currentSqrt < sqrtUpper:
		amount0Raw = liq * (sqrtUpper - currentSqrt) / (currentSqrt * sqrtUpper)
		amount1Raw = liq * (currentSqrt - sqrtLower)
	default:
		amount1Raw = liq * (sqrtUpper - sqrtLower)
	}

	amount0 := normalizeByDecimals(amount0Raw, decimals0)
	amount1 := normalizeByDecimals(amount1Raw, decimals1)
	return maxZero(amount0), maxZero(amount1)
}

func tickToSqrtPrice(tick int) float64 {
	return math.Pow(1.0001, float64(tick)/2.0)
}

func normalizeByDecimals(value float64, decimals int) float64 {
	if decimals <= 0 {
		return value
	}
	return value / math.Pow10(decimals)
}

func normalizeAssetSymbol(symbol string) string {
	symbol = strings.TrimSpace(strings.ToUpper(symbol))
	if symbol == "WETH" {
		return "ETH"
	}
	return symbol
}

func bigIntToFloat64(value *big.Int) float64 {
	if value == nil {
		return 0
	}
	result, _ := new(big.Float).SetInt(value).Float64()
	return result
}

func maxZero(value float64) float64 {
	if value < 0 {
		return 0
	}
	return value
}
