package pipeline

import (
	"context"
	"sync"

	"market-data-gateway/pkg/types"
)

type Streamer interface {
	StreamUpdates(ctx context.Context, symbol string, out chan<- types.Update) error
}

type StreamConfig struct {
	Streamer Streamer
	Symbol   string
}

func Run(ctx context.Context, streams []StreamConfig, out chan<- types.Update) {
	var wg sync.WaitGroup // wait for all streamers to finish when context is cancelled

	for _, s := range streams {
		wg.Add(1) // increment wait group counter for each streamer
		go func(cfg StreamConfig) {
			defer wg.Done() // decrement counter when streamer exits
			cfg.Streamer.StreamUpdates(ctx, cfg.Symbol, out)
		}(s)
	}

	wg.Wait() // wait for all streamers to finish before exiting
}
