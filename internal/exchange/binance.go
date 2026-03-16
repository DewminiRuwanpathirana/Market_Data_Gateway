package exchange

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"market-data-gateway/pkg/types"
)

type BinanceResponse struct {
	LastUpdateID int64      `json:"lastUpdateId"`
	Bids         [][]string `json:"bids"`
	Asks         [][]string `json:"asks"`
}

type BinanceClient struct {
	baseURL string
}

func NewBinanceClient() *BinanceClient {
	return &BinanceClient{
		baseURL: "https://api.binance.com/api/v3/depth?symbol=%s&limit=5",
	}
}

// Gets the current order book from Binance
func (b *BinanceClient) FetchSnapshot(symbol string) (*types.OrderBook, error) {
	url := fmt.Sprintf(b.baseURL, strings.ToUpper(symbol))

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch: %w", err)
	}
	defer resp.Body.Close()

	var data BinanceResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to parse: %w", err)
	}

	// Convert to OrderBook type
	book := &types.OrderBook{
		Symbol:    symbol,
		Exchange:  "binance",
		Timestamp: time.Now().UnixMilli(),
		Bids:      []types.PriceLevel{},
		Asks:      []types.PriceLevel{},
	}

	// Convert bids (strings to floats)
	for _, bid := range data.Bids {
		price, err := strconv.ParseFloat(bid[0], 64)
		if err != nil {
			return nil, fmt.Errorf("invalid bid price '%s': %w", bid[0], err)
		}

		qty, err := strconv.ParseFloat(bid[1], 64)
		if err != nil {
			return nil, fmt.Errorf("invalid bid quantity '%s': %w", bid[1], err)
		}

		book.Bids = append(book.Bids, types.PriceLevel{
			Price:    price,
			Quantity: qty,
		})
	}

	// Convert asks (strings to floats)
	for _, ask := range data.Asks {
		price, err := strconv.ParseFloat(ask[0], 64)
		if err != nil {
			return nil, fmt.Errorf("invalid ask price '%s': %w", ask[0], err)
		}

		qty, err := strconv.ParseFloat(ask[1], 64)
		if err != nil {
			return nil, fmt.Errorf("invalid ask quantity '%s': %w", ask[1], err)
		}

		book.Asks = append(book.Asks, types.PriceLevel{
			Price:    price,
			Quantity: qty,
		})
	}

	return book, nil
}
