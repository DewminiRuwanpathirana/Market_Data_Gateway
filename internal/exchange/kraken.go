package exchange

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"market-data-gateway/pkg/types"

	"github.com/gorilla/websocket"
)

type krakenSubscribeMsg struct {
	Method string                `json:"method"`
	Params krakenSubscribeParams `json:"params"`
}

type krakenSubscribeParams struct {
	Channel string   `json:"channel"`
	Symbol  []string `json:"symbol"`
}

type krakenBookMsg struct {
	Channel string           `json:"channel"`
	Type    string           `json:"type"`
	Data    []krakenBookData `json:"data"`
}

type krakenBookData struct {
	Symbol    string        `json:"symbol"`
	Bids      []krakenLevel `json:"bids"`
	Asks      []krakenLevel `json:"asks"`
	Timestamp string        `json:"timestamp"`
}

type krakenLevel struct {
	Price float64 `json:"price"`
	Qty   float64 `json:"qty"`
}

type KrakenConnection struct {
	wsURL string
}

func NewKrakenConnection(wsURL string) *KrakenConnection {
	return &KrakenConnection{
		wsURL: wsURL,
	}
}

func (k *KrakenConnection) StreamUpdates(ctx context.Context, symbol string, out chan<- types.Update) error {
	conn, _, err := websocket.DefaultDialer.DialContext(ctx, k.wsURL, nil) // opens a WebSocket connection
	if err != nil {
		return fmt.Errorf("kraken ws %s: %w", symbol, err)
	}
	defer conn.Close()

	// close connection when context is cancelled
	go func() {
		<-ctx.Done()
		conn.Close()
	}()

	// send subscribe message to start receiving book updates
	sub := krakenSubscribeMsg{
		Method: "subscribe",
		Params: krakenSubscribeParams{
			Channel: "book",
			Symbol:  []string{symbol},
		},
	}
	if err := conn.WriteJSON(sub); err != nil {
		return fmt.Errorf("kraken ws subscribe %s: %w", symbol, err)
	}

	rawCh := k.readBookStream(conn) // starts buffering incoming WebSocket messages into a channel

	for {
		select {
		case <-ctx.Done():
			return nil
		case raw, ok := <-rawCh: // read from WebSocket stream
			if !ok {
				return nil
			}

			for _, d := range raw.Data {
				out <- toKrakenUpdate(d, raw.Type)
			}
		}
	}
}

func (k *KrakenConnection) readBookStream(conn *websocket.Conn) <-chan krakenBookMsg {
	rawCh := make(chan krakenBookMsg, 100)

	go func() {
		defer close(rawCh)
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}
			var raw krakenBookMsg
			if err := json.Unmarshal(msg, &raw); err != nil {
				continue
			}

			if raw.Channel != "book" || (raw.Type != "snapshot" && raw.Type != "update") {
				continue
			}

			rawCh <- raw
		}
	}()
	return rawCh
}

func toKrakenUpdate(d krakenBookData, msgType string) types.Update {
	update := types.Update{
		IsSnapshot: msgType == "snapshot",
		Symbol:     d.Symbol,
		Exchange:   "kraken",
		Timestamp:  time.Now().UnixMilli(),
		Bids:       make(map[string]string),
		Asks:       make(map[string]string),
	}
	for _, b := range d.Bids {
		update.Bids[strconv.FormatFloat(b.Price, 'f', -1, 64)] = strconv.FormatFloat(b.Qty, 'f', -1, 64)
	}
	for _, a := range d.Asks {
		update.Asks[strconv.FormatFloat(a.Price, 'f', -1, 64)] = strconv.FormatFloat(a.Qty, 'f', -1, 64)
	}
	return update
}
