# Tradernet SDK Microservice - API Endpoints

Complete documentation for all HTTP endpoints exposed by the Tradernet SDK microservice.

## Authentication

All endpoints (except plain requests) require authentication via HTTP headers:
- `X-Tradernet-API-Key`: Your Tradernet public API key
- `X-Tradernet-API-Secret`: Your Tradernet private API secret

**Plain requests** (no authentication required):
- `GET /most-traded`
- `GET /find-symbol`
- `POST /new-user`

## Response Format

All endpoints return JSON in the following format:

**Success:**
```json
{
  "success": true,
  "data": { ... },
  "error": null
}
```

**Error:**
```json
{
  "success": false,
  "error": "Error message"
}
```

---

## Health & Status

### `GET /health`

Health check endpoint. No authentication required.

**Response:**
```json
{
  "status": "ok"
}
```

---

## User & Account

### `GET /user-info`

Retrieves user information from the Tradernet API.

**Headers:**
- `X-Tradernet-API-Key`: Required
- `X-Tradernet-API-Secret`: Required

**Response:**
```json
{
  "success": true,
  "data": {
    "id": 12345,
    "login": "username",
    "first_name": "John",
    "last_name": "Doe",
    "email": "user@example.com"
  }
}
```

**API Reference:** https://freedom24.com/tradernet-api/get-user-info

---

### `GET /account-summary`

Retrieves account summary including positions and cash balances.

**Headers:**
- `X-Tradernet-API-Key`: Required
- `X-Tradernet-API-Secret`: Required

**Response:**
```json
{
  "success": true,
  "data": {
    "result": {
      "ps": {
        "pos": [
          {
            "i": "AAPL.US",
            "q": 10,
            "bal_price_a": 150.0,
            "mkt_price": 155.0,
            "profit_close": 50.0,
            "curr": "USD"
          }
        ],
        "acc": [
          {
            "curr": "USD",
            "s": 1000.0
          }
        ]
      }
    }
  }
}
```

**API Reference:** https://freedom24.com/tradernet-api/portfolio-get-changes

---

### `GET /user-data`

Retrieves initial user data including orders, portfolio, markets, and open sessions.

**Headers:**
- `X-Tradernet-API-Key`: Required
- `X-Tradernet-API-Secret`: Required

**Response:**
```json
{
  "success": true,
  "data": {
    "orders": [...],
    "portfolio": {...},
    "markets": {...},
    "sessions": [...]
  }
}
```

**API Reference:** https://freedom24.com/tradernet-api/auth-get-opq

---

## Trading Operations

### `POST /buy`

Places a buy order for the specified symbol.

**Headers:**
- `X-Tradernet-API-Key`: Required
- `X-Tradernet-API-Secret`: Required

**Request Body:**
```json
{
  "symbol": "AAPL.US",
  "quantity": 10,
  "price": 150.0,
  "duration": "day",
  "use_margin": true,
  "custom_order_id": null
}
```

**Parameters:**
- `symbol` (string, required): Tradernet symbol (e.g., "AAPL.US")
- `quantity` (int, required): Number of shares to buy (must be positive)
- `price` (float, optional): Limit price (0.0 for market order, default: 0.0)
- `duration` (string, optional): Order duration - "day", "ext", "gtc", or "ioc" (default: "day")
- `use_margin` (bool, optional): Whether to use margin credit (default: true)
- `custom_order_id` (int, optional): Custom order ID (null to auto-generate)

**Response:**
```json
{
  "success": true,
  "data": {
    "order_id": 12345,
    "price": 150.0
  }
}
```

**API Reference:** https://freedom24.com/tradernet-api/orders-place

---

### `POST /sell`

Places a sell order for the specified symbol.

**Headers:**
- `X-Tradernet-API-Key`: Required
- `X-Tradernet-API-Secret`: Required

**Request Body:**
```json
{
  "symbol": "AAPL.US",
  "quantity": 10,
  "price": 150.0,
  "duration": "day",
  "use_margin": true,
  "custom_order_id": null
}
```

**Parameters:** Same as `/buy`

**Response:** Same format as `/buy`

**API Reference:** https://freedom24.com/tradernet-api/orders-place

---

### `GET /pending-orders`

Retrieves pending/active orders for the current period.

**Headers:**
- `X-Tradernet-API-Key`: Required
- `X-Tradernet-API-Secret`: Required

**Query Parameters:**
- `active` (bool, optional): If true, returns only active orders. If false, returns all orders (default: true)

**Example:**
```
GET /pending-orders?active=true
```

**Response:**
```json
{
  "success": true,
  "data": {
    "result": {
      "orders": {
        "order": [
          {
            "id": 12345,
            "instr_name": "AAPL.US",
            "buy_sell": "BUY",
            "qty": 10,
            "price": 150.0,
            "curr": "USD"
          }
        ]
      }
    }
  }
}
```

**API Reference:** https://freedom24.com/tradernet-api/orders-get-current-history

---

### `POST /cancel-order`

Cancels an active order by order ID.

**Headers:**
- `X-Tradernet-API-Key`: Required
- `X-Tradernet-API-Secret`: Required

**Request Body:**
```json
{
  "order_id": 12345
}
```

**Parameters:**
- `order_id` (int, required): Order ID to cancel

**Response:**
```json
{
  "success": true,
  "data": {
    "result": "cancelled",
    "order_id": 12345
  }
}
```

**API Reference:** https://freedom24.com/tradernet-api/orders-cancel

---

### `POST /cancel-all`

Cancels all active orders.

**Headers:**
- `X-Tradernet-API-Key`: Required
- `X-Tradernet-API-Secret`: Required

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "result": "cancelled",
      "order_id": 12345
    },
    {
      "result": "cancelled",
      "order_id": 12346
    }
  ]
}
```

---

### `POST /stop`

Places a stop loss order on an open position.

**Headers:**
- `X-Tradernet-API-Key`: Required
- `X-Tradernet-API-Secret`: Required

**Request Body:**
```json
{
  "symbol": "AAPL.US",
  "price": 140.0
}
```

**Parameters:**
- `symbol` (string, required): Tradernet symbol
- `price` (float, required): Stop loss price

**Response:**
```json
{
  "success": true,
  "data": {
    "order_id": 12345
  }
}
```

**API Reference:** https://freedom24.com/tradernet-api/orders-stop-loss

---

### `POST /trailing-stop`

Places a trailing stop order on an open position.

**Headers:**
- `X-Tradernet-API-Key`: Required
- `X-Tradernet-API-Secret`: Required

**Request Body:**
```json
{
  "symbol": "AAPL.US",
  "percent": 5
}
```

**Parameters:**
- `symbol` (string, required): Tradernet symbol
- `percent` (int, optional): Stop loss percentage (default: 1)

**Response:**
```json
{
  "success": true,
  "data": {
    "order_id": 12345
  }
}
```

**API Reference:** https://freedom24.com/tradernet-api/orders-stop-loss

---

### `POST /take-profit`

Places a take profit order on an open position.

**Headers:**
- `X-Tradernet-API-Key`: Required
- `X-Tradernet-API-Secret`: Required

**Request Body:**
```json
{
  "symbol": "AAPL.US",
  "price": 200.0
}
```

**Parameters:**
- `symbol` (string, required): Tradernet symbol
- `price` (float, required): Take profit price

**Response:**
```json
{
  "success": true,
  "data": {
    "order_id": 12345
  }
}
```

**API Reference:** https://freedom24.com/tradernet-api/orders-stop-loss

---

### `GET /orders-history`

Retrieves historical orders for the specified period.

**Headers:**
- `X-Tradernet-API-Key`: Required
- `X-Tradernet-API-Secret`: Required

**Query Parameters:**
- `start` (string, optional): Start date in RFC3339 or "2006-01-02T15:04:05" format (default: 2011-01-11T00:00:00)
- `end` (string, optional): End date in RFC3339 or "2006-01-02T15:04:05" format (default: now)

**Example:**
```
GET /orders-history?start=2023-01-01T00:00:00&end=2023-12-31T23:59:59
```

**Response:**
```json
{
  "success": true,
  "data": {
    "result": {
      "orders": [...]
    }
  }
}
```

**API Reference:** https://freedom24.com/tradernet-api/get-orders-history

---

## Transactions & Reports

### `GET /trades-history`

Retrieves executed trades history.

**Headers:**
- `X-Tradernet-API-Key`: Required
- `X-Tradernet-API-Secret`: Required

**Query Parameters:**
- `start` (string, optional): Start date in ISO format (YYYY-MM-DD) (default: 1970-01-01)
- `end` (string, optional): End date in ISO format (YYYY-MM-DD) (default: today)
- `trade_id` (int, optional): Trade ID to start from (pagination)
- `limit` (int, optional): Maximum number of trades (0 or omitted for all)
- `symbol` (string, optional): Symbol filter
- `currency` (string, optional): Currency filter

**Example:**
```
GET /trades-history?start=2023-01-01&end=2023-12-31&limit=100
```

**Response:**
```json
{
  "success": true,
  "data": {
    "trades": {
      "trade": [
        {
          "id": 12345,
          "instr_nm": "AAPL.US",
          "q": 10,
          "price": 150.0,
          "type": "1",
          "date": "2023-01-15"
        }
      ]
    }
  }
}
```

**API Reference:** https://freedom24.com/tradernet-api/get-trades-history

---

### `GET /cash-movements`

Retrieves cash movements history (withdrawals, deposits, etc.).

**Headers:**
- `X-Tradernet-API-Key`: Required
- `X-Tradernet-API-Secret`: Required

**Query Parameters:**
- `date_from` (string, optional): Start date in "2006-01-02T15:04:05" format
- `date_to` (string, optional): End date in "2006-01-02T15:04:05" format
- `cps_doc_id` (int, optional): Request type ID
- `id` (int, optional): Order ID
- `limit` (int, optional): Maximum number of records
- `offset` (int, optional): Pagination offset
- `cps_status` (int, optional): Request status (0=draft, 1=in process, 2=rejected, 3=executed)

**Example:**
```
GET /cash-movements?date_from=2023-01-01T00:00:00&date_to=2023-12-31T23:59:59&limit=500
```

**Response:**
```json
{
  "success": true,
  "data": {
    "result": [...]
  }
}
```

**API Reference:** https://freedom24.com/tradernet-api/get-client-cps-history

---

### `GET /order-files`

Retrieves order files/documents.

**Headers:**
- `X-Tradernet-API-Key`: Required
- `X-Tradernet-API-Secret`: Required

**Query Parameters:**
- `order_id` (int, optional): Order ID (required if internal_id not provided)
- `internal_id` (int, optional): Draft order ID (required if order_id not provided)

**Example:**
```
GET /order-files?order_id=12345
```

**Response:**
```json
{
  "success": true,
  "data": {
    "result": {
      "files": [...]
    }
  }
}
```

**API Reference:** https://freedom24.com/tradernet-api/get-cps-files

---

### `GET /broker-report`

Retrieves broker's report.

**Headers:**
- `X-Tradernet-API-Key`: Required
- `X-Tradernet-API-Secret`: Required

**Query Parameters:**
- `start` (string, optional): Start date in "2006-01-02" format (default: 1970-01-01)
- `end` (string, optional): End date in "2006-01-02" format (default: today)
- `period` (string, optional): Time period in "15:04:05" format (default: 23:59:59)
- `type` (string, optional): Data block type (default: "account_at_end")

**Example:**
```
GET /broker-report?start=2023-01-01&end=2023-12-31&period=23:59:59&type=account_at_end
```

**Response:**
```json
{
  "success": true,
  "data": {
    "result": {
      "report": {...}
    }
  }
}
```

**API Reference:** https://freedom24.com/tradernet-api/get-broker-report

---

## Market Data

### `POST /quotes`

Retrieves current quotes for one or more symbols.

**Headers:**
- `X-Tradernet-API-Key`: Required
- `X-Tradernet-API-Secret`: Required

**Request Body:**
```json
{
  "symbols": ["AAPL.US", "MSFT.US"]
}
```

**Parameters:**
- `symbols` (array of strings, required): Array of Tradernet symbols

**Response:**
```json
{
  "success": true,
  "data": {
    "AAPL.US": {
      "ltp": 155.0,
      "chg": 5.0,
      "chg_pc": 3.33,
      "v": 1000000
    },
    "MSFT.US": {
      "ltp": 350.0,
      "chg": -2.0,
      "chg_pc": -0.57,
      "v": 500000
    }
  }
}
```

**API Reference:** https://freedom24.com/tradernet-api/quotes-get

---

### `POST /candles`

Retrieves historical OHLC (Open, High, Low, Close) data for a symbol.

**Headers:**
- `X-Tradernet-API-Key`: Required
- `X-Tradernet-API-Secret`: Required

**Request Body:**
```json
{
  "symbol": "AAPL.US",
  "start": "2023-01-01T00:00:00Z",
  "end": "2023-12-31T23:59:59Z",
  "timeframe_seconds": 86400
}
```

**Parameters:**
- `symbol` (string, required): Tradernet symbol
- `start` (string, required): Start date/time in RFC3339 format
- `end` (string, required): End date/time in RFC3339 format
- `timeframe_seconds` (int, required): Timeframe in seconds (e.g., 86400 for daily, 3600 for hourly)

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "date": "2023-01-01",
      "o": 150.0,
      "h": 155.0,
      "l": 149.0,
      "c": 154.0,
      "v": 1000000
    }
  ]
}
```

**API Reference:**
- https://freedom24.com/tradernet-api/quotes-get-hloc
- https://freedom24.com/tradernet-api/get-trades

---

### `GET /market-status`

Retrieves market status information.

**Headers:**
- `X-Tradernet-API-Key`: Required
- `X-Tradernet-API-Secret`: Required

**Query Parameters:**
- `market` (string, optional): Market code (default: "*" for all markets)
- `mode` (string, optional): Request mode (e.g., "demo")

**Example:**
```
GET /market-status?market=NYSE
```

**Response:**
```json
{
  "success": true,
  "data": {
    "result": {
      "status": "open"
    }
  }
}
```

**API Reference:** https://freedom24.com/tradernet-api/market-status

---

### `GET /most-traded`

Retrieves most traded securities or top gainers. **No authentication required.**

**Query Parameters:**
- `instrument_type` (string, optional): Instrument type (default: "stocks")
- `exchange` (string, optional): Exchange (default: "usa")
- `gainers` (bool, optional): If true, returns top gainers; if false, returns most traded (default: true)
- `limit` (int, optional): Number of results (default: 10)

**Example:**
```
GET /most-traded?instrument_type=stocks&exchange=usa&gainers=true&limit=20
```

**Response:**
```json
{
  "success": true,
  "data": {
    "result": [
      {
        "symbol": "AAPL.US",
        "change": 5.0,
        "change_pct": 3.33
      }
    ]
  }
}
```

**API Reference:** https://freedom24.com/tradernet-api/quotes-get-top-securities

---

### `POST /export-securities`

Exports securities data in bulk. Uses direct HTTP GET to `/securities/export`.

**Headers:**
- `X-Tradernet-API-Key`: Required
- `X-Tradernet-API-Secret`: Required

**Request Body:**
```json
{
  "symbols": ["AAPL.US", "MSFT.US", "GOOGL.US"],
  "fields": ["ticker", "name", "isin"]
}
```

**Parameters:**
- `symbols` (array of strings, required): Array of symbols to export
- `fields` (array of strings, optional): Fields to include (omitted for all fields)

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "ticker": "AAPL.US",
      "name": "Apple Inc.",
      "isin": "US0378331005"
    }
  ]
}
```

**Note:** Processes symbols in chunks of 100 (MAX_EXPORT_SIZE).

**API Reference:** https://freedom24.com/tradernet-api/quotes-get

---

### `GET /news`

Retrieves news on securities.

**Headers:**
- `X-Tradernet-API-Key`: Required
- `X-Tradernet-API-Secret`: Required

**Query Parameters:**
- `query` (string, required): Search query (ticker or any word)
- `symbol` (string, optional): If provided, query is ignored and only news for this symbol is returned
- `story_id` (string, optional): If provided, query and symbol are ignored and only this story is returned
- `limit` (int, optional): Maximum number of news items (default: 30)

**Example:**
```
GET /news?query=AAPL&limit=10
```

**Response:**
```json
{
  "success": true,
  "data": {
    "result": [
      {
        "title": "Apple announces new product",
        "date": "2023-01-15",
        "storyId": "12345"
      }
    ]
  }
}
```

**API Reference:** https://freedom24.com/tradernet-api/quotes-get-news

---

## Securities

### `GET /find-symbol`

Searches for securities by symbol name or ISIN. **No authentication required.**

**Query Parameters:**
- `symbol` (string, required): Symbol name or ISIN to search for
- `exchange` (string, optional): Exchange name to filter results

**Example:**
```
GET /find-symbol?symbol=AAPL&exchange=NYSE
```

**Response:**
```json
{
  "success": true,
  "data": {
    "found": [
      {
        "t": "AAPL.US",
        "nm": "Apple Inc.",
        "isin": "US0378331005",
        "x_curr": "USD",
        "mkt": "NYSE",
        "codesub": "XNAS"
      }
    ]
  }
}
```

**API Reference:** https://freedom24.com/tradernet-api/quotes-finder

---

### `GET /security-info`

Retrieves detailed information about a specific security.

**Headers:**
- `X-Tradernet-API-Key`: Required
- `X-Tradernet-API-Secret`: Required

**Query Parameters:**
- `symbol` (string, required): Tradernet symbol
- `sup` (bool, optional): IMS and trading system format (default: true)

**Example:**
```
GET /security-info?symbol=AAPL.US&sup=true
```

**Response:**
```json
{
  "success": true,
  "data": {
    "ticker": "AAPL.US",
    "name": "Apple Inc.",
    "isin": "US0378331005",
    "lot_size": 1,
    "currency": "USD",
    "exchange": "NASDAQ"
  }
}
```

**API Reference:** https://freedom24.com/tradernet-api/quotes-get-info

---

### `GET /symbol`

Retrieves stock data (different from SecurityInfo - used for shop/display data).

**Headers:**
- `X-Tradernet-API-Key`: Required
- `X-Tradernet-API-Secret`: Required

**Query Parameters:**
- `symbol` (string, required): Tradernet symbol
- `lang` (string, optional): Language code (default: "en")

**Example:**
```
GET /symbol?symbol=AAPL.US&lang=en
```

**Response:**
```json
{
  "success": true,
  "data": {
    "symbol": "AAPL.US",
    "name": "Apple Inc.",
    "description": "..."
  }
}
```

**API Reference:** https://freedom24.com/tradernet-api/shop-get-stock-data

---

### `GET /symbols`

Retrieves ready list of securities by exchange.

**Headers:**
- `X-Tradernet-API-Key`: Required
- `X-Tradernet-API-Secret`: Required

**Query Parameters:**
- `exchange` (string, optional): Exchange name (e.g., "USA", "Russia")

**Example:**
```
GET /symbols?exchange=USA
```

**Response:**
```json
{
  "success": true,
  "data": {
    "result": {
      "symbols": [...]
    }
  }
}
```

**API Reference:** https://freedom24.com/tradernet-api/get-ready-list

---

### `GET /options`

Retrieves options by underlying asset and exchange.

**Headers:**
- `X-Tradernet-API-Key`: Required
- `X-Tradernet-API-Secret`: Required

**Query Parameters:**
- `underlying` (string, required): Underlying symbol
- `exchange` (string, required): Exchange name

**Example:**
```
GET /options?underlying=AAPL.US&exchange=CBOE
```

**Response:**
```json
{
  "success": true,
  "data": {
    "result": [
      {
        "symbol": "AAPL220121C00150000",
        "strike": 150.0,
        "expiry": "2022-01-21"
      }
    ]
  }
}
```

**API Reference:** https://freedom24.com/tradernet-api/get-options-by-mkt

---

### `GET /corporate-actions`

Retrieves planned corporate actions for a specific office.

**Headers:**
- `X-Tradernet-API-Key`: Required
- `X-Tradernet-API-Secret`: Required

**Query Parameters:**
- `reception` (int, optional): Office number (default: 35)

**Example:**
```
GET /corporate-actions?reception=35
```

**Response:**
```json
{
  "success": true,
  "data": {
    "result": [
      {
        "symbol": "AAPL.US",
        "action_type": "dividend",
        "date": "2023-03-15"
      }
    ]
  }
}
```

**API Reference:** https://freedom24.com/tradernet-api/get-planned-corp-actions

---

### `POST /get-all`

Retrieves all securities with filters. **Currently returns error - refbook functionality not yet implemented.**

**Headers:**
- `X-Tradernet-API-Key`: Required
- `X-Tradernet-API-Secret`: Required

**Request Body:**
```json
{
  "filters": {
    "mkt_short_code": "FIX",
    "instr_type_c": 1
  },
  "show_expired": false
}
```

**Response:**
```json
{
  "success": false,
  "error": "GetAll() requires refbook download functionality which is not yet implemented. Use FindSymbol() or GetQuotes() instead"
}
```

**Note:** This endpoint requires refbook download functionality (HTML parsing, ZIP extraction) which is not yet implemented. Use `FindSymbol()` or `GetQuotes()` instead.

**API Reference:**
- https://freedom24.com/tradernet-api/securities
- https://freedom24.com/tradernet-api/instruments

---

## Price Alerts

### `GET /price-alerts`

Retrieves list of price alerts.

**Headers:**
- `X-Tradernet-API-Key`: Required
- `X-Tradernet-API-Secret`: Required

**Query Parameters:**
- `symbol` (string, optional): Symbol to filter alerts

**Example:**
```
GET /price-alerts?symbol=AAPL.US
```

**Response:**
```json
{
  "success": true,
  "data": {
    "result": [
      {
        "id": 12345,
        "ticker": "AAPL.US",
        "price": 150.0,
        "trigger_type": "crossing"
      }
    ]
  }
}
```

**API Reference:** https://freedom24.com/tradernet-api/alerts-get-list

---

### `POST /add-price-alert`

Adds a price alert.

**Headers:**
- `X-Tradernet-API-Key`: Required
- `X-Tradernet-API-Secret`: Required

**Request Body:**
```json
{
  "symbol": "AAPL.US",
  "price": 150.0,
  "trigger_type": "crossing",
  "quote_type": "ltp",
  "send_to": "email",
  "frequency": 0,
  "expire": 0
}
```

**Parameters:**
- `symbol` (string, required): Symbol to add alert for
- `price` (number or array of numbers, required): Alert price(s)
- `trigger_type` (string, optional): Trigger method (default: "crossing")
- `quote_type` (string, optional): Price type - "ltp", "bap", "bbp", "op", "pp" (default: "ltp")
- `send_to` (string, optional): Notification type - "email", "sms", "push", "all" (default: "email")
- `frequency` (int, optional): Frequency (default: 0)
- `expire` (int, optional): Alert period (default: 0)

**Response:**
```json
{
  "success": true,
  "data": {
    "alert_id": 12345
  }
}
```

**API Reference:** https://freedom24.com/tradernet-api/alerts-add

---

### `POST /delete-price-alert`

Deletes a price alert.

**Headers:**
- `X-Tradernet-API-Key`: Required
- `X-Tradernet-API-Secret`: Required

**Request Body:**
```json
{
  "alert_id": 12345
}
```

**Parameters:**
- `alert_id` (int, required): Alert ID to delete

**Response:**
```json
{
  "success": true,
  "data": {
    "status": "deleted"
  }
}
```

**API Reference:** https://freedom24.com/tradernet-api/alerts-delete

---

## User Management

### `POST /new-user`

Creates a new user account. **No authentication required.**

**Request Body:**
```json
{
  "login": "newuser",
  "reception": "35",
  "phone": "1234567890",
  "lastname": "Doe",
  "firstname": "John",
  "password": "securepassword",
  "utm_campaign": null,
  "tariff_id": 1
}
```

**Parameters:**
- `login` (string, required): User login name
- `reception` (string, required): Reception number
- `phone` (string, required): User's phone number
- `lastname` (string, required): User's last name
- `firstname` (string, required): User's first name
- `password` (string, optional): Password (if omitted, will be generated automatically)
- `utm_campaign` (string, optional): Referral link
- `tariff_id` (int, optional): Tariff ID to assign during registration

**Response:**
```json
{
  "success": true,
  "data": {
    "clientId": "12345",
    "userId": "67890"
  }
}
```

**API Reference:** https://freedom24.com/tradernet-api/primary-registration

---

### `GET /check-missing-fields`

Checks for missing profile fields.

**Headers:**
- `X-Tradernet-API-Key`: Required
- `X-Tradernet-API-Secret`: Required

**Query Parameters:**
- `step` (int, required): Step number in registration/profile completion
- `office` (string, required): Office name

**Example:**
```
GET /check-missing-fields?step=1&office=office1
```

**Response:**
```json
{
  "success": true,
  "data": {
    "result": {
      "not_completed": ["field1", "field2"]
    }
  }
}
```

**API Reference:** https://freedom24.com/tradernet-api/check-step

---

### `GET /profile-fields`

Retrieves profile fields configuration for different offices.

**Headers:**
- `X-Tradernet-API-Key`: Required
- `X-Tradernet-API-Secret`: Required

**Query Parameters:**
- `reception` (int, required): Reception number (office identifier)

**Example:**
```
GET /profile-fields?reception=35
```

**Response:**
```json
{
  "success": true,
  "data": {
    "result": {
      "fields": [...]
    }
  }
}
```

**API Reference:** https://freedom24.com/tradernet-api/get-anketa-fields

---

## Other

### `GET /tariffs-list`

Retrieves list of available tariffs.

**Headers:**
- `X-Tradernet-API-Key`: Required
- `X-Tradernet-API-Secret`: Required

**Response:**
```json
{
  "success": true,
  "data": {
    "result": [
      {
        "id": 1,
        "name": "Basic",
        "price": 0.0
      }
    ]
  }
}
```

**API Reference:** https://freedom24.com/tradernet-api/get-list-tariff

---

## Error Handling

All endpoints return consistent error responses:

**HTTP Status Codes:**
- `200 OK`: Request successful
- `400 Bad Request`: Invalid request parameters or missing credentials
- `500 Internal Server Error`: Server error or API request failed

**Error Response Format:**
```json
{
  "success": false,
  "error": "Error message describing what went wrong"
}
```

**Common Errors:**
- `"Missing credentials. Provide X-Tradernet-API-Key and X-Tradernet-API-Secret headers."`: Authentication headers missing
- `"API returned status 403: Forbidden"`: Invalid credentials or insufficient permissions
- `"API returned status 500: Internal Server Error"`: Tradernet API error

---

## Examples

### Complete Workflow: Get Portfolio and Place Order

```bash
# 1. Get account summary (positions and cash)
curl -H "X-Tradernet-API-Key: $TRADERNET_API_KEY" \
     -H "X-Tradernet-API-Secret: $TRADERNET_API_SECRET" \
     http://localhost:9001/account-summary

# 2. Get current quotes
curl -X POST \
     -H "X-Tradernet-API-Key: $TRADERNET_API_KEY" \
     -H "X-Tradernet-API-Secret: $TRADERNET_API_SECRET" \
     -H "Content-Type: application/json" \
     -d '{"symbols": ["AAPL.US", "MSFT.US"]}' \
     http://localhost:9001/quotes

# 3. Place a buy order
curl -X POST \
     -H "X-Tradernet-API-Key: $TRADERNET_API_KEY" \
     -H "X-Tradernet-API-Secret: $TRADERNET_API_SECRET" \
     -H "Content-Type: application/json" \
     -d '{
       "symbol": "AAPL.US",
       "quantity": 10,
       "price": 150.0,
       "duration": "day",
       "use_margin": true
     }' \
     http://localhost:9001/buy

# 4. Check pending orders
curl -H "X-Tradernet-API-Key: $TRADERNET_API_KEY" \
     -H "X-Tradernet-API-Secret: $TRADERNET_API_SECRET" \
     http://localhost:9001/pending-orders?active=true
```

---

## Notes

1. **Date Formats:**
   - `/trades-history`: ISO format (YYYY-MM-DD)
   - `/cash-movements`: "2006-01-02T15:04:05" format
   - `/orders-history`: "2006-01-02T15:04:05" format
   - `/candles`: RFC3339 format (e.g., "2023-01-01T00:00:00Z")
   - `/broker-report`: "2006-01-02" for dates, "15:04:05" for time

2. **Plain Requests:**
   - `/most-traded`, `/find-symbol`, `/new-user` do not require authentication
   - All other endpoints require `X-Tradernet-API-Key` and `X-Tradernet-API-Secret` headers

3. **Rate Limiting:**
   - The microservice does not implement rate limiting
   - Rate limiting is handled by the Tradernet API

4. **Stateless Design:**
   - The microservice is stateless - credentials are passed per request
   - No session management or credential storage
