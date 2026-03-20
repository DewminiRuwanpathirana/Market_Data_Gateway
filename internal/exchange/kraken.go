package exchange

import (
	"encoding/json"
	"fmt"
	"net/http"
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

	book := &types.OrderBook{
		Symbol:    symbol,
		Exchange:  "kraken",
		Timestamp: time.Now().UnixMilli(),
		Bids:      make(map[string]string),
		Asks:      make(map[string]string),
	}

	for _, pairData := range data.Result {
		for _, bid := range pairData.Bids {
			price, ok1 := bid[0].(string)
			qty, ok2 := bid[1].(string)
			if !ok1 || !ok2 {
				return nil, fmt.Errorf("kraken snapshot: invalid bid entry")
			}
			book.Bids[price] = qty
		}
		for _, ask := range pairData.Asks {
			price, ok1 := ask[0].(string)
			qty, ok2 := ask[1].(string)
			if !ok1 || !ok2 {
				return nil, fmt.Errorf("kraken snapshot: invalid ask entry")
			}
			book.Asks[price] = qty
		}
	}

	return book, nil
}
