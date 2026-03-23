package server

import (
	"net/http"
	"sync"

	"market-data-gateway/internal/orderbook"
	"market-data-gateway/pkg/types"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Message struct {
	Type     string            `json:"type"`
	Symbol   string            `json:"symbol"`
	Exchange string            `json:"exchange"`
	Bids     map[string]string `json:"bids"`
	Asks     map[string]string `json:"asks"`
}

type client struct {
	send chan Message
}

type Server struct {
	manager *orderbook.Manager
	clients map[*client]struct{}
	mu      sync.Mutex
}

func NewServer(m *orderbook.Manager) *Server {
	return &Server{
		manager: m,
		clients: make(map[*client]struct{}),
	}
}

// reads from the updates channel, applies to manager and broadcasts to all connected clients
func (s *Server) Run(updates <-chan types.Update) {
	for update := range updates {
		s.manager.ApplyUpdate(update)
		s.broadcast(Message{
			Type:     "update",
			Symbol:   update.Symbol,
			Exchange: update.Exchange,
			Bids:     update.Bids,
			Asks:     update.Asks,
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
			// need to implement proper client handling and disconnection logic here
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
			Type:     "snapshot",
			Symbol:   book.Symbol,
			Exchange: book.Exchange,
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
			close(c.send)
			return
		}
	}
}
