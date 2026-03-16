package main

import (
	"fmt"
	"log"
	"market-data-gateway/internal/exchange"
)

func main() {

	binance := exchange.NewBinanceClient()

	book, err := binance.FetchSnapshot("BTCUSDT")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Exchange: %s\n", book.Exchange)
	fmt.Printf("Symbol: %s\n", book.Symbol)

	fmt.Println("\nBids:")
	for _, bid := range book.Bids {
		fmt.Printf(" Price: $%.2f - Quantity: %.4f \n", bid.Price, bid.Quantity)
	}

	fmt.Println("\nAsks:")
	for _, ask := range book.Asks {
		fmt.Printf(" Price: $%.2f - Quantity: %.4f \n", ask.Price, ask.Quantity)
	}
}
