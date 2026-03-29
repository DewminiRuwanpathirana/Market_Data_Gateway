package exchange

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
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

type BinanceConnection struct {
	baseURL string
	wsURL   string
}

func NewBinanceConnection() *BinanceConnection {
	return &BinanceConnection{
		baseURL: "https://api.binance.com/api/v3/depth?symbol=%s",
		wsURL:   "wss://stream.binance.com:9443/ws/%s@depth",
	}
}

// Gets the current order book from Binance
func (b *BinanceConnection) FetchSnapshot(symbol string) (*types.OrderBook, error) {
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

	book := &types.OrderBook{
		Symbol:       symbol,
		Exchange:     "binance",
		Timestamp:    time.Now().UnixMilli(),
		LastUpdateID: data.LastUpdateID,
		Bids:         make(map[string]string),
		Asks:         make(map[string]string),
	}

	for _, bid := range data.Bids {
		if len(bid) < 2 {
			continue
		}
		book.Bids[bid[0]] = bid[1]
	}
	for _, ask := range data.Asks {
		if len(ask) < 2 {
			continue
		}
		book.Asks[ask[0]] = ask[1]
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
		valid = raw.FirstUpdateID <= s.lastID && raw.LastUpdateID >= s.lastID // check if snapshot is within the update range
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

func (b *BinanceConnection) StreamUpdates(ctx context.Context, symbol string, out chan<- types.Update) error {
	wsURL := fmt.Sprintf(b.wsURL, strings.ToLower(symbol))

	conn, _, err := websocket.DefaultDialer.DialContext(ctx, wsURL, nil) // opens a WebSocket connection
	if err != nil {
		return fmt.Errorf("binance ws %s: %w", symbol, err)
	}
	defer conn.Close()

	// close connection when context is cancelled
	go func() {
		<-ctx.Done()
		conn.Close()
	}()

	rawCh := b.readDepthStream(conn) // starts buffering incoming WebSocket messages into a channel

sync: // sync label used to restart the loop when detect a gap
	for {
		snapshot, err := b.FetchSnapshot(symbol)
		if err != nil {
			return fmt.Errorf("binance snapshot %s: %w", symbol, err)
		}

		// send snapshot as initial state through the pipeline
		out <- types.Update{
			Symbol:       snapshot.Symbol,
			Exchange:     snapshot.Exchange,
			Bids:         snapshot.Bids,
			Asks:         snapshot.Asks,
			Timestamp:    snapshot.Timestamp,
			LastUpdateID: snapshot.LastUpdateID,
		}

		seq := binanceSeq{lastID: snapshot.LastUpdateID}

		for {
			select {
			case <-ctx.Done():
				return nil
			case raw, ok := <-rawCh: // read from WebSocket stream
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

				out <- toUpdate(symbol, raw)
			}
		}
	}
}

func (b *BinanceConnection) readDepthStream(conn *websocket.Conn) <-chan binanceDepthUpdate {
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

func toUpdate(symbol string, raw binanceDepthUpdate) types.Update {
	update := types.Update{
		Symbol:       symbol,
		Exchange:     "binance",
		Timestamp:    raw.EventTime,
		LastUpdateID: raw.LastUpdateID,
		Bids:         make(map[string]string),
		Asks:         make(map[string]string),
	}

	for _, b := range raw.Bids {
		if len(b) < 2 {
			continue
		}
		update.Bids[b[0]] = b[1]
	}
	for _, a := range raw.Asks {
		if len(a) < 2 {
			continue
		}
		update.Asks[a[0]] = a[1]
	}

	return update
}
