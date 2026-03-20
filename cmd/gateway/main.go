package main

import (
	"context"
	"fmt"

	"market-data-gateway/internal/exchange"
	"market-data-gateway/internal/orderbook"
	"market-data-gateway/internal/pipeline"
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

	for update := range updates {
		manager.ApplyUpdate(update)
		book := manager.GetBook("BTCUSDT")
		fmt.Printf("bids=%d asks=%d lastUpdateID=%d\n", len(book.Bids), len(book.Asks), book.LastUpdateID)
	}
}
