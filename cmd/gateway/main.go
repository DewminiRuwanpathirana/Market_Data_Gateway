package main

import (
	"fmt"
	"log"
	"market-data-gateway/internal/exchange"
	"market-data-gateway/internal/orderbook"
	"market-data-gateway/pkg/types"
)

func main() {

	manager := orderbook.NewManager()

	// Test Binance
	fmt.Println("BINANCE")
	binance := exchange.NewBinanceClient()
	btcBinance, err := binance.FetchSnapshot("BTCUSDT")
	if err != nil {
		log.Printf("Binance error: %v\n", err)
	} else {
		manager.SetSnapshot(btcBinance)
		printOrderBook(btcBinance)
	}

	// Test Kraken
	fmt.Println("KRAKEN")
	kraken := exchange.NewKrakenClient()
	btcKraken, err := kraken.FetchSnapshot("XBTUSD")
	if err != nil {
		log.Printf("Kraken error: %v\n", err)
	} else {
		manager.SetSnapshot(btcKraken)
		printOrderBook(btcKraken)
	}
}

func printOrderBook(book *types.OrderBook) {
	fmt.Printf("Exchange: %s\n", book.Exchange)
	fmt.Printf("Symbol: %s\n", book.Symbol)
	fmt.Printf("Timestamp: %d\n", book.Timestamp)

	fmt.Println("\nBids:")
	for _, bid := range book.Bids {
		fmt.Printf(" Price: %.2f - Quantity: %.4f\n", bid.Price, bid.Quantity)
	}

	fmt.Println("\nAsks:")
	for _, ask := range book.Asks {
		fmt.Printf(" Price: %.2f - Quantity: %.4f\n", ask.Price, ask.Quantity)
	}
}
