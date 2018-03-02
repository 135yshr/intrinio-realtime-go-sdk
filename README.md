# Intrinio Go SDK for Real-Time Stock Prices

[Intrinio](https://intrinio.com/) provides real-time stock prices via a two-way WebSocket connection.To get started, [subscribe to a real-time data feed](https://intrinio.com/marketplace/data/prices/realtime) and follow the instructions below.

## Requirements

- Go 1.8+
- [gollira websocket](https://github.com/gorilla/websocket)

## Features

- Receive streaming, real-time price quotes (last trade, bid, ask)
- Subscribe to updates from individual securities
- Subscribe to updates for all securities (contact us for special access)

## Installation

```bash
go get github.com/135yshr/intrinio-realtime-go-sdk
go install github.com/135yshr/intrinio-realtime-go-sdk
```

## Example Usage

```Go
package main

import (
  "encoding/json"
  "fmt"
  "time"

  realtime "github.com/135yshr/intrinio-realtime-go-sdk"
)

const (
  cUserName = "YOUR_INTRINIO_API_USERNAME"
  cPassword = "YOUR_INTRINIO_API_PASSWORD"
)

var (
  quoteCount int
  tradeCount int
)

func main() {
  fmt.Println("start!")
  c := realtime.New(cUserName, cPassword, realtime.IEX)
  c.OnQuote(quoteHandler)
  c.OnError(errorHandler)

  if err := c.Connect(); err != nil {
    fmt.Println(err)
    return
  }
  defer c.Disconnect()
  fmt.Println("connected!")

  c.Join("AAPL")
  time.Sleep(10 * time.Minute)
  c.LeaveAll()
}

func quoteHandler(d map[string]interface{}) {
  j, err := json.Marshal(d)
  if err != nil {
    fmt.Println(err)
  }
  fmt.Println(string(j))

  event, ok := d["event"].(string)
  if ok && event == "quote" {
    quoteCount++
  } else if ok && event == "trade" {
    tradeCount++
  }
}

func errorHandler(err error) {
  fmt.Printf("It is serious! An error has occurred!! error = %v\n", err)
}
```

## Providers

Currently, Intrinio offers realtime data from the following providers:

- IEX - [Homepage](https://iextrading.com/)
- QUODD - [Homepage](http://home.quodd.com/)

Each has distinct price channels and quote formats, but a very similar API.

Each data provider has a different format for their quote data.

### QUODD

NOTE: Messages from QUOOD reflect _changes_ in market data. Not all fields will be present in every message. Upon subscribing to a channel, you will receive one quote and one trade message containing all fields of the latest data available.

#### Trade Message

```json
{ "ticker": "AAPL.NB",
  "root_ticker": "AAPL",
  "protocol_id": 301,
  "last_price_4d": 1594850,
  "trade_volume": 100,
  "trade_exchange": "t",
  "change_price_4d": 24950,
  "percent_change_4d": 15892,
  "trade_time": 1508165070052,
  "up_down": "v",
  "vwap_4d": 1588482,
  "total_volume": 10209883,
  "day_high_4d": 1596600,
  "day_high_time": 1508164532269,
  "day_low_4d": 1576500,
  "day_low_time": 1508160605345,
  "prev_close_4d": 1569900,
  "volume_plus": 6333150,
  "ext_last_price_4d": 1579000,
  "ext_trade_volume": 100,
  "ext_trade_exchange": "t",
  "ext_change_price_4d": 9100,
  "ext_percent_change_4d": 5796,
  "ext_trade_time": 1508160600567,
  "ext_up_down": "-",
  "open_price_4d": 1582200,
  "open_volume": 100,
  "open_time": 1508141103583,
  "rtl": 30660,
  "is_halted": false,
  "is_short_restricted": false }
```

- **ticker** - Stock Symbol for the security
- **root_ticker** - Underlying symbol for a particular contract
- **last_price_4d** - The price at which the security most recently traded
- **trade_volume** - The number of shares that that were traded on the last trade
- **trade_exchange** - The market center where the last trade occurred
- **trade_time** - The time at which the security last traded in milliseconds
- **up_down** - Tick indicator - up or down - indicating if the last trade was up or down from the previous trade
- **change_price_4d** - The difference between the closing price of a security on the current trading day and the previous day's closing price.
- **percent_change_4d** - The percentage at which the security is up or down since the previous day's trading
- **total_volume** - The accumulated total amount of shares traded
- **volume_plus** - NASDAQ volume plus the volumes from other market centers to more accurately match composite volume. Used for NASDAQ Basic
- **vwap_4d** - Volume weighted Average Price. VWAP is calculated by adding up the dollars traded for every transaction (price multiplied by number of shares traded) and then dividing by the total shares traded for the day.
- **day_high_4d** - A security's intra-day high trading price.
- **day_high_time** - Time that the security reached a new high
- **day_low_4d** - A security's intra-day low trading price.
- **day_low_time** - Time that the security reached a new low
- **ext_last_price_4d** - Extended hours last price (pre or post market)
- **ext_trade_volume** - The amount of shares traded for a single extended hours trade
- **ext_trade_exchange** - Extended hours exchange where last trade took place (Pre or post market)
- **ext_trade_time** - Time of the extended hours trade in milliseconds
- **ext_up_down** - Extended hours tick indicator - up or down
- **ext_change_price_4d** - Extended hours change price (pre or post market)
- **ext_percent_change_4d** - Extended hours percent change (pre or post market)
- **is_halted** - A flag indicating that the stock is halted and not currently trading
- **is_short_restricted** - A flag indicating the stock is current short sale restricted - meaning you can not short sale the stock when true
- **open_price_4d** - The price at which a security first trades upon the opening of an exchange on a given trading day
- **open_time** - The time at which the security opened in milliseconds
- **open_volume** - The number of shares that that were traded on the opening trade
- **prev_close_4d** - The security's closing price on the preceding day of trading
- **protocol_id** - Internal Quodd ID defining Source of Data
- **rtl** - Record Transaction Level - number of records published that day

#### Quote Message

```json
{ "ticker": "AAPL.NB",
  "root_ticker": "AAPL",
  "bid_size": 500,
  "ask_size": 600,
  "bid_price_4d": 1594800,
  "ask_price_4d": 1594900,
  "ask_exchange": "t",
  "bid_exchange": "t",
  "quote_time": 1508165070850,
  "protocol_id": 302,
  "rtl": 129739 }
```

- **ticker** - Stock Symbol for the security
- **root_ticker** - Underlying symbol for a particular contract
- **ask_price_4d** - The price a seller is willing to accept for a security
- **ask_size** - The amount of a security that a market maker is offering to sell at the ask price
- **ask_exchange** - The market center from which the ask is being quoted
- **bid_price_4d** - A bid price is the price a buyer is willing to pay for a security.
- **bid_size** - The bid size number of shares being offered for purchase at a specified bid price
- **bid_exchange** - The market center from which the bid is being quoted
- **quote_time** - Time of the quote in milliseconds
- **rtl** - Record Transaction Level - number of records published that day
- **protocol_id** - Internal Quodd ID defining Source of Data

### IEX

```json
{ "type": "ask",
  "timestamp": 1493409509.3932788,
  "ticker": "GE",
  "size": 13750,
  "price": 28.97 }
```

- **type** - the quote type
  - **`last`** - represents the last traded price
  - **`bid`** - represents the top-of-book bid price
  - **`ask`** - represents the top-of-book ask price
- **timestamp** - a Unix timestamp (with microsecond precision)
- **ticker** - the ticker of the security
- **size** - the size of the `last` trade, or total volume of orders at the top-of-book `bid` or `ask` price
- **price** - the price in USD

## Channels

### QUODD

To receive price quotes from QUODD, you need to instruct the client to "join" a channel. A channel can be

- A security ticker with data feed designation (`AAPL.NB`, `MSFT.NB`, `GE.NB`, etc)

### IEX

To receive price quotes from the Intrinio Real-Time API, you need to instruct the client to "join" a channel. A channel can be

- A security ticker (`AAPL`, `MSFT`, `GE`, etc)
- The security lobby (`$lobby`) where all price quotes for all securities are posted
- The security last price lobby (`$lobby_last_price`) where only last price quotes for all securities are posted

Special access is required for both lobby channels. [Contact us](mailto:sales@intrinio.com) for more information.

## API Keys

You will receive your Intrinio API Username and Password after [creating an account](https://intrinio.com/signup). You will need a subscription to the [IEX Real-Time Stock Prices](https://intrinio.com/data/realtime-stock-prices) data feed as well.

## Documentation

### Methods

`New(options)` - Creates a new instance of the IntrinioRealtime client.

- **Parameter** `username`: Your Intrinio API Username
- **Parameter** `password`: Your Intrinio API Password
- **Parameter** `provider`: The real-time data provider to use (IEX, QUODD)

```Go
client := realtime.New("INTRINIO_API_USERNAME", "INTRINIO_API_PASSWORD", realtime.IEX)
```

---------

`client.Connect()` - Opens the WebSocket connection and joins the requested channels. This method blocks indefinitely.

---------

`client.Disconnect()` - Closes the WebSocket, stops the self-healing and heartbeat intervals. You must call this to dispose of the client.

---------

`client.OnQuote(f func(map[string]interface{}))` - Adds a QuoteHandler for handling quotes. Each quote handler will wait to receive a quote from the client's queue. Note that all quote handlers will not receive all quotes. Each handler receives the next quote in the queue once the handler finishes handling its current quote. Register multiple quote handlers to handle quotes quicker in cases of I/O.

- **Parameter** `data` -  The data to invoke. The quote will be passed as an argument to the data.

```Go
client.OnQuote(func(data map[string]interface{}) {
  fmt.Println(data)
})
```

---------

`client.OnError(f func(err error))` - Invokes the given callback when a fatal error is encountered. If no callback has been registered and no `error` event listener has been registered, the error will be thrown.

- **Parameter** `err` - The callback to invoke. The error will be passed as an argument to the callback.

```Go
client.OnError(func(err error) {
  fmt.Println(err)
})
```

---------

`client.Join(channels ...string)` - Joins the given channels. This can be called at any time. The client will automatically register joined channels and establish the proper subscriptions with the WebSocket connection.

- **Parameter** `channels` - An argument list or array of channels to join. See Channels section above for more details.

```Go
client.Join("AAPL", "MSFT", "GE")
client.Join("$lobby")
```

---------

`client.Leave(channels ...string)` - Leaves the given channels.

- **Parameter** `channels` - An argument list or array of channels to leave.

```Go
client.Leave("AAPL", "MSFT", "GE")
client.Leave("$lobby")
```

---------

`client.LeaveAll()` - Leaves all joined channels.
