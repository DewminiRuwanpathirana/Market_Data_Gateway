package exchange

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"market-data-gateway/pkg/types"
)

type KrakenResponse struct {
	Error  []string                       `json:"error"`
	Result map[string]KrakenOrderBookData `json:"result"`
}

type KrakenOrderBookData struct {
	Bids [][]any `json:"bids"`
	Asks [][]any `json:"asks"`
}

type KrakenClient struct {
	baseURL string
}

func NewKrakenClient() *KrakenClient {
	return &KrakenClient{
		baseURL: "https://api.kraken.com/0/public/Depth?pair=%s&count=5",
	}
}

// Gets the current order book from Kraken
func (k *KrakenClient) FetchSnapshot(symbol string) (*types.OrderBook, error) {
	url := fmt.Sprintf(k.baseURL, symbol)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch: %w", err)
	}
	defer resp.Body.Close()

	var data KrakenResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to parse: %w", err)
	}

	if len(data.Error) > 0 {
		return nil, fmt.Errorf("kraken API error: %v", data.Error)
	}

	// Convert to OrderBook type
	book := &types.OrderBook{
		Symbol:    symbol,
		Exchange:  "kraken",
		Timestamp: time.Now().UnixMilli(),
		Bids:      []types.PriceLevel{},
		Asks:      []types.PriceLevel{},
	}

	for _, pairData := range data.Result {
		// Convert bids (any to floats)
		for _, bid := range pairData.Bids {
			priceStr, ok := bid[0].(string) // check if the price is a string
			if !ok {
				return nil, fmt.Errorf("invalid bid price type")
			}

			qtyStr, ok := bid[1].(string)
			if !ok {
				return nil, fmt.Errorf("invalid bid quantity type")
			}

			price, err := strconv.ParseFloat(priceStr, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid bid price '%s': %w", priceStr, err)
			}

			qty, err := strconv.ParseFloat(qtyStr, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid bid quantity '%s': %w", qtyStr, err)
			}

			book.Bids = append(book.Bids, types.PriceLevel{
				Price:    price,
				Quantity: qty,
			})
		}

		// Convert asks (any to floats)
		for _, ask := range pairData.Asks {
			priceStr, ok := ask[0].(string)
			if !ok {
				return nil, fmt.Errorf("invalid ask price type")
			}

			qtyStr, ok := ask[1].(string)
			if !ok {
				return nil, fmt.Errorf("invalid ask quantity type")
			}

			price, err := strconv.ParseFloat(priceStr, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid ask price '%s': %w", priceStr, err)
			}

			qty, err := strconv.ParseFloat(qtyStr, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid ask quantity '%s': %w", qtyStr, err)
			}

			book.Asks = append(book.Asks, types.PriceLevel{
				Price:    price,
				Quantity: qty,
			})
		}
	}

	return book, nil
}
