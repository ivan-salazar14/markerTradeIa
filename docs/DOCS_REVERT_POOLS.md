# Revert Finance Integration - Pool Monitoring API

This document describes how the backend monitors Uniswap V3 pools using Revert Finance (Subgraph and API) and the data schema returned by the service.

## Data Sources

1. **Uniswap V3 Subgraphs (via Revert):** Used to fetch the list of top liquidity pools based on TVL and Volume.
2. **Revert Positions API:** Used to fetch detailed statistics for specific LP positions (identified by Token ID).

## Data Models

### Liquidity Pool (`domain.LiquidityPool`)

Represents a Uniswap V3 pool.

| Field | Type | Description |
| :--- | :--- | :--- |
| `id` | string | The smart contract address of the pool. |
| `network` | string | The blockchain network (e.g., "ethereum", "polygon", "arbitrum"). |
| `protocol` | string | Always "uniswap_v3" for this adapter. |
| `symbol0` | string | Symbol of the first token in the pair. |
| `symbol1` | string | Symbol of the second token in the pair. |
| `fee_tier` | int | The fee tier of the pool (e.g., 500, 3000, 10000). |
| `tvl_usd` | float64 | Total Value Locked in USD. |
| `volume_usd` | float64 | 24h trading volume in USD. |
| `updated_at` | datetime | Timestamp of the last update. |

### Position Stats (`domain.PositionStats`)

Represents performance metrics for a specific liquidity position.

| Field | Type | Description |
| :--- | :--- | :--- |
| `id` | string | The Token ID of the NFT position. |
| `network` | string | The blockchain network. |
| `uncollected_fee` | float64 | Fees earned but not yet claimed (in USD). |
| `apr` | float64 | Annual Percentage Rate (estimated). |
| `roi` | float64 | Return on Investment (percentage). |
| `updated_at` | datetime | Timestamp of the last update. |

## Implementation Details

The adapter is implemented in `internal/infrastructure/adapters/monitoring/revert/revert_adapter.go`.

### Example Mapping

When fetching pools, the adapter maps Subgraph fields as follows:

- `totalValueLockedUSD` (string) -> `TVLUSD` (float64)
- `volumeUSD` (string) -> `VolumeUSD` (float64)
- `feeTier` (string) -> `FeeTier` (int)

## Usage for Frontend

The frontend can consume these models through the monitoring service endpoints. Use the `id` of the pool to query specific position details if available.

### Sample JSON Response (Pools)

```json
[
  {
    "id": "0x88e6a0c2ddd26feeb64f039a2c41296fcb3f5640",
    "network": "ethereum",
    "protocol": "uniswap_v3",
    "symbol0": "WETH",
    "symbol1": "USDC",
    "fee_tier": 500,
    "tvl_usd": 1000000.50,
    "volume_usd": 500000.25,
    "updated_at": "2026-03-10T17:50:00Z"
  }
]
```
