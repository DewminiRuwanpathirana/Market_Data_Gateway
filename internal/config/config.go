package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server    ServerConfig              `yaml:"server"`
	Exchanges map[string]ExchangeConfig `yaml:"exchanges"`
	Symbols   []SymbolConfig            `yaml:"symbols"`
}

type ServerConfig struct {
	Port int `yaml:"port"`
}

type ExchangeConfig struct {
	BaseURL string `yaml:"base_url"`
	WsURL   string `yaml:"ws_url"`
}

type SymbolConfig struct {
	Exchange string `yaml:"exchange"`
	Symbol   string `yaml:"symbol"`
}

func Load(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open config: %w", err)
	}
	defer f.Close()

	var cfg Config
	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	return &cfg, nil
}
