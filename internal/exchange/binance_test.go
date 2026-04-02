package exchange

import (
	"testing"
)

func TestToUpdate(t *testing.T) {
	tests := []struct {
		name         string
		symbol       string
		raw          binanceDepthUpdate
		expectedBids map[string]string
		expectedAsks map[string]string
	}{
		{
			name:   "single bid and ask",
			symbol: "btcusdt",
			raw: binanceDepthUpdate{
				EventTime:    1000,
				LastUpdateID: 10,
				Bids:         [][]string{{"100.0", "1.5"}},
				Asks:         [][]string{{"101.0", "2.0"}},
			},
			expectedBids: map[string]string{"100.0": "1.5"},
			expectedAsks: map[string]string{"101.0": "2.0"},
		},
		{
			name:   "multiple bids and asks",
			symbol: "ethusdt",
			raw: binanceDepthUpdate{
				EventTime:    2000,
				LastUpdateID: 20,
				Bids:         [][]string{{"200.0", "1"}, {"199.0", "2"}},
				Asks:         [][]string{{"201.0", "3"}, {"202.0", "4"}},
			},
			expectedBids: map[string]string{"200.0": "1", "199.0": "2"},
			expectedAsks: map[string]string{"201.0": "3", "202.0": "4"},
		},
		{
			name:   "invalid input",
			symbol: "btcusdt",
			raw: binanceDepthUpdate{
				EventTime:    3000,
				LastUpdateID: 30,
				Bids:         [][]string{{"100.0"}},
				Asks:         [][]string{{"101.0"}},
			},
			expectedBids: map[string]string{},
			expectedAsks: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			update := toUpdate(tt.symbol, tt.raw)

			if len(update.Bids) != len(tt.expectedBids) {
				t.Errorf("expected %d bids, got %d", len(tt.expectedBids), len(update.Bids))
			}

			if len(update.Asks) != len(tt.expectedAsks) {
				t.Errorf("expected %d asks, got %d", len(tt.expectedAsks), len(update.Asks))
			}

			if update.Symbol != tt.symbol {
				t.Errorf("expected symbol %s, got %s", tt.symbol, update.Symbol)
			}

			if update.LastUpdateID != tt.raw.LastUpdateID {
				t.Errorf("expected LastUpdateID %d, got %d", tt.raw.LastUpdateID, update.LastUpdateID)
			}
		})
	}
}

func TestBinanceSeqAccept(t *testing.T) {
	tests := []struct {
		name          string
		seq           binanceSeq
		input         binanceDepthUpdate
		expectedValid bool
		expectedGap   bool
	}{
		{
			name: "old update",
			seq:  binanceSeq{lastID: 100},
			input: binanceDepthUpdate{
				LastUpdateID: 90,
			},
			expectedValid: false,
			expectedGap:   false,
		},
		{
			name: "gap",
			seq: binanceSeq{
				lastID:      100,
				prevU:       110,
				initialized: true,
			},
			input: binanceDepthUpdate{
				FirstUpdateID: 120, // should be 111
				LastUpdateID:  130,
			},
			expectedValid: false,
			expectedGap:   true,
		},
		{
			name: "correct sequence",
			seq: binanceSeq{
				lastID:      100,
				prevU:       110,
				initialized: true,
			},
			input: binanceDepthUpdate{
				FirstUpdateID: 111,
				LastUpdateID:  120,
			},
			expectedValid: true,
			expectedGap:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, gap := tt.seq.accept(tt.input)

			if valid != tt.expectedValid {
				t.Errorf("valid mismatch: got %v, expected %v", valid, tt.expectedValid)
			}

			if gap != tt.expectedGap {
				t.Errorf("gap mismatch: got %v, expected %v", gap, tt.expectedGap)
			}
		})
	}
}
