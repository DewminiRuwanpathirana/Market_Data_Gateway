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
	var wg sync.WaitGroup

	for _, s := range streams {
		wg.Add(1)
		go func(cfg StreamConfig) {
			defer wg.Done()
			cfg.Streamer.StreamUpdates(ctx, cfg.Symbol, out)
		}(s)
	}

	wg.Wait()
}
