package server

import (
	"net/http"
	"strconv"
	"sync"

	"market-data-gateway/pkg/types"

	"github.com/gorilla/websocket"
)

type BookStore interface {
	GetBook(exchange, symbol string) *types.OrderBook
	ApplyUpdate(u types.Update)
}

// converts http requests to WebSocket connections
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Message struct {
	Type     string            `json:"type"`
	Exchange string            `json:"exchange,omitempty"`
	Symbol   string            `json:"symbol,omitempty"`
	Bids     map[string]string `json:"bids,omitempty"`
	Asks     map[string]string `json:"asks,omitempty"`
	Error    string            `json:"error,omitempty"`
}

type subscribeMsg struct {
	Exchange string `json:"exchange"`
	Symbol   string `json:"symbol"`
}

// each websocket client has a send channel to receive messages from the server
type client struct {
	send      chan Message
	closeOnce sync.Once
	exchange  string
	symbol    string
}

type Server struct {
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
			Type:     "update",
			Exchange: update.Exchange,
			Symbol:   update.Symbol,
			Bids:     filterZeros(update.Bids),
			Asks:     filterZeros(update.Asks),
		})
	}
}

func (s *Server) broadcast(msg Message) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for c := range s.clients {
		if c.exchange != msg.Exchange || c.symbol != msg.Symbol {
			continue
		}
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

	// read and validate subscription from client before sending data
	var sub subscribeMsg
	for {
		if err := conn.ReadJSON(&sub); err != nil {
			conn.WriteJSON(Message{Type: "error", Error: "Invalid subscription message"})
			continue
		}
		break
	}

	c := &client{
		send:     make(chan Message, 256),
		exchange: sub.Exchange,
		symbol:   sub.Symbol,
	}

	// send snapshot for the requested symbol
	if book := s.manager.GetBook(sub.Exchange, sub.Symbol); book != nil {
		c.send <- Message{
			Type:     "snapshot",
			Exchange: sub.Exchange,
			Symbol:   book.Symbol,
			Bids:     book.Bids,
			Asks:     book.Asks,
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
