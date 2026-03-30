package types

type OrderBook struct {
	Symbol       string
	Exchange     string
	Bids         map[string]string // price -> quantity
	Asks         map[string]string // price -> quantity
	Timestamp    int64
	LastUpdateID int64
}

type Update struct {
	IsSnapshot   bool
	Symbol       string
	Exchange     string
	Bids         map[string]string // price -> quantity
	Asks         map[string]string // price -> quantity
	Timestamp    int64
	LastUpdateID int64
}
