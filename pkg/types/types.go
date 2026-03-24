package types

type PriceLevel struct {
	Price    float64 `json:"price"`
	Quantity float64 `json:"quantity"`
}

type OrderBook struct {
	Symbol       string
	Bids         map[string]string
	Asks         map[string]string
	Timestamp    int64
	LastUpdateID int64
}

type Update struct {
	Symbol       string
	Exchange     string
	Bids         map[string]string
	Asks         map[string]string
	Timestamp    int64
	LastUpdateID int64
}
