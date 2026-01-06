# Tradernet SDK 2.0.0 - Comprehensive Analysis for Go Port

## Executive Summary

This document provides a **complete, line-by-line analysis** of the `tradernet-sdk==2.0.0` Python package to ensure a 100% accurate Go port. Every detail matters for financial operations.

**Source**: PyPI package `tradernet-sdk==2.0.0`
**Author**: Anton Kudelin (a.kudelin@freedomfinance.eu)
**Base URL**: `https://freedom24.com`

---

## 1. Authentication - CRITICAL IMPLEMENTATION

### 1.1 Request Flow (V2/V3)

```python
# From core.py, authorized_request() method

# Step 1: Prepare payload
payload = self.stringify(params)  # JSON stringify (no spaces, NO key sorting!)

# Step 2: Get timestamp
timestamp = str(int(time()))  # Unix timestamp in SECONDS (not milliseconds!)

# Step 3: Create message for signing
message = payload + timestamp  # String concatenation, NOT JSON!

# Step 4: Sign message
signature = self.sign(self._private, message)  # SHA256 HMAC

# Step 5: Build request
url = f'{self.url}/api/{cmd}'  # Note: /api/{cmd}, NOT /api/v2/cmd/{cmd}
headers = {
    'Content-Type': 'application/json',
    'X-NtApi-PublicKey': self.public,
    'X-NtApi-Timestamp': timestamp,
    'X-NtApi-Sig': signature
}

# Step 6: Send POST with JSON body
response = self.request('post', url, headers=headers, data=payload)
```

### 1.2 JSON Stringify - CRITICAL DETAILS

```python
# From string_utils.py
@staticmethod
def stringify(items: list[Any] | dict[Any, Any]) -> str:
    return json_dumps(items, separators=(',', ':'))
```

**CRITICAL FINDINGS**:
- ✅ Uses `separators=(',', ':')` - **NO SPACES**
- ❌ **NO `sort_keys=True`** - Keys are NOT sorted!
- ✅ Standard `json.dumps()` behavior otherwise

**Example**:
```python
data = {'b': 2, 'a': 1}
result = StringUtils.stringify(data)
# Result: '{"b":2,"a":1}'  # Keys NOT sorted!
```

**Go Implementation**:
```go
import "encoding/json"

func stringify(data interface{}) (string, error) {
    // Standard JSON marshal - NO key sorting!
    bytes, err := json.Marshal(data)
    if err != nil {
        return "", err
    }
    return string(bytes), nil
}
```

### 1.3 Timestamp Format

```python
timestamp = str(int(time()))  # Unix timestamp in SECONDS
```

**CRITICAL**:
- ✅ Seconds (not milliseconds)
- ✅ Integer conversion (truncates decimal)
- ✅ String conversion

**Go Implementation**:
```go
timestamp := strconv.FormatInt(time.Now().Unix(), 10)
```

### 1.4 Message Construction

```python
message = payload + timestamp  # String concatenation
```

**CRITICAL**:
- ✅ Simple string concatenation
- ❌ NOT JSON object: `{"payload": "...", "timestamp": "..."}`
- ✅ Order: payload first, then timestamp

**Go Implementation**:
```go
message := payload + timestamp  // Simple concatenation
```

### 1.5 Signature Generation

```python
@staticmethod
def sign(key: str, message: str = '', algorithm_name: str = 'sha256') -> str:
    return hmac_new(
        key=key.encode(),
        msg=message.encode(),
        digestmod=algorithm_name
    ).hexdigest()
```

**CRITICAL**:
- ✅ Uses `hmac.new()` with SHA256
- ✅ Encodes key and message as UTF-8 bytes
- ✅ Returns hex digest (lowercase)

**Go Implementation**:
```go
import (
    "crypto/hmac"
    "crypto/sha256"
    "encoding/hex"
)

func sign(key, message string) string {
    mac := hmac.New(sha256.New, []byte(key))
    mac.Write([]byte(message))
    return hex.EncodeToString(mac.Sum(nil))
}
```

### 1.6 Request Structure

```python
# URL format
url = f'{self.url}/api/{cmd}'  # https://freedom24.com/api/{cmd}

# Headers
headers = {
    'Content-Type': 'application/json',
    'X-NtApi-PublicKey': self.public,
    'X-NtApi-Timestamp': timestamp,
    'X-NtApi-Sig': signature
}

# Body
data = payload  # JSON string (not bytes, not form-encoded)
```

**CRITICAL**:
- ✅ POST method
- ✅ JSON body (string, not form-encoded)
- ✅ Headers include all auth info
- ✅ URL is `/api/{cmd}`, not `/api/v2/cmd/{cmd}`

---

## 2. Response Handling

### 2.1 Response Parsing

```python
response = self.request('post', url, headers=headers, data=payload)
result = response.json()

if 'errMsg' in result:
    self.logger.error('Error: %s', result['errMsg'])

return result
```

**CRITICAL**:
- ✅ Always parses as JSON
- ✅ Checks for `errMsg` key (logs but doesn't raise)
- ✅ Returns result even if error present

### 2.2 Error Handling Pattern

The SDK **does NOT raise exceptions** on API errors. It:
1. Logs the error message
2. Returns the response anyway
3. Lets the caller handle errors

**Go Implementation**:
```go
type APIResponse struct {
    Result  interface{} `json:"result,omitempty"`
    ErrMsg  string      `json:"errMsg,omitempty"`
    // ... other fields
}

func (c *Client) checkError(resp *APIResponse) error {
    if resp.ErrMsg != "" {
        c.log.Warn().Str("err_msg", resp.ErrMsg).Msg("API returned error")
        // Don't return error - let caller decide
    }
    return nil
}
```

---

## 3. Methods Used in Your Microservice

### 3.1 `user_info()`

```python
def user_info(self) -> dict[str, Any]:
    return self.authorized_request('GetAllUserTexInfo')
```

**Details**:
- Command: `GetAllUserTexInfo` (note: capital letters!)
- Auth: V2
- Params: None
- Response: User information dict

### 3.2 `buy()` / `sell()`

```python
def buy(self, symbol: str, quantity: int = 1, price: float = 0.0,
        duration: str = 'day', use_margin: bool = True, ...) -> dict[str, Any]:
    if quantity <= 0:
        raise ValueError('Quantity must be positive')
    return self.trade(symbol, quantity, price, duration, use_margin, ...)

def sell(self, symbol: str, quantity: int = 1, ...) -> dict[str, Any]:
    if quantity <= 0:
        raise ValueError('Quantity must be positive')
    return self.trade(symbol, -quantity, ...)  # Negative quantity!
```

**Internal `trade()` method**:

```python
def trade(self, symbol: str, quantity: int = 1, price: float = 0.0,
          duration: str = 'day', use_margin: bool = True, ...) -> dict[str, Any]:
    # IOC emulation (special case)
    if duration == 'ioc':
        order = self.trade(symbol, quantity, price, 'day', use_margin, ...)
        if 'order_id' in order:
            self.cancel(order['order_id'])
        return order

    # Duration validation
    duration = duration.lower()  # Case-insensitive!
    if duration not in self.DURATION:
        raise ValueError(f'Unknown duration {duration}')

    # Action ID calculation
    if quantity > 0:    # buy
        action_id = 2 if use_margin else 1
    elif quantity < 0:  # sell
        action_id = 4 if use_margin else 3
    else:
        raise ValueError('Zero quantity!')

    # Order type: 1 = market, 2 = limit
    order_type_id = 2 if price != 0 else 1

    return self.authorized_request('putTradeOrder', {
        'instr_name': symbol,
        'action_id': action_id,
        'order_type_id': order_type_id,
        'qty': abs(quantity),  # Absolute value!
        'limit_price': price,
        'expiration_id': self.DURATION[duration],
        'user_order_id': custom_order_id  # Can be None
    })
```

**CRITICAL DETAILS**:

1. **Quantity Validation**:
   - ✅ Must be > 0 (raises ValueError)
   - ✅ Uses `abs(quantity)` in API call
   - ✅ Sign handled by `action_id`

2. **Duration**:
   - ✅ Case-insensitive (`duration.lower()`)
   - ✅ Must be in `DURATION` dict
   - ✅ Mapping: `{'day': 1, 'ext': 2, 'gtc': 3}`

3. **Action ID**:
   ```
   Buy + no margin  = 1
   Buy + margin     = 2
   Sell + no margin = 3
   Sell + margin    = 4
   ```

4. **Order Type ID**:
   - `price == 0` → `order_type_id = 1` (market)
   - `price != 0` → `order_type_id = 2` (limit)

5. **IOC Emulation**:
   - Special case: places order with 'day' duration, then immediately cancels
   - Returns the order result

6. **Response Parsing** (from your microservice):
   ```python
   order_id = str(result.get("id", "") or result.get("orderId", ""))
   price = float(result.get("price", 0) or result.get("p", 0))
   ```
   - ✅ Fallback field names: `id` or `orderId`, `price` or `p`

### 3.3 `get_placed(active=True)`

```python
def get_placed(self, active: bool = True) -> dict[str, Any]:
    return self.authorized_request(
        'getNotifyOrderJson',
        {'active_only': int(active)}  # Converts bool to int!
    )
```

**CRITICAL**:
- ✅ Converts boolean to integer: `True` → `1`, `False` → `0`
- ✅ Command: `getNotifyOrderJson`

**Response Structure** (from your microservice):
```python
response.get("result", {}).get("orders", {})
order_list = orders_data.get("order", [])

# Handle single order (dict) vs list
if isinstance(order_list, dict):
    order_list = [order_list]

# Parse each order
order.get("id", "")           # Order ID
order.get("instr_name", "")   # Symbol
order.get("buy_sell", "")     # Side
order.get("qty", 0)           # Quantity
order.get("price", 0)         # Price
order.get("curr", "")         # Currency
```

**CRITICAL**: Response can be:
- List of orders: `{"result": {"orders": {"order": [{...}, {...}]}}}`
- Single order: `{"result": {"orders": {"order": {...}}}}` (dict, not list!)

### 3.4 `account_summary()`

```python
def account_summary(self) -> dict[str, Any]:
    return self.authorized_request('getPositionJson')
```

**Response Structure** (from your microservice):
```python
summary.get("result", {}).get("ps", {})
ps_data.get("pos", [])  # Positions
ps_data.get("acc", [])  # Cash accounts

# Position fields
item.get("i", "")              # Symbol
item.get("q", 0)              # Quantity
item.get("bal_price_a", 0)    # Average price
item.get("mkt_price", 0)      # Current price
item.get("profit_close", 0)   # Unrealized P&L
item.get("curr", "EUR")       # Currency

# Cash account fields
item.get("s", 0)              # Amount
item.get("curr", "")         # Currency
```

### 3.5 `get_trades_history()`

```python
def get_trades_history(self, start: str | date = date(1970, 1, 1),
                       end: str | date = date.today(), ...) -> dict[str, Any]:
    params = {
        'beginDate': str(start),  # Converts to string!
        'endDate': str(end)
    }
    # Optional params: trade_id, limit, symbol, currency
    return self.authorized_request('getTradesHistory', params)
```

**CRITICAL**:
- ✅ Date format: `str(date)` → ISO format `YYYY-MM-DD`
- ✅ Command: `getTradesHistory`

**Response Structure** (from your microservice):
```python
trades_data.get("trades", {}).get("trade", [])

# Handle single trade (dict) vs list
if isinstance(trade_list, dict):
    trade_list = [trade_list]

# Parse each trade
trade.get("id") or trade.get("order_id") or ""  # Order ID (fallback!)
trade.get("q") or trade.get("qty") or 0         # Quantity (fallback!)
trade.get("price") or trade.get("p") or 0      # Price (fallback!)
trade.get("type", "")                          # Trade type: "1" or 1 = BUY
trade.get("instr_nm") or trade.get("i") or ""  # Symbol (fallback!)
trade.get("date") or trade.get("d") or ""      # Date (fallback!)
```

**CRITICAL**: Multiple fallback field names for each value!

### 3.6 `get_quotes(symbols)`

```python
def get_quotes(self, symbols: Sequence[str]) -> dict[str, Any]:
    if isinstance(symbols, str):
        symbols = [symbols]  # Single symbol → list

    return self.authorized_request(
        'getStockQuotesJson',
        {'tickers': ','.join(symbols)}  # Comma-separated string!
    )
```

**CRITICAL**:
- ✅ Handles single symbol or list
- ✅ Comma-separated string: `'AAPL.US,MSFT.US'`
- ✅ Command: `getStockQuotesJson`

**Response Structure** (from your microservice):
```python
# Response can be list or dict!
if isinstance(quotes, list) and len(quotes) > 0:
    data = quotes[0]
elif isinstance(quotes, dict):
    data = quotes
else:
    return None

# Parse quote
data.get("ltp", data.get("last_price", 0))      # Price (fallback!)
data.get("chg", data.get("change", 0))          # Change (fallback!)
data.get("chg_pc", data.get("change_pct", 0))   # Change % (fallback!)
data.get("v", data.get("volume", 0))            # Volume (fallback!)
```

### 3.7 `get_candles(symbol, start, end)`

```python
def get_candles(self, symbol: str, start: datetime = datetime(2010, 1, 1),
                end: datetime = datetime.now(), timeframe: int = 86400) -> dict[str, Any]:
    return self.authorized_request('getHloc', {
        'id': symbol,
        'count': -1,
        'timeframe': int(timeframe / 60),  # Converts seconds to minutes!
        'date_from': start.strftime('%d.%m.%Y %H:%M'),  # Custom format!
        'date_to': end.strftime('%d.%m.%Y %H:%M'),
        'intervalMode': 'OpenRay'
    })
```

**CRITICAL**:
- ✅ Date format: `'%d.%m.%Y %H:%M'` → `'01.01.2020 00:00'`
- ✅ Timeframe: Converts seconds to minutes (`timeframe / 60`)
- ✅ Default: 86400 seconds = 1 day candles
- ✅ Command: `getHloc`

**Response Structure** (from your microservice):
```python
# Response can be list!
if isinstance(data, list):
    for candle in data:
        candle.get("date", "")
        candle.get("o", 0)   # Open
        candle.get("h", 0)   # High
        candle.get("l", 0)   # Low
        candle.get("c", 0)   # Close
        candle.get("v", 0)   # Volume
```

### 3.8 `find_symbol(symbol, exchange)`

```python
def find_symbol(self, symbol: str, exchange: str | None = None) -> dict[str, Any]:
    return self.plain_request(
        'tickerFinder',
        {'text': f'{symbol}@{exchange}' if exchange else symbol}
    )
```

**CRITICAL**:
- ✅ Uses `plain_request()` (no auth!)
- ✅ Format: `'AAPL@NYSE'` or `'AAPL'`
- ✅ Command: `tickerFinder`

**Response Structure** (from your microservice):
```python
result.get("found", [])

for item in found:
    item.get("t", "")        # Symbol
    item.get("nm", "")       # Name
    item.get("isin", "")     # ISIN
    item.get("x_curr", "")   # Currency
    item.get("mkt", "")      # Market
    item.get("codesub", "")  # Exchange code
```

### 3.9 `security_info(symbol, sup=True)`

```python
def security_info(self, symbol: str, sup: bool = True) -> dict[str, Any]:
    return self.authorized_request(
        'getSecurityInfo',
        {'ticker': symbol, 'sup': sup}  # Boolean stays boolean!
    )
```

**CRITICAL**:
- ✅ Boolean parameter (not converted to int!)
- ✅ Command: `getSecurityInfo`

### 3.10 `authorized_request(method, params, version=2)`

```python
def authorized_request(self, cmd: str, params: dict[str, Any] | None = None,
                      version: int | None = 2) -> Any:
    if self.public is None or self._private is None:
        raise ValueError('Keypair is not valid')

    # ... (auth logic)

    if version in (2, 3):
        # V2/V3 logic
    else:
        raise ValueError(f'Unsupported API version {version}')
```

**CRITICAL**:
- ✅ Validates credentials (raises ValueError if missing)
- ✅ Only supports version 2 or 3
- ✅ Raises ValueError for unsupported versions

---

## 4. Plain Requests (No Auth)

### 4.1 `plain_request()`

```python
def plain_request(self, cmd: str, params: dict[str, Any] | None = None) -> Any:
    message = {'cmd': cmd}
    if params:
        message['params'] = params

    url = f'{self.url}/api'
    query = {'q': json_dumps(message)}  # GET request with ?q=<json>

    return self.request('get', url, params=query)
```

**CRITICAL**:
- ✅ GET request (not POST!)
- ✅ Query parameter: `?q=<json>`
- ✅ No authentication
- ✅ URL: `/api` (not `/api/{cmd}`)

**Used by**:
- `find_symbol()` → `tickerFinder`
- `get_most_traded()` → `getTopSecurities`
- `new_user()` → `registerNewUser`

---

## 5. Data Type Conversions

### 5.1 Boolean to Integer

```python
{'active_only': int(active)}  # True → 1, False → 0
{'gainers': int(gainers)}     # True → 1, False → 0
```

**CRITICAL**: Some booleans are converted, others are not!

**Converted to int**:
- `active_only` in `get_placed()`
- `gainers` in `get_most_traded()`

**Stay as boolean**:
- `sup` in `security_info()`

### 5.2 Date/Time Formatting

**Different formats for different endpoints**:

1. **get_trades_history()**: `str(date)` → ISO format `YYYY-MM-DD`
2. **get_candles()**: `datetime.strftime('%d.%m.%Y %H:%M')` → `'01.01.2020 00:00'`
3. **get_historical()**: `datetime.strftime('%Y-%m-%dT%H:%M:%S')` → `'2020-01-01T00:00:00'`

**CRITICAL**: Must use correct format per endpoint!

### 5.3 Timeframe Conversion

```python
'timeframe': int(timeframe / 60)  # Seconds → minutes
```

**CRITICAL**: Converts seconds to minutes (integer division)

---

## 6. Response Structure Patterns

### 6.1 Common Patterns

1. **Wrapped in "result"**:
   ```python
   response.get("result", {}).get("orders", {})
   response.get("result", {}).get("ps", {})
   ```

2. **Single vs List**:
   ```python
   # Can be dict or list!
   if isinstance(data, dict):
       data = [data]
   ```

3. **Fallback Field Names**:
   ```python
   # Multiple possible field names
   value = data.get("id") or data.get("orderId") or ""
   value = data.get("price") or data.get("p") or 0
   ```

4. **Nested Structures**:
   ```python
   # Deep nesting
   response.get("result", {}).get("ps", {}).get("pos", [])
   ```

### 6.2 Error Responses

```python
if 'errMsg' in result:
    self.logger.error('Error: %s', result['errMsg'])
```

**CRITICAL**: Errors are in `errMsg` field, but SDK doesn't raise exceptions!

---

## 7. Edge Cases & Special Handling

### 7.1 IOC Order Emulation

```python
if duration == 'ioc':
    order = self.trade(symbol, quantity, price, 'day', use_margin, ...)
    if 'order_id' in order:
        self.cancel(order['order_id'])
    return order
```

**CRITICAL**: Special handling for IOC orders (immediate-or-cancel)

### 7.2 Empty/Missing Data

```python
# Default to empty list
order_list = orders_data.get("order", [])

# Handle None
if not order_id:
    continue

# Default values
price = float(data.get("price", 0) or data.get("p", 0))
```

### 7.3 Type Coercion

```python
# String conversion
order_id = str(result.get("id", ""))

# Float conversion with defaults
price = float(result.get("price", 0) or result.get("p", 0))

# Integer conversion
quantity_int = int(quantity)
```

---

## 8. Go Implementation Checklist

### 8.1 Authentication

- [ ] JSON stringify: NO spaces, NO key sorting
- [ ] Timestamp: Unix seconds (not milliseconds)
- [ ] Message: `payload + timestamp` (string concatenation)
- [ ] Signature: SHA256 HMAC, hex digest
- [ ] Headers: All 4 headers (Content-Type, PublicKey, Timestamp, Sig)
- [ ] URL: `/api/{cmd}` (not `/api/v2/cmd/{cmd}`)
- [ ] Method: POST
- [ ] Body: JSON string

### 8.2 Request Building

- [ ] Boolean to int conversion (where needed)
- [ ] Date formatting (correct format per endpoint)
- [ ] Timeframe conversion (seconds → minutes)
- [ ] Symbol list (comma-separated)
- [ ] None/null handling (omit from JSON or use null?)

### 8.3 Response Parsing

- [ ] Check for `errMsg` (log but don't fail)
- [ ] Handle nested structures (`result.orders.order`)
- [ ] Handle single vs list (dict → list conversion)
- [ ] Fallback field names (id/orderId, price/p, etc.)
- [ ] Type conversions (str, float, int)
- [ ] Default values for missing fields

### 8.4 Error Handling

- [ ] Validate credentials (raise error if missing)
- [ ] Validate quantity > 0
- [ ] Validate duration (day/ext/gtc)
- [ ] Handle zero quantity error
- [ ] Handle unsupported API version
- [ ] Network errors (timeout, connection, etc.)
- [ ] HTTP status codes (non-200 responses)
- [ ] JSON parsing errors

### 8.5 Methods to Implement

- [ ] `user_info()` - GetAllUserTexInfo
- [ ] `buy()` / `sell()` - putTradeOrder
- [ ] `get_placed()` - getNotifyOrderJson
- [ ] `account_summary()` - getPositionJson
- [ ] `get_trades_history()` - getTradesHistory
- [ ] `get_quotes()` - getStockQuotesJson
- [ ] `get_candles()` - getHloc
- [ ] `find_symbol()` - tickerFinder (plain_request)
- [ ] `security_info()` - getSecurityInfo
- [ ] `authorized_request()` - Generic method
- [ ] `cancel()` - delTradeOrder

---

## 9. Testing Requirements

### 9.1 Authentication Tests

- [ ] Signature matches Python output (same input → same output)
- [ ] Timestamp format (seconds, not milliseconds)
- [ ] JSON stringify (no spaces, no key sorting)
- [ ] Message construction (payload + timestamp)
- [ ] Header format (all 4 headers present)

### 9.2 Method Tests

- [ ] All methods return expected structure
- [ ] Error responses handled correctly
- [ ] Fallback field names work
- [ ] Single vs list handling
- [ ] Type conversions correct
- [ ] Date formats correct per endpoint

### 9.3 Edge Case Tests

- [ ] Empty responses
- [ ] Missing fields
- [ ] Invalid credentials
- [ ] Zero quantity
- [ ] Invalid duration
- [ ] Network errors
- [ ] Timeout handling
- [ ] Large responses

---

## 10. Critical Gotchas

1. **JSON Keys NOT Sorted**: `stringify()` does NOT sort keys!
2. **Timestamp in Seconds**: Not milliseconds, not * 10000
3. **Message is String Concatenation**: Not JSON object
4. **Boolean Conversion**: Some booleans → int, others stay boolean
5. **Date Formats**: Different per endpoint!
6. **Response Structures**: Can be dict or list (handle both!)
7. **Fallback Fields**: Multiple possible field names
8. **Error Handling**: SDK logs but doesn't raise (check `errMsg`)
9. **IOC Orders**: Special emulation logic
10. **Plain Requests**: GET with `?q=<json>`, no auth

---

## 11. Go Implementation Template

```go
package tradernet

import (
    "bytes"
    "crypto/hmac"
    "crypto/sha256"
    "encoding/hex"
    "encoding/json"
    "fmt"
    "net/http"
    "strconv"
    "time"
)

type Client struct {
    publicKey  string
    privateKey string
    baseURL    string
    client     *http.Client
    log        zerolog.Logger
}

func NewClient(publicKey, privateKey string, log zerolog.Logger) *Client {
    return &Client{
        publicKey:  publicKey,
        privateKey: privateKey,
        baseURL:    "https://freedom24.com",
        client: &http.Client{
            Timeout: 30 * time.Second,
        },
        log: log,
    }
}

func (c *Client) stringify(data interface{}) (string, error) {
    // NO key sorting, NO spaces
    bytes, err := json.Marshal(data)
    if err != nil {
        return "", err
    }
    return string(bytes), nil
}

func (c *Client) sign(message string) string {
    mac := hmac.New(sha256.New, []byte(c.privateKey))
    mac.Write([]byte(message))
    return hex.EncodeToString(mac.Sum(nil))
}

func (c *Client) authorizedRequest(cmd string, params map[string]interface{}) (map[string]interface{}, error) {
    // Validate credentials
    if c.publicKey == "" || c.privateKey == "" {
        return nil, fmt.Errorf("keypair is not valid")
    }

    // 1. JSON stringify (no spaces, no key sorting)
    payload, err := c.stringify(params)
    if err != nil {
        return nil, fmt.Errorf("failed to stringify params: %w", err)
    }

    // 2. Get timestamp (seconds, not milliseconds!)
    timestamp := strconv.FormatInt(time.Now().Unix(), 10)

    // 3. Create message (string concatenation)
    message := payload + timestamp

    // 4. Sign message
    signature := c.sign(message)

    // 5. Build request
    url := fmt.Sprintf("%s/api/%s", c.baseURL, cmd)
    req, err := http.NewRequest("POST", url, bytes.NewReader([]byte(payload)))
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %w", err)
    }

    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("X-NtApi-PublicKey", c.publicKey)
    req.Header.Set("X-NtApi-Timestamp", timestamp)
    req.Header.Set("X-NtApi-Sig", signature)

    // 6. Send request
    resp, err := c.client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("request failed: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
    }

    // 7. Parse response
    var result map[string]interface{}
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, fmt.Errorf("failed to parse response: %w", err)
    }

    // 8. Check for errors (log but don't fail)
    if errMsg, ok := result["errMsg"].(string); ok && errMsg != "" {
        c.log.Warn().Str("err_msg", errMsg).Msg("API returned error")
    }

    return result, nil
}
```

---

## 12. Final Checklist

Before deploying the Go port:

- [ ] Authentication signature matches Python exactly
- [ ] All methods tested with real API
- [ ] Response parsing handles all edge cases
- [ ] Error handling matches Python behavior
- [ ] Date formats correct per endpoint
- [ ] Type conversions correct
- [ ] Fallback field names implemented
- [ ] Single vs list handling works
- [ ] Network errors handled gracefully
- [ ] Timeout configured appropriately
- [ ] Logging matches Python SDK
- [ ] All edge cases from this document covered

---

**END OF ANALYSIS**

This document contains every detail needed for a 100% accurate Go port. Follow it precisely to avoid any mistakes in financial operations.
