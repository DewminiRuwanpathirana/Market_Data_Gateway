package orderbook

import (
	"strconv"
	"sync"

	"market-data-gateway/pkg/types"
)

type Manager struct {
	books map[string]*types.OrderBook
	mu    sync.RWMutex
}

func NewManager() *Manager {
	return &Manager{
		books: make(map[string]*types.OrderBook), // symbol -> order book
	}
}

func (m *Manager) SetSnapshot(book *types.OrderBook) {
	m.mu.Lock()
	m.books[book.Symbol] = book
	m.mu.Unlock()
}

func (m *Manager) ApplyUpdate(u types.Update) {
	m.mu.Lock()
	defer m.mu.Unlock()

	book, ok := m.books[u.Symbol]
	if !ok {
		return
	}

	applyLevels(book.Bids, u.Bids)
	applyLevels(book.Asks, u.Asks)
	book.Timestamp = u.Timestamp
	book.LastUpdateID = u.LastUpdateID
}

func applyLevels(book map[string]string, updates map[string]string) {
	for price, qty := range updates {
		q, _ := strconv.ParseFloat(qty, 64)
		if q == 0 {
			delete(book, price)
		} else {
			book[price] = qty
		}
	}
}

func (m *Manager) GetBook(symbol string) *types.OrderBook {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.books[symbol]
}

func (m *Manager) GetAll() []*types.OrderBook {
	m.mu.RLock()
	defer m.mu.RUnlock()
	books := make([]*types.OrderBook, 0, len(m.books))
	for _, b := range m.books {
		books = append(books, b)
	}
	return books
}
