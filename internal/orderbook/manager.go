package orderbook

import (
	"strconv"
	"sync"

	"market-data-gateway/pkg/types"
)

type bookKey struct {
	Exchange string
	Symbol   string
}

type Manager struct {
	books map[bookKey]*types.OrderBook
	mu    sync.RWMutex
}

func NewManager() *Manager {
	return &Manager{
		books: make(map[bookKey]*types.OrderBook),
	}
}

func (m *Manager) InitSymbol(symbol, exchange string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.books[bookKey{Exchange: exchange, Symbol: symbol}] = &types.OrderBook{
		Symbol:   symbol,
		Exchange: exchange,
		Bids:     make(map[string]string),
		Asks:     make(map[string]string),
	}
}

func (m *Manager) ApplyUpdate(u types.Update) {
	m.mu.Lock()
	defer m.mu.Unlock()

	book, ok := m.books[bookKey{Exchange: u.Exchange, Symbol: u.Symbol}]
	if !ok {
		return
	}

	if u.IsSnapshot { // replace the entire book
		book.Bids = u.Bids
		book.Asks = u.Asks
	} else { // apply incremental updates
		applyLevels(book.Bids, u.Bids)
		applyLevels(book.Asks, u.Asks)
	}
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

func (m *Manager) GetBook(exchange, symbol string) *types.OrderBook {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.books[bookKey{Exchange: exchange, Symbol: symbol}]
}
