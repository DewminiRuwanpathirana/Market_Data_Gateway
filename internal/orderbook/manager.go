package orderbook

import (
	"sync"

	"market-data-gateway/pkg/types"
)

type Manager struct {
	books map[string]*types.OrderBook
	mu    sync.RWMutex
}

func NewManager() *Manager {
	return &Manager{
		books: make(map[string]*types.OrderBook),
	}
}

func (m *Manager) SetSnapshot(book *types.OrderBook) {
	key := book.Symbol

	m.mu.Lock()
	m.books[key] = book
	m.mu.Unlock()
}
