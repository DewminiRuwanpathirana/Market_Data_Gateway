package server

import (
	"net/http"
	"strconv"
	"sync"

	// "market-data-gateway/internal/orderbook"
	"market-data-gateway/pkg/types"

	"github.com/gorilla/websocket"
)

type BookStore interface {
	GetAll() []*types.OrderBook
	ApplyUpdate(u types.Update)
}

// converts http requests to WebSocket connections
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Message struct {
	Type   string            `json:"type"`
	Symbol string            `json:"symbol"`
	Bids   map[string]string `json:"bids"`
	Asks   map[string]string `json:"asks"`
}

// each websocket client has a send channel to receive messages from the server
type client struct {
	send      chan Message
	closeOnce sync.Once
}

type Server struct {
	// manager *orderbook.Manager
	manager BookStore
	clients map[*client]struct{}
	mu      sync.Mutex
}

func NewServer(m BookStore) *Server {
	return &Server{
		manager: m,
		clients: make(map[*client]struct{}),
	}
}
func filterZeros(levels map[string]string) map[string]string {
	out := make(map[string]string, len(levels))
	for price, qty := range levels {
		q, _ := strconv.ParseFloat(qty, 64)
		if q != 0 {
			out[price] = qty
		}
	}
	return out
}

// reads from the updates channel, applies to manager and broadcasts to all connected clients
func (s *Server) Run(updates <-chan types.Update) {
	for update := range updates {
		s.manager.ApplyUpdate(update)
		s.broadcast(Message{
			Type:   "update",
			Symbol: update.Symbol,
			Bids:   filterZeros(update.Bids),
			Asks:   filterZeros(update.Asks),
		})
	}
}

func (s *Server) broadcast(msg Message) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for c := range s.clients {
		select {
		case c.send <- msg:
		default:
			delete(s.clients, c)
			c.closeOnce.Do(func() {
				close(c.send)
			})
		}
	}
}

// handles a new downstream client WebSocket connection
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	c := &client{send: make(chan Message, 256)}

	// send current snapshot for every symbol
	for _, book := range s.manager.GetAll() {
		c.send <- Message{
			Type:   "snapshot",
			Symbol: book.Symbol,
			Bids:   book.Bids,
			Asks:   book.Asks,
		}
	}

	s.mu.Lock()
	s.clients[c] = struct{}{}
	s.mu.Unlock()

	// writes to the WebSocket connection
	go func() {
		defer conn.Close()
		for msg := range c.send {
			if err := conn.WriteJSON(msg); err != nil {
				return
			}
		}
	}()

	// detect client disconnection by reading from the WebSocket
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			s.mu.Lock()
			delete(s.clients, c)
			s.mu.Unlock()
			c.closeOnce.Do(func() {
				close(c.send)
			})
			return
		}
	}
}
