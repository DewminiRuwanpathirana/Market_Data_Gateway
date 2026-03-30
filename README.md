# Market Data Gateway

## Run

```
go run ./cmd/gateway
```

## Configuration

Edit 'config.yaml' to add or remove exchange symbols.


## Client Connection

Connect to 'ws://localhost:8080/ws' and send a subscription message:

```
{"exchange": "binance", "symbol": "BTCUSDT"}
```

The server responds with a full snapshot with incremental updates.

## Shutdown

Press `Ctrl+C` to shut down gracefully.
