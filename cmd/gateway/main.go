package main

import (
	"context"
	"fmt"
	"net/http"

	"market-data-gateway/internal/exchange"
	"market-data-gateway/internal/orderbook"
	"market-data-gateway/internal/pipeline"
	"market-data-gateway/internal/server"
	"market-data-gateway/pkg/types"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	binance := exchange.NewBinanceClient()

	snapshot, err := binance.FetchSnapshot("BTCUSDT")
	if err != nil {
		fmt.Println("snapshot error:", err)
		return
	}

	manager := orderbook.NewManager()
	manager.SetSnapshot(snapshot)

	updates := make(chan types.Update, 100)

	streams := []pipeline.StreamConfig{
		{
			Streamer: binance,
			Symbol:   "BTCUSDT",
		},
	}

	go pipeline.Run(ctx, streams, updates)

	srv := server.NewServer(manager)
	go srv.Run(updates)

	http.Handle("/ws", srv)
	fmt.Println("listening on :8080")
	http.ListenAndServe(":8080", nil)
}
