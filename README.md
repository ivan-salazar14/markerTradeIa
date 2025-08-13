# MarkerTradeIa

MarkerTradeIa is a Go-based trading service that processes trading signals, executes trades on Binance, and persists trade executions in PostgreSQL. It uses Kafka for event-driven communication.

## Architecture Overview

- **Kafka Consumer** receives trading signals.
- **Trading Service** validates, executes, and persists trades.
- **Binance Adapter** executes trades on Binance.
- **Postgres Adapter** saves trade executions.

## Flow Diagram

```mermaid
flowchart TD
    A[Kafka Consumer<br>(EventReceiver)] -->|Receives TradingSignal| B[Trading Service]
    B -->|Executes Trade| C[Binance Adapter]
    B -->|Persists Execution| D[Postgres Adapter]
    C -->|TradeExecution| B
    D -->|Save Success/Error| B
```

## Main Components

- [`cmd/main.go`](cmd/main.go): Application entrypoint, wires dependencies.
- [`internal/adapters/kafka/consumer.go`](internal/adapters/kafka/consumer.go): Kafka consumer adapter ([`kafka.ConsumerAdapter`](internal/adapters/kafka/consumer.go)).
- [`internal/service/trading_service.go`](internal/service/trading_service.go): Trading service ([`service.TradingService`](internal/service/trading_service.go)).
- [`internal/adapters/trading/binance/trader.go`](internal/adapters/trading/binance/trader.go): Binance trading adapter ([`binance.BinanceTrader`](internal/adapters/trading/binance/trader.go)).
- [`internal/adapters/repository/postgres/trade_repository.go`](internal/adapters/repository/postgres/trade_repository.go): Postgres repository adapter ([`postgres.TradeRepository`](internal/adapters/repository/postgres/trade_repository.go)).
- [`internal/core/domain/trading_signal.go`](internal/core/domain/trading_signal.go): Trading signal domain ([`domain.TradingSignal`](internal/core/domain/trading_signal.go)).
- [`internal/core/domain/trade_execution.go`](internal/core/domain/trade_execution.go): Trade execution domain ([`domain.TradeExecution`](internal/core/domain/trade_execution.go)).

## Logic Details

1. **Kafka Consumer** listens for trading signals and passes them to the Trading Service.
2. **Trading Service** receives the signal, validates it, and calls the Binance Adapter to execute the trade.
3. After execution, the Trading Service calls the Postgres Adapter to persist the trade execution result.
4. All components communicate via well-defined ports/interfaces for loose coupling.

## Example Signal

A trading signal is represented by [`domain.TradingSignal`](internal/core/domain/trading_signal.go):

```go
type TradingSignal struct {
    ID        string
    Symbol    string
    Price     float64
    Timestamp time.Time
    Type      SignalType
}
```

## Running

Build and run the project:

```sh
go build -o markerTradeIa ./cmd
./markerTradeIa
```

## Configuration

See [`config/config.go`](config/config.go) for configuration options.