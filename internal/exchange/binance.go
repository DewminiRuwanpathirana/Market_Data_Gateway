package exchange

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"market-data-gateway/pkg/types"

	"github.com/gorilla/websocket"
)

type binanceSnapshotResponse struct {
	LastUpdateID int64      `json:"lastUpdateId"`
	Bids         [][]string `json:"bids"`
	Asks         [][]string `json:"asks"`
}

type binanceDepthUpdate struct {
	EventType     string     `json:"e"`
	EventTime     int64      `json:"E"`
	Symbol        string     `json:"s"`
	FirstUpdateID int64      `json:"U"`
	LastUpdateID  int64      `json:"u"`
	Bids          [][]string `json:"b"`
	Asks          [][]string `json:"a"`
}

type BinanceClient struct {
	baseURL string
	wsURL   string
}

func NewBinanceClient() *BinanceClient {
	return &BinanceClient{
		baseURL: "https://api.binance.com/api/v3/depth?symbol=%s",
		wsURL:   "wss://stream.binance.com:9443/ws/%s@depth",
	}
}

// Gets the current order book from Binance
func (b *BinanceClient) FetchSnapshot(symbol string) (*types.OrderBook, error) {
	url := fmt.Sprintf(b.baseURL, strings.ToUpper(symbol))

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("binance snapshot fetch %s: %w", symbol, err)
	}
	defer resp.Body.Close()

	var data binanceSnapshotResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("binance snapshot parse %s: %w", symbol, err)
	}

	// Convert to OrderBook type
	book := &types.OrderBook{
		Symbol:       symbol,
		Exchange:     "binance",
		Timestamp:    time.Now().UnixMilli(),
		LastUpdateID: data.LastUpdateID,
		Bids:         []types.PriceLevel{},
		Asks:         []types.PriceLevel{},
	}

	// Convert bids (strings to floats)
	for _, bid := range data.Bids {
		pl, err := parsePriceLevel(bid)
		if err != nil {
			return nil, fmt.Errorf("binance snapshot bid: %w", err)
		}
		book.Bids = append(book.Bids, pl)
	}

	// Convert asks (strings to floats)
	for _, ask := range data.Asks {
		pl, err := parsePriceLevel(ask)
		if err != nil {
			return nil, fmt.Errorf("binance snapshot ask: %w", err)
		}
		book.Asks = append(book.Asks, pl)
	}

	return book, nil
}

type binanceSeq struct {
	lastID      int64
	prevU       int64
	initialized bool
}

func (s *binanceSeq) accept(raw binanceDepthUpdate) (valid, gap bool) {
	if raw.LastUpdateID < s.lastID {
		return false, false // old update
	}

	if !s.initialized {
		valid = raw.FirstUpdateID <= s.lastID && raw.LastUpdateID >= s.lastID
		if valid {
			s.initialized = true
			s.prevU = raw.LastUpdateID
		}
		return valid, false
	}

	if raw.FirstUpdateID != s.prevU+1 {
		return false, true // gap
	}

	s.prevU = raw.LastUpdateID
	return true, false
}

func (b *BinanceClient) StreamUpdates(ctx context.Context, symbol string, out chan<- types.Update) error {
	wsURL := fmt.Sprintf(b.wsURL, strings.ToLower(symbol))

	conn, _, err := websocket.DefaultDialer.DialContext(ctx, wsURL, nil)
	if err != nil {
		return fmt.Errorf("binance ws %s: %w", symbol, err)
	}
	defer conn.Close()

	// close connection when context is cancelled
	go func() {
		<-ctx.Done()
		conn.Close()
	}()

	rawCh := b.readDepthStream(conn)

sync: // sync label used to restart the loop when detect a gap
	for {
		snapshot, err := b.FetchSnapshot(symbol)
		if err != nil {
			return fmt.Errorf("binance snapshot %s: %w", symbol, err)
		}

		seq := binanceSeq{lastID: snapshot.LastUpdateID}

		for {
			select {
			case <-ctx.Done():
				return nil
			case raw, ok := <-rawCh:
				if !ok {
					return nil
				}

				valid, gap := seq.accept(raw)
				if gap {
					continue sync
				}
				if !valid {
					continue
				}

				update, err := toUpdate(symbol, raw)
				if err != nil {
					return err
				}

				out <- update
			}
		}
	}
}

func (b *BinanceClient) readDepthStream(conn *websocket.Conn) <-chan binanceDepthUpdate {
	rawCh := make(chan binanceDepthUpdate, 100)

	go func() {
		defer close(rawCh)
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}
			var raw binanceDepthUpdate
			if err := json.Unmarshal(msg, &raw); err != nil {
				continue
			}

			rawCh <- raw
		}
	}()
	return rawCh
}

func toUpdate(symbol string, raw binanceDepthUpdate) (types.Update, error) {
	update := types.Update{
		Symbol:       symbol,
		Exchange:     "binance",
		Timestamp:    raw.EventTime,
		LastUpdateID: raw.LastUpdateID,
	}

	for _, b := range raw.Bids {
		pl, err := parsePriceLevel(b)
		if err != nil {
			return types.Update{}, fmt.Errorf("binance bid: %w", err)
		}
		update.Bids = append(update.Bids, pl)
	}

	for _, a := range raw.Asks {
		pl, err := parsePriceLevel(a)
		if err != nil {
			return types.Update{}, fmt.Errorf("binance ask: %w", err)
		}
		update.Asks = append(update.Asks, pl)
	}

	return update, nil
}

func parsePriceLevel(entry []string) (types.PriceLevel, error) {
	price, err := strconv.ParseFloat(entry[0], 64)
	if err != nil {
		return types.PriceLevel{}, fmt.Errorf("invalid price '%s': %w", entry[0], err)
	}
	qty, err := strconv.ParseFloat(entry[1], 64)
	if err != nil {
		return types.PriceLevel{}, fmt.Errorf("invalid quantity '%s': %w", entry[1], err)
	}
	return types.PriceLevel{Price: price, Quantity: qty}, nil
}
