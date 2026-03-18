package types

type PriceLevel struct {
	Price    float64 `json:"price"`
	Quantity float64 `json:"quantity"`
}

type OrderBook struct {
	Symbol       string       `json:"symbol"`
	Exchange     string       `json:"exchange"`
	Bids         []PriceLevel `json:"bids"`
	Asks         []PriceLevel `json:"asks"`
	Timestamp    int64        `json:"timestamp"`
	LastUpdateID int64        `json:"last_update_id"`
}

type Update struct {
	Symbol       string       `json:"symbol"`
	Exchange     string       `json:"exchange"`
	Bids         []PriceLevel `json:"bids"`
	Asks         []PriceLevel `json:"asks"`
	Timestamp    int64        `json:"timestamp"`
	LastUpdateID int64        `json:"last_update_id"`
}
