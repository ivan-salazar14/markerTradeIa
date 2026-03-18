# MarkerTradeIa

MarkerTradeIa is a Go backend for monitoring and synchronizing a delta-neutral crypto strategy. The current MVP focuses on measuring LP exposure from an EVM wallet, comparing it against a hedge on Hyperliquid, and exposing that state through an authenticated HTTP API.

## Current MVP Scope

- Connect and inspect an LP wallet.
- Read active pool exposure from the wallet adapter.
- Read the current short exposure from the Hyperliquid adapter.
- Evaluate hedge drift through a strategy use case.
- Persist hedge states, actions, sync events, and wallet connections in PostgreSQL.
- Expose hedge and monitoring endpoints through HTTP.
- Run background monitoring for liquidity pools and hedge reconciliation.

## Architecture

- `cmd/main.go`: dependency wiring and application startup.
- `config/config.go`: environment-based configuration.
- `internal/application/usecases/hedge/`: hedge synchronization use case and strategy.
- `internal/application/services/monitoring/`: background monitoring and cached pool snapshots.
- `internal/infrastructure/adapters/dex/uniswap/`: EVM wallet adapter for LP exposure.
- `internal/infrastructure/adapters/perps/`: Hyperliquid adapter.
- `internal/infrastructure/adapters/repository/hedgeAdapter/`: persistence for hedge state.
- `internal/infrastructure/adapters/api/`: HTTP server, controllers, and middleware.

## Required Environment Variables

Use [.env.example](C:\Users\Asus\Documents\GitHub\makerTradeia\.env.example) as the base:

- `PORT`
- `DATABASE_URL`
- `JWT_SECRET`
- `MONITORING_SERVICE_API_KEY`
- `EVM_RPC_URL`
- `UNISWAP_POSITION_MANAGER`
- `DEFAULT_HEDGE_ASSET`
- `DEFAULT_LP_WALLET_ADDRESS`
- `HYPERLIQUID_PRIVATE_KEY`
- `HYPERLIQUID_ADDRESS`

Useful operational flags:

- `SAFE_MODE=true`
- `DRY_RUN=true`

## API Endpoints

Public:

- `POST /auth/login`
- `POST /auth/refresh`

Protected:

- `GET /api/v1/pools`
- `GET /api/hedge/strategy`
- `GET /api/hedge/stats`
- `GET /api/hedge/wallets`
- `POST /api/hedge/wallets/connect`
- `POST /api/hedge/wallets/disconnect`
- `POST /api/hedge/sync`
- `GET /api/hedge/delta`
- `GET /api/hedge/permissions`
- `GET /api/hedge/safe-mode`
- `GET /api/hedge/sync-flow`

Authentication works with either:

- `Authorization: Bearer <jwt>`
- `X-API-Key: <service-api-key>`

## Running

```powershell
go build -o markerTradeIa.exe ./cmd
.\markerTradeIa.exe
```

## Status

The hedge orchestration, persistence, API surface, and monitoring cache are in place for the MVP. Some exchange-side integrations still use partial simulation and should be hardened before handling real funds.
