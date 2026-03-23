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


## 2026-03-23 — Downstream WebSocket server

**Goal:** Build the WebSocket server >> downstream clients can connect and receive order book data.

**What worked:**
- Server design with one send channel for each client and a single broadcast loop.
- manager.ApplyUpdate wired into server.Run so the state stays uptodate for late joining clients.
- closeOnce pattern to safely close client channel from both the read loop and broadcast without panic.

**What broke (and why):**
- Server.Run was broadcasting updates but not updates >> manager state not uptodate after snapshot >> late joining clients got stale book >> fixed by applying update before broadcasting.

**Concept unlocked:**
- In a fan-out broadcast, the producer must never block on a slow consumer.
- A channel can only be closed once. If two goroutines both reach close(ch), the second panics. sync.Once is the fix.

**Still fuzzy:**
- Per client symbol subscription >> what is the industry standard

**Next:**
- Complete the implementation of downstream websocket client
- Implement Kraken WebSocket
