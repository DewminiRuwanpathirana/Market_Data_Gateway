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

---

## 2026-03-24 — Graceful shutdown and config symbol setup

**Goal:** Fix graceful shutdown on Ctrl+C, remove hardcoded symbols from main, wire snapshot through the pipeline correctly and implement interfaces.

**What worked:**
- Defined 'BookStore' interface in the server package (consumer side) >> server no need to imports concrete '*orderbook.Manager'.
- Implement greacefull shutdown.
- Configured symbol setup - 'config.Symbols' slice include both exchange and symbols >> can easily add new symbols from config.go >> no need to modify main.
- Snapshot sent through 'out' channel inside 'StreamUpdates' after REST fetch >> manager gets initial snapshot through the same pipeline as live updates, no separate snapshot call needed in main.

**What broke (and why):**
- 'http.ListenAndServe' blocked main forever >> Ctrl+C cancelled context but HTTP server kept running >> replaced with 'http.Server' struct and implement shutdown explicitly.
- 'FetchSnapshot' called in main inside 'StreamUpdates' >> wrong order (snapshot before WS open) >> removed from main >> snapshot now getting from 'out' channel inside 'StreamUpdates' after WS is open and buffering.

**Concept unlocked:**
- Channel - only the sender should close a channel. 
'updates' channel is created in main and used from pipeline, so main should close it after 'pipeline.Run' returns. 'srv.Run' only receiving that channel.

**Still fuzzy:**
- Whether 'httpSrv.Shutdown' needs a timeout context for production use

**Next:**
- Complete the implementation of downstream websocket client
- Implement Kraken WebSocket StreamUpdates

---

## 2026-03-25 — Per client symbol subscription on websocket connection and order book key redesign

**Goal:** Implement per client symbol subscription >> each client receives snapshot and updates only for requested symbol.

**What worked:**
- Implement per client symbol subscription.
- created struct to keep the orderbook key (exchange and symbol) since downstram clients should get only subscribed symbol's updates.
- Implement filtering requests in broadcast using 'c.exchange' and 'symbol' >> each client only receives updates for their subscribed symbol

**What broke (and why):**
- 'GetAll()' removed from 'BookStore' interface after switching to per-client subscription >> server no longer needs all books on connect, only the requested one >> replaced with 'GetBook'

**Concept unlocked:**
- 'http.Handler' interface - any type with 'ServeHTTP(ResponseWriter, *Request)' method satisfies it.
'http.Handle("/ws", srv)' registers the server and HTTP router calls 'srv.ServeHTTP' automatically on every incoming request.
```
type Handler interface {
    ServeHTTP(ResponseWriter, *Request)
}
```
- Struct keys in Go maps -- structs with only comparable fields (strings, ints) can be used as map keys directly. 
- If any method on a struct uses a pointer receiver, best practice is to make all the usages as pointer receivers for that struct.

**Still fuzzy:**
- Slow client - when 'c.send' channel is full during broadcast, client is removed and disconnect from server without blocking the entire broadcast loop. What is the correct way of handling slow clients?

**Next:**
- Implement Kraken WebSocket StreamUpdates

---

## 2026-03-27 — Kraken WebSocket Implementation

**Goal:** Implement Kraken WebSocket StreamUpdates 

**What worked:**
- Kraken WS connected and receiving book snapshots and updates correctly.

**What broke (and why):**
- Kraken sends bids/asks as named objects '{"price": 66484.6, "qty": 0.78743900}' unlike Binance (binance sends as arrays like this '["66484.6", "0.78743900"]') >> used 'krakenLevel' struct to unmarshal >> then converted to 'map[string]string' to match internal format.
- Kraken sends bids/asks as 'float64' but internal 'OrderBook' uses 'map[string]string' >> had to convert using 'strconv.FormatFloat(price, 'f', -1, 64)' to get clean decimal strings.

**Concept unlocked:**
- Go interface implicit implementation — 'BinanceConnection' and 'KrakenConnection' both satisfy 'Streamer' just by having 'StreamUpdates' with the matching signature. No explicit declaration needed.

**Still fuzzy:**
- Kraken websocket sends both snapshot and update on the same connection >> need to check whether explicit REST snapshot fetch is still needed or WS connection is enough.

**Next:**
- Complete the Kraken websocket implementation.
- Move hardcoded symbols, URL, port to config file.

---

## 2026-03-30 — Config file, subscription validation, Snapshot handling

**Goal:** Move hardcoded config to 'config.yaml', handle subscription validation, fix snapshot handling.

**What worked:**
- Moved symbols, URLs, and port to 'config.yaml' >> now 'config.go' has structs and 'Load(path)' function >> 'main.go' loads config at startup.
- Exchange constructors now use URLs as parameters instead of hardcoding.
- Added 'IsSnapshot' flag to 'types.Update' >> 'ApplyUpdate' now replace entire book on snapshot instead of merging >> fix stale data on WS reconnection.
- Added subscription validation in 'ServeHTTP' >> bad JSON sends error message back to client instead of disconnecting.
- Added 'omitempty' to 'Message' struct fields >> error message not include empty bids/asks/exchange fields.

**What broke (and why):**
- 'ApplyUpdate' always merged incoming data into the existing order book >> kraken send fresh snapshot when reconnects but old price levels remained in order book >> stale data in order book >> fixed by adding 'IsSnapshot' flag >> snapshot now replaces the entire book instead of merging.

**Concept unlocked:**
- Hardcoded config requires recompile to change symbols or URLs >> moving to 'config.yaml' >> only require edit the file and restart, no code change needed.

**Still fuzzy:**
- buffer sizes for channels

**Next:**
- Add tests
- Add DESIGN.md

---


