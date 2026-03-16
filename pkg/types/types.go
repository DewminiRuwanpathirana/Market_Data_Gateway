package types

type PriceLevel struct {
	Price    float64
	Quantity float64
}

type OrderBook struct {
	Symbol    string
	Exchange  string
	Bids      []PriceLevel
	Asks      []PriceLevel
	Timestamp int64
}
