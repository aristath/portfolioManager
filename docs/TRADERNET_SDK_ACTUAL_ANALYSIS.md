# Tradernet SDK 2.0.0 - Actual Source Code Analysis

## Source
**Package**: `tradernet-sdk==2.0.0` (PyPI)
**Author**: Anton Kudelin (a.kudelin@freedomfinance.eu)
**Homepage**: https://freedom24.com/tradernet-api/python-sdk

## Architecture Overview

The SDK uses a **single authentication method** (V2/V3) with SHA256 HMAC signing. Unlike `tradernet-api`, there's no V1/MD5 authentication.

### Key Classes

1. **`Core`** - Base class with authentication and request methods
2. **`Tradernet`** - Main client class (extends `Core`)
3. **`TraderNetAPI`** - Deprecated alias for `Tradernet`
4. **`Trading`** - Deprecated alias for `Tradernet`

## Authentication Implementation

### Request Signing (V2/V3)

```python
# From core.py - authorized_request()
if version in (2, 3):
    payload = self.stringify(params)  # JSON stringify
    timestamp = str(int(time()))       # Unix timestamp (seconds)
    message = payload + timestamp       # Concatenate JSON + timestamp

    headers['X-NtApi-PublicKey'] = self.public
    headers['X-NtApi-Timestamp'] = timestamp
    headers['X-NtApi-Sig'] = self.sign(self._private, message)
```

**Critical Details**:
- **Payload**: JSON stringified params (no spaces: `separators=(',', ':')`)
- **Timestamp**: Unix timestamp in **seconds** (not milliseconds!)
- **Message**: `payload + timestamp` (concatenated, not JSON)
- **Signature**: SHA256 HMAC of `message` (payload + timestamp)
- **Headers**:
  - `Content-Type: application/json`
  - `X-NtApi-PublicKey: <public_key>`
  - `X-NtApi-Timestamp: <timestamp>`
  - `X-NtApi-Sig: <sha256_hex>`
- **Endpoint**: `POST https://freedom24.com/api/{cmd}`
- **Body**: JSON payload (not form-encoded!)

### Signing Function

```python
@staticmethod
def sign(key: str, message: str = '', algorithm_name: str = 'sha256') -> str:
    return hmac_new(
        key=key.encode(),
        msg=message.encode(),
        digestmod=algorithm_name
    ).hexdigest()
```

**Note**: Uses `hmac.new()` with SHA256, returns hex digest.

### Plain Requests (No Auth)

Some endpoints use `plain_request()` which doesn't require authentication:

```python
def plain_request(self, cmd: str, params: dict[str, Any] | None = None) -> Any:
    message = {'cmd': cmd}
    if params:
        message['params'] = params

    url = f'{self.url}/api'
    query = {'q': json_dumps(message)}  # GET request with ?q=<json>

    return self.request('get', url, params=query)
```

**Used for**:
- `tickerFinder` (find_symbol)
- `getTopSecurities` (get_most_traded)
- `registerNewUser` (new_user)

## Methods Used in Your Microservice

Based on `tradernet_service.py`, here are the methods you're using:

### 1. `user_info()`
```python
def user_info(self) -> dict[str, Any]:
    return self.authorized_request('GetAllUserTexInfo')
```
- **Auth**: V2 (authorized_request)
- **Command**: `GetAllUserTexInfo`
- **Purpose**: Test connection, get user info

### 2. `buy(symbol, quantity)` / `sell(symbol, quantity)`
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
    # Duration validation
    duration = duration.lower()
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
        'qty': abs(quantity),
        'limit_price': price,
        'expiration_id': self.DURATION[duration],
        'user_order_id': custom_order_id
    })
```

**Action ID Mapping**:
- Buy + no margin = 1
- Buy + margin = 2
- Sell + no margin = 3
- Sell + margin = 4

**Duration Mapping**:
```python
DURATION = {
    'day': 1,  # Until end of trading day
    'ext': 2,  # Extended hours
    'gtc': 3   # Good-til-cancelled
}
```

**Edge Cases**:
- ✅ Validates `quantity > 0` (raises ValueError)
- ✅ Validates `duration` (must be 'day', 'ext', or 'gtc')
- ✅ Handles IOC emulation (cancels immediately after placing)
- ✅ Uses `abs(quantity)` for API (sign handled by action_id)

### 3. `get_placed(active=True)`
```python
def get_placed(self, active: bool = True) -> dict[str, Any]:
    return self.authorized_request(
        'getNotifyOrderJson',
        {'active_only': int(active)}  # Converts bool to int!
    )
```
- **Auth**: V2
- **Command**: `getNotifyOrderJson`
- **Params**: `{'active_only': 1}` or `{'active_only': 0}`
- **Note**: Converts boolean to integer!

### 4. `account_summary()`
```python
def account_summary(self) -> dict[str, Any]:
    return self.authorized_request('getPositionJson')
```
- **Auth**: V2
- **Command**: `getPositionJson`
- **Returns**: Portfolio positions, cash balances, etc.

### 5. `get_trades_history()`
```python
def get_trades_history(self, start: str | date = date(1970, 1, 1),
                       end: str | date = date.today(), ...) -> dict[str, Any]:
    params = {
        'beginDate': str(start),  # Converts to string!
        'endDate': str(end)
    }
    # Optional: trade_id, limit, symbol, currency
    return self.authorized_request('getTradesHistory', params)
```
- **Auth**: V2
- **Command**: `getTradesHistory`
- **Date Format**: Converts dates to strings (ISO format: `YYYY-MM-DD`)

### 6. `get_quotes(symbols)`
```python
def get_quotes(self, symbols: Sequence[str]) -> dict[str, Any]:
    if isinstance(symbols, str):
        symbols = [symbols]  # Single symbol -> list

    return self.authorized_request(
        'getStockQuotesJson',
        {'tickers': ','.join(symbols)}  # Comma-separated string!
    )
```
- **Auth**: V2
- **Command**: `getStockQuotesJson`
- **Params**: `{'tickers': 'AAPL.US,MSFT.US'}` (comma-separated!)
- **Note**: Handles both single symbol and list

### 7. `get_candles(symbol, start, end)`
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
- **Auth**: V2
- **Command**: `getHloc`
- **Date Format**: `'%d.%m.%Y %H:%M'` (e.g., `'01.01.2020 00:00'`)
- **Timeframe**: Converts seconds to minutes (`timeframe / 60`)
- **Default**: 86400 seconds = 1 day candles

### 8. `find_symbol(symbol, exchange)`
```python
def find_symbol(self, symbol: str, exchange: str | None = None) -> dict[str, Any]:
    return self.plain_request(
        'tickerFinder',
        {'text': f'{symbol}@{exchange}' if exchange else symbol}
    )
```
- **Auth**: None (plain_request)
- **Command**: `tickerFinder`
- **Format**: `'AAPL@NYSE'` or `'AAPL'` (no `.US` suffix needed!)

### 9. `security_info(symbol, sup=True)`
```python
def security_info(self, symbol: str, sup: bool = True) -> dict[str, Any]:
    return self.authorized_request(
        'getSecurityInfo',
        {'ticker': symbol, 'sup': sup}  # Boolean, not converted!
    )
```
- **Auth**: V2
- **Command**: `getSecurityInfo`
- **Params**: `{'ticker': 'AAPL.US', 'sup': True}` (boolean stays boolean)

### 10. `authorized_request(method, params, version=2)`
```python
def authorized_request(self, cmd: str, params: dict[str, Any] | None = None,
                      version: int | None = 2) -> Any:
    if self.public is None or self._private is None:
        raise ValueError('Keypair is not valid')

    headers = {'Content-Type': 'application/json'}
    params = params or {}

    url = f'{self.url}/api/{cmd}'  # Note: /api/{cmd}, not /api/v2/cmd/{cmd}

    if version in (2, 3):
        payload = self.stringify(params)  # JSON stringify
        timestamp = str(int(time()))
        message = payload + timestamp

        headers['X-NtApi-PublicKey'] = self.public
        headers['X-NtApi-Timestamp'] = timestamp
        headers['X-NtApi-Sig'] = self.sign(self._private, message)
    else:
        raise ValueError(f'Unsupported API version {version}')

    response = self.request('post', url, headers=headers, data=payload)
    result = response.json()

    if 'errMsg' in result:
        self.logger.error('Error: %s', result['errMsg'])

    return result
```

**Critical Details**:
- **URL Format**: `https://freedom24.com/api/{cmd}` (NOT `/api/v2/cmd/{cmd}`)
- **Body**: JSON string (not form-encoded)
- **Error Handling**: Checks for `errMsg` in response

## Key Differences from tradernet-api

| Feature | tradernet-api | tradernet-sdk (actual) |
|---------|---------------|------------------------|
| **Auth Method** | V1 (MD5) + V2 (SHA256) | V2/V3 (SHA256 only) |
| **V1 Endpoint** | `/api` (form data) | N/A |
| **V2 Endpoint** | `/api/v2/cmd/{command}` | `/api/{cmd}` |
| **V2 Format** | URL-encoded form | JSON body |
| **V2 Signature** | Sorted query string | JSON payload + timestamp |
| **Nonce** | `time * 10000` | Unix timestamp (seconds) |
| **Ticker Format** | Auto `.US` suffix | Manual (you provide full symbol) |

## Edge Cases & Validations

### 1. Quantity Validation
```python
if quantity <= 0:
    raise ValueError('Quantity must be positive')
```
- ✅ Validates positive quantity
- ✅ Uses `abs(quantity)` in API call (sign handled by action_id)

### 2. Duration Validation
```python
duration = duration.lower()
if duration not in self.DURATION:
    raise ValueError(f'Unknown duration {duration}')
```
- ✅ Case-insensitive
- ✅ Must be 'day', 'ext', or 'gtc'

### 3. Action ID Calculation
```python
if quantity > 0:    # buy
    action_id = 2 if use_margin else 1
elif quantity < 0:  # sell
    action_id = 4 if use_margin else 3
else:
    raise ValueError('Zero quantity!')
```
- ✅ Handles buy/sell via quantity sign
- ✅ Combines with margin flag
- ❌ Raises error for zero quantity

### 4. Order Type
```python
order_type_id = 2 if price != 0 else 1
```
- ✅ Simple: non-zero price = limit order, zero = market order
- ✅ No complex validation (unlike tradernet-api)

### 5. Boolean to Integer Conversion
```python
{'active_only': int(active)}  # True -> 1, False -> 0
```
- ✅ Converts boolean to int for API

### 6. Date Formatting
- **get_trades_history**: `str(date)` → ISO format (`YYYY-MM-DD`)
- **get_candles**: `datetime.strftime('%d.%m.%Y %H:%M')` → `'01.01.2020 00:00'`
- ✅ Different formats for different endpoints!

### 7. Symbol Handling
- **get_quotes**: Comma-separated string: `'AAPL.US,MSFT.US'`
- **find_symbol**: No `.US` suffix needed, can use `@exchange`
- ✅ Handles single symbol or list

### 8. Error Handling
```python
if 'errMsg' in result:
    self.logger.error('Error: %s', result['errMsg'])
```
- ✅ Checks for `errMsg` in response
- ✅ Logs but doesn't raise (returns result anyway)

## Go Port Implementation Guide

### 1. Authentication Structure

```go
type Client struct {
    publicKey  string
    privateKey string
    baseURL    string
    client     *http.Client
    log        zerolog.Logger
}

func (c *Client) sign(message string) string {
    mac := hmac.New(sha256.New, []byte(c.privateKey))
    mac.Write([]byte(message))
    return hex.EncodeToString(mac.Sum(nil))
}

func (c *Client) authorizedRequest(cmd string, params map[string]interface{}) ([]byte, error) {
    // 1. JSON stringify params (no spaces)
    payload, err := json.Marshal(params)
    if err != nil {
        return nil, err
    }
    payloadStr := string(payload)

    // 2. Get timestamp (seconds, not milliseconds!)
    timestamp := strconv.FormatInt(time.Now().Unix(), 10)

    // 3. Create message: payload + timestamp
    message := payloadStr + timestamp

    // 4. Sign message
    signature := c.sign(message)

    // 5. Build request
    url := fmt.Sprintf("%s/api/%s", c.baseURL, cmd)
    req, err := http.NewRequest("POST", url, bytes.NewReader(payload))
    if err != nil {
        return nil, err
    }

    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("X-NtApi-PublicKey", c.publicKey)
    req.Header.Set("X-NtApi-Timestamp", timestamp)
    req.Header.Set("X-NtApi-Sig", signature)

    // 6. Send request
    resp, err := c.client.Do(req)
    // ... handle response
}
```

### 2. Critical Implementation Points

#### JSON Stringify
```go
// Must match Python's json.dumps(params, separators=(',', ':'))
// No spaces, no trailing commas
payload, err := json.Marshal(params)
```

#### Timestamp
```go
// Unix timestamp in SECONDS (not milliseconds!)
timestamp := strconv.FormatInt(time.Now().Unix(), 10)
```

#### Message Construction
```go
// Concatenate JSON string + timestamp (not JSON object!)
message := string(payload) + timestamp
```

#### Boolean to Integer
```go
activeOnly := 0
if active {
    activeOnly = 1
}
params := map[string]interface{}{
    "active_only": activeOnly,
}
```

#### Date Formatting
```go
// get_trades_history: ISO format
dateStr := date.Format("2006-01-02")

// get_candles: Custom format
dateStr := date.Format("02.01.2006 15:04")
```

#### Symbol Lists
```go
// get_quotes: Comma-separated
tickers := strings.Join(symbols, ",")
params := map[string]interface{}{
    "tickers": tickers,
}
```

### 3. Method Implementations

#### Buy/Sell
```go
func (c *Client) Buy(symbol string, quantity int, price float64,
                     duration string, useMargin bool) (map[string]interface{}, error) {
    if quantity <= 0 {
        return nil, fmt.Errorf("quantity must be positive")
    }
    return c.Trade(symbol, quantity, price, duration, useMargin)
}

func (c *Client) Sell(symbol string, quantity int, price float64,
                      duration string, useMargin bool) (map[string]interface{}, error) {
    if quantity <= 0 {
        return nil, fmt.Errorf("quantity must be positive")
    }
    return c.Trade(symbol, -quantity, price, duration, useMargin)
}

func (c *Client) Trade(symbol string, quantity int, price float64,
                       duration string, useMargin bool) (map[string]interface{}, error) {
    // Validate duration
    durationID, ok := durationMap[duration]
    if !ok {
        return nil, fmt.Errorf("unknown duration: %s", duration)
    }

    // Calculate action_id
    var actionID int
    if quantity > 0 {
        if useMargin {
            actionID = 2
        } else {
            actionID = 1
        }
    } else if quantity < 0 {
        if useMargin {
            actionID = 4
        } else {
            actionID = 3
        }
    } else {
        return nil, fmt.Errorf("zero quantity")
    }

    // Order type
    orderTypeID := 2
    if price == 0 {
        orderTypeID = 1  // Market order
    }

    params := map[string]interface{}{
        "instr_name":    symbol,
        "action_id":     actionID,
        "order_type_id": orderTypeID,
        "qty":           abs(quantity),
        "limit_price":   price,
        "expiration_id": durationID,
    }

    return c.authorizedRequest("putTradeOrder", params)
}
```

## Testing Checklist

- [ ] Authentication signature matches Python output
- [ ] Timestamp format (seconds, not milliseconds)
- [ ] JSON stringify (no spaces)
- [ ] Message construction (payload + timestamp)
- [ ] Action ID calculation (all 4 combinations)
- [ ] Duration validation and mapping
- [ ] Boolean to integer conversion
- [ ] Date formatting (different formats per endpoint)
- [ ] Symbol list handling (comma-separated)
- [ ] Error handling (`errMsg` check)
- [ ] Quantity validation
- [ ] Zero quantity error
- [ ] Invalid duration error

## Summary

The actual `tradernet-sdk` is **simpler** than `tradernet-api`:
- ✅ Single authentication method (V2/V3)
- ✅ JSON body (not form-encoded)
- ✅ Simpler signature (payload + timestamp)
- ✅ No complex validation (unlike tradernet-api's action_id/order_type_id)
- ✅ Direct endpoint (`/api/{cmd}` not `/api/v2/cmd/{cmd}`)

**Key Takeaway**: The authentication is straightforward - just JSON payload + timestamp, signed with SHA256 HMAC.
