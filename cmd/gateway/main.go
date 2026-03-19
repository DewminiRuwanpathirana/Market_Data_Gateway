package main

import (
	"context"
	"fmt"

	"market-data-gateway/internal/exchange"
	"market-data-gateway/internal/pipeline"
	"market-data-gateway/pkg/types"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	updates := make(chan types.Update, 100)

	streams := []pipeline.StreamConfig{
		{
			Streamer: exchange.NewBinanceClient(),
			Symbol:   "BTCUSDT",
		},
	}

	go pipeline.Run(ctx, streams, updates)

	for {
		select {
		case update := <-updates:
			fmt.Printf("Exchange: %s Symbol: %s LastUpdateID: %d Bids: %v Asks: %v\n",
				update.Exchange, update.Symbol, update.LastUpdateID, update.Bids, update.Asks)
		case <-ctx.Done():
			return
		}
	}
}
