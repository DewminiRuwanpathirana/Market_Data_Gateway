package config

type SymbolConfig struct {
	Exchange string
	Symbol   string
}

var Symbols = []SymbolConfig{
	{Exchange: "binance", Symbol: "BTCUSDT"},
	{Exchange: "binance", Symbol: "SOLUSDT"},
	{Exchange: "kraken", Symbol: "BTC/USD"},
	{Exchange: "kraken", Symbol: "SOL/USD"},
}
