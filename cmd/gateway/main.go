package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"market-data-gateway/internal/config"
	"market-data-gateway/internal/exchange"
	"market-data-gateway/internal/orderbook"
	"market-data-gateway/internal/pipeline"
	"market-data-gateway/internal/server"
	"market-data-gateway/pkg/types"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// cancel context on SIGINT/SIGTERM for shutdown
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
		<-ch
		cancel()
	}()

	exchanges := map[string]pipeline.Streamer{
		"binance": exchange.NewBinanceConnection(),
		"kraken":  exchange.NewKrakenConnection(),
	}

	manager := orderbook.NewManager()

	var streams []pipeline.StreamConfig
	for _, s := range config.Symbols {
		manager.InitSymbol(s.Symbol, s.Exchange) // initialize empty order book for each symbol
		streams = append(streams, pipeline.StreamConfig{
			Streamer: exchanges[s.Exchange], // get the streamer object to call StreamUpdates
			Symbol:   s.Symbol,
		})
	}

	updates := make(chan types.Update, 100)

	go func() {
		pipeline.Run(ctx, streams, updates)
		close(updates)
	}()

	srv := server.NewServer(manager)

	go srv.Run(updates)

	http.Handle("/ws", srv) // route incoming WebSocket connections(clients) to the server
	fmt.Println("listening on :8080")

	httpSrv := &http.Server{Addr: ":8080"}
	go func() {
		<-ctx.Done()
		fmt.Println("shutting down...")
		httpSrv.Shutdown(context.Background()) // shutdown server when context is cancelled
	}()

	if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		fmt.Println("server error:", err)
	}
}
