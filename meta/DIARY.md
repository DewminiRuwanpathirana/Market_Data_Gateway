## 2026-03-20 — Binance WebSocket order book updates and ApplyUpdate

**Goal:** Implement live order book updates from Binance WebSocket

**What worked:**
- Binance WS depth stream connected and receiving updates correctly.
- Pipeline design with Streamer interface on the consumer side.

**What broke (and why):**
- Compilation errors after changing asks and bids types >> Had to update both binance.go and kraken.go to use make(map[string]string) and map them into other files.

**Concept unlocked:**
- Stored bids/asks as float64 in OrderBook >> float precision issues when updating prices. Fixed by keeping them as string (map[string]string) and only converting to float when sending to clients
- qty == 0 means remove that price level from the book. cant compare string value for zero because Binance sends "0.00000000" not "0" >> convert to float just for this zero check and then can delete from map

**Still fuzzy:**
- orderbook asks and bids types

**Next:**
- Implement Kraken WebSocket StreamUpdates
