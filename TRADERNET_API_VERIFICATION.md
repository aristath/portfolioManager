# Tradernet API Complete Verification Document

**Purpose:** Verify every API method we use against official documentation
**Date:** 2026-01-08
**Status:** IN PROGRESS - Collecting documentation

---

## ENDPOINTS TO VERIFY (Priority Order)

### 🔴 CRITICAL - Trading & Money (Must verify immediately)

1. **Sending of order for execution** - `putTradeOrder`
   - Status: ✅ DOCUMENTED BELOW
   - Our method: `Buy()`, `Sell()`, `Trade()`

2. **Getting information on a portfolio and subscribing for changes** - `getPositionJson`
   - Status: ✅ DOCUMENTED BELOW
   - Our method: `AccountSummary()`

3. **Receiving clients' requests history** - `getClientCpsHistory`
   - Status: ✅ DOCUMENTED BELOW
   - Our method: `GetClientCpsHistory()`

4. **Money funds movement** - `getUserCashFlows`
   - Status: ✅ DOCUMENTED BELOW
   - Our method: ❌ NOT IMPLEMENTED (different from CPS history)

5. **Cancel the order** - `delTradeOrder`
   - Status: ✅ DOCUMENTED BELOW
   - Our method: `Cancel()`

6. **Sending Stop Loss and Take Profit losses** - `putStopLoss`
   - Status: ✅ DOCUMENTED BELOW
   - Our methods: `Stop()`, `TakeProfit()`, `TrailingStop()`

---

### 🟡 HIGH PRIORITY - Order & Trade Management

7. **Receive orders in the current period and subscribe for changes** - `getNotifyOrderJson`
   - Status: ✅ DOCUMENTED BELOW
   - Our method: `GetPlaced()`

8. **Get orders list for the period** - `getOrdersHistory`
   - Status: ✅ DOCUMENTED BELOW
   - Our method: `GetHistorical()`

9. **Retrieving trades history** - `getTradesHistory`
   - Status: ✅ DOCUMENTED BELOW
   - Our method: `GetTradesHistory()`

---

### 🟢 MEDIUM PRIORITY - Market Data & Quotes

10. **Get stock ticker data** - `getStockQuotesJson`
    - Status: ⏳ NEED DOCS
    - Our method: `GetQuotes()`

11. **Get quote historical data (candlesticks)** - `getHloc`
    - Status: ⏳ NEED DOCS
    - Our method: `GetCandles()`

12. **Stock ticker search** - `tickerFinder`
    - Status: ⏳ NEED DOCS
    - Our method: `FindSymbol()`

13. **Get updates on market status** - `getMarketStatus`
    - Status: ⏳ NEED DOCS
    - Our method: `GetMarketStatus()`

---

### 🔵 LOWER PRIORITY - Advanced Features

14. **News on securities** - `getNews`
    - Status: ⏳ NEED DOCS
    - Our method: `GetNews()`

15. **Getting the most traded securities** - `getTopSecurities`
    - Status: ⏳ NEED DOCS
    - Our method: `GetMostTraded()`

16. **Options demonstration** - `getOptionsByMktNameAndBaseAsset`
    - Status: ⏳ NEED DOCS
    - Our method: `GetOptions()`

17. **Get current price alerts** - `getAlertsList`
    - Status: ⏳ NEED DOCS
    - Our method: `GetPriceAlerts()`

18. **Add price alert** - `addPriceAlert`
    - Status: ⏳ NEED DOCS
    - Our method: `AddPriceAlert()`

19. **Delete price alert** - `addPriceAlert` (with del flag)
    - Status: ⏳ NEED DOCS
    - Our method: `DeletePriceAlert()`

20. **Receiving broker report** - `getBrokerReport`
    - Status: ⏳ NEED DOCS
    - Our method: `GetBrokerReport()`

21. **Receiving order files** - `getCpsFiles`
    - Status: ⏳ NEED DOCS
    - Our method: `GetOrderFiles()`

---

### 🟣 OPTIONAL - Reference Data & Utilities

22. **Exchange rate by date** - `getCrossRates` (?)
    - Status: ⏳ NEED DOCS
    - Our method: Not implemented (we use live FX quotes instead)

23. **List of currencies** - `getCurrencyList` (?)
    - Status: ⏳ NEED DOCS
    - Our method: Not implemented

24. **Initial user data** - `getOPQ`
    - Status: ⏳ NEED DOCS
    - Our method: `GetUserData()`

25. **Instruments details** - `getSecurityInfo`
    - Status: ⏳ NEED DOCS
    - Our method: `SecurityInfo()`

26. **Directory of securities** - `getReadyList`
    - Status: ⏳ NEED DOCS
    - Our method: `Symbols()`

---

## AUTHENTICATION (Need to verify but not critical)

27. **API key** - Authentication method
    - Status: ⏳ NEED DOCS
    - Our method: `SetCredentials()`

---

---

# DOCUMENTATION COLLECTED

## 1. ✅ Sending of order for execution - `putTradeOrder`

**Command:** `putTradeOrder`
**Method:** HTTPS POST (API V2)

### Request Parameters:

```json
{
    "instr_name"    : "AAPL.US",     // string - Required
    "action_id"     : 1,              // int - Required
    "order_type_id" : 2,              // int - Required
    "qty"           : 100,            // int - Required
    "limit_price"   : 40,             // null|float - Optional
    "stop_price"    : 0,              // null|float - Optional
    "expiration_id" : 3,              // int - Required
    "user_order_id" : 146615630       // null|int - Optional
}
```

### Parameter Descriptions:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `instr_name` | string | Yes | Instrument ticker (e.g., "AAPL.US") |
| `action_id` | int | Yes | 1=Buy, 2=Buy on Margin, 3=Sell, 4=Sell Short* |
| `order_type_id` | int | Yes | 1=Market, 2=Limit, 3=Stop, 4=Stop Limit |
| `qty` | int | Yes | Quantity in the order |
| `limit_price` | null\|float | Optional | Limit price |
| `stop_price` | null\|float | Optional | Stop price |
| `expiration_id` | int | Yes | 1=Day, 2=Day+Ext, 3=GTC |
| `user_order_id` | null\|int | Optional | Custom order ID |

**Note:** *Values 3 and 4 for action_id are now the same in the system (Tradernet allows margin at all times).

### Response (Success):

```json
{
    "order_id": 4982349829328
}
```

### Response (Error):

```json
// Common error
{
    "errMsg": "Unsupported query method",
    "code": 2
}

// Method error
{
    "error": "Invalid transaction identifier, allowed values 1 - purchase, 3 - sale",
    "code": 0
}
```

### Our Implementation Issues Found:

❌ **BUG #1:** Missing `stop_price` parameter - cannot place stop orders
❌ **BUG #2:** Only supports order_type_id 1 and 2, not 3 (Stop) or 4 (Stop Limit)
⚠️ **ISSUE #3:** `limit_price` should be optional (null), we always send float

---

## 2. ✅ Getting information on a portfolio - `getPositionJson`

**Command:** `getPositionJson`
**Method:** HTTPS POST (API V2)

### Request Parameters:

```json
{
    "cmd": "getPositionJson",
    "params": {}
}
```

No parameters required.

### Response Structure:

```typescript
// Account info (cash balances)
type AccountInfoRow = {
    curr: string,          // Account currency
    currval: number,       // Account currency exchange rate
    forecast_in: number,
    forecast_out: number,
    s: number,            // Available funds
    t2_in: string,
    t2_out: string
}

// Position info
type PositionInfoRow = {
    acc_pos_id: number,      // Unique position ID
    accruedint_a: number,    // Accrued coupon income (ACI)
    curr: string,            // Position currency
    currval: number,         // Currency exchange rate
    fv: number,              // Coefficient for initial margin
    go: number,              // Initial margin per position
    i: string,               // Position ticker
    q: number,               // Number of securities
    vm: number,              // Variable margin
    name: string,            // Issuer name
    name2: string,           // Issuer alternative name
    mkt_price: number,       // Market value
    market_value: number,    // Asset value
    bal_price_a: number,     // Book value
    open_bal: number,        // Position book value
    price_a: number,         // Book value when opened
    profit_close: number,    // Previous day profit
    profit_price: number,    // Current position profit
    close_price: number,     // Closing price
    trade: {trade_count: number}[]
}

// Response
type PortfolioResponse = {
    key: string,
    acc: AccountInfoRow[],
    pos: PositionInfoRow[]
}
```

### Example Response:

```json
{
    "key": "%test@test.com",
    "acc": [
        {
            "s": ".00000000",
            "forecast_in": ".00000000",
            "forecast_out": ".00000000",
            "curr": "USD",
            "currval": 78.95,
            "t2_in": ".00000000",
            "t2_out": ".00000000"
        }
    ],
    "pos": [
        {
            "i": "AAPL.US",
            "q": 100,
            "curr": "USD",
            "currval": 1,
            "name": "Apple Inc.",
            "name2": "Apple Inc.",
            "open_bal": 299.4,
            "mkt_price": 23.81,
            "profit_close": -2.4,
            "profit_price": 2.83,
            "acc_pos_id": 85600002,
            "bal_price_a": 29.924,
            "price_a": 29.924,
            "market_value": 2020,
            "close_price": 20.83
        }
    ]
}
```

### Our Implementation Issues Found:

⚠️ **ISSUE #1:** We use `profit_close` (previous day) - should we use `profit_price` (current)?
⚠️ **ISSUE #2:** We ignore `currval` (exchange rate) - we fetch it separately
📝 **INFO:** Many optional fields available but not extracted (metadata/margin info)

---

## 3. ✅ Receiving clients' requests history - `getClientCpsHistory`

**Command:** `getClientCpsHistory`
**Method:** HTTPS POST (API V2)

### Request Parameters:

```json
{
    "cmd": "getClientCpsHistory",
    "params": {
        "cpsDocId"   : 181,          // null|int - Optional
        "id"         : 123123123,    // null|int - Optional
        "date_from"  : "2020-04-10", // null|string|date - Optional
        "date_to"    : "2020-05-10", // null|string|date - Optional
        "limit"      : 100,          // null|int - Optional
        "offset"     : 20,           // null|int - Optional
        "cps_status" : 1             // null|int - Optional (0-3)
    }
}
```

### Parameter Descriptions:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `cpsDocId` | null\|int | Optional | Specific CPS document ID |
| `id` | null\|int | Optional | Specific request ID |
| `date_from` | null\|string\|date | Optional | Start date (YYYY-MM-DD or ISO: "2011-01-11T00:00:00") |
| `date_to` | null\|string\|date | Optional | End date (YYYY-MM-DD or ISO: "2024-01-01T00:00:00") |
| `limit` | null\|int | Optional | Max records (pagination) |
| `offset` | null\|int | Optional | Skip records (pagination) |
| `cps_status` | null\|int | Optional | 0=Draft, 1=Processing, 2=Rejected, 3=Executed |

### Response Structure:

Array of CPS history records with fields:
- `id`, `transaction_id`, `type_doc_id`
- `type`, `transaction_type` (both present)
- `dt`, `date` (short and full forms)
- `sm`, `amount` (short and full forms)
- `curr`, `currency` (short and full forms)
- `sm_eur`, `amount_eur` (short and full forms)
- `status`, `status_c`, `description`
- `params` (additional data map)

**Note:** API returns both short cryptic names (dt, sm, curr) and full names (date, amount, currency). Applications must handle both.

### Our Implementation Review:

✅ **Parameters:** Our `GetClientCpsHistoryParams` struct matches the API specification
- `DateFrom` and `DateTo` as strings ✅
- `CpsDocID`, `ID`, `Limit`, `Offset` as nullable ints ✅
- `CpsStatus` as nullable int (0-3) ✅

⚠️ **Response Parsing:** We have fallback logic for field name variants
- Uses helper functions: `getDateField()`, `getAmountField()`, `getCurrencyField()`, `getAmountEURField()`
- Priority: clear names (date, amount, currency) over cryptic (dt, sm, curr)
- Preserves `type_doc_id` in params to prevent data loss

⚠️ **Potential Issues:**
1. **Field Coverage:** Cannot confirm we extract ALL fields the API returns
2. **Type Normalization:** `type` vs `transaction_type` - are we handling both correctly?
3. **Zero-value handling:** Previously had bug where zero values were ignored (FIXED)

📝 **Recommendations:**
- Add logging to detect unknown fields in API responses
- Verify we're correctly handling status codes (0-3)
- Test with real data to ensure field extraction priority is correct

---

## 4. ✅ Money funds movement - `getUserCashFlows`

**Command:** `getUserCashFlows`
**Method:** HTTPS GET (different from API V2 POST pattern!)

### Key Difference from `getClientCpsHistory`:

| Feature | `getClientCpsHistory` | `getUserCashFlows` |
|---------|----------------------|-------------------|
| **Purpose** | Client requests/transactions | Actual cash flow movements |
| **Data Type** | Deposits, withdrawals, transfers | Commissions, dividends, trade settlements |
| **Field Names** | dt/sm/curr (cryptic) | date/sum/currency (clear) |
| **Advanced Features** | Basic filtering by status | Full filter/sort/group capabilities |
| **Grouping** | No | By type_code |
| **Daily Totals** | No | Yes (cash_totals) |
| **Limits Info** | No | Yes (per currency) |
| **Refund Control** | No | Yes (without_refund flag) |

### Request Parameters:

```json
{
    "cmd": "getUserCashFlows",
    "SID": "[SID by authorization]",
    "params": {
        "user_id"        : null,    // int|null - Optional
        "groupByType"    : 1,       // int|null - Optional (1=group, 0=no)
        "cash_totals"    : 1,       // int|null - Optional (1=show, 0=hide)
        "hide_limits"    : 0,       // int|null - Optional (1=hide, 0=show)
        "take"           : 10,      // int|null - Optional (limit)
        "skip"           : 5,       // int|null - Optional (offset)
        "without_refund" : 1,       // int|null - Optional (1=exclude, 0=include)
        "filters"        : [...],   // array|null - Advanced filtering
        "sort"           : [...]    // array|null - Sorting
    }
}
```

### Filter Capabilities:

**Fields:** date, sum, currency, comment, type_code

**Operators:**
- Comparison: `eq`, `neq`, `more`, `eqormore`, `eqorless`
- String: `contains`, `doesnotcontain`, `startswith`, `endswith`
- Array: `in`

### Response Structure:

```typescript
type CashFlowResponse = {
    total: number,              // Total record count
    cashflow: CashFlowRecord[], // Cash flow entries
    limits?: {                  // Per-currency limits (optional)
        [currency: string]: {
            minimum: number,
            multiplicity: number,
            maximum: number
        }
    },
    cash_totals?: {             // Daily totals (optional)
        currency: string,
        list: Array<{
            date: string,
            sum: number
        }>
    }
}

type CashFlowRecord = {
    id: string,              // Unique ID
    type_code: string,       // Type code (commission_for_trades, dividend, etc.)
    icon: string,            // UI icon identifier
    date: string,            // Transaction date
    sum: string,             // Amount (string, can be negative)
    comment: string,         // Description with trade reference
    currency: string,        // Currency code
    type_code_name: string   // Human-readable type name
}
```

### Our Implementation Status:

❌ **NOT IMPLEMENTED**

We currently use `GetClientCpsHistory()` for cash flow tracking, which provides:
- Client requests (deposits, withdrawals)
- Transaction status tracking
- Basic date range filtering

**Missing Capabilities:**
- ❌ Detailed cash flow by type (commissions, dividends, settlements)
- ❌ Advanced filtering and sorting
- ❌ Grouping by type_code
- ❌ Daily cash totals
- ❌ Per-currency limits information
- ❌ Refund exclusion

### Should We Implement This?

**Analysis:**

**Pros:**
- More detailed breakdown of cash movements
- Better categorization (type_code)
- Advanced query capabilities (filters, sorting, grouping)
- Daily totals for reporting
- Limits information useful for withdrawal planning

**Cons:**
- Overlaps with `getClientCpsHistory` (different view of same data)
- More complex API (filters, sorting)
- May return more data than needed for basic portfolio management

**Recommendation:**

🟡 **MEDIUM PRIORITY** - Consider implementing if we need:
1. Detailed commission tracking
2. Dividend history separate from positions
3. Advanced cash flow analytics
4. Daily cash total reporting

For now, `GetClientCpsHistory()` provides sufficient data for:
- Tracking deposits/withdrawals
- Monitoring transaction status
- Basic cash flow reconciliation

**If implementing later:**
- Add `GetUserCashFlows()` method to SDK
- Create comprehensive filter/sort builders
- Transform response to domain `BrokerCashFlow` type
- Consider caching due to advanced query capabilities

---

## 5. ✅ Cancel the order - `delTradeOrder`

**Command:** `delTradeOrder`
**Method:** HTTPS POST (API V2)

### Request Parameters:

```json
{
    "order_id": 2929292929  // int - Required
}
```

### Parameter Descriptions:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `order_id` | int | Yes | ID of the order to cancel |

### Response (Success):

```json
{
    "order_id": 2929292929
}
```

### Response (Error):

```json
// Common error (code 2)
{
    "errMsg": "Unsupported query method",
    "code": 2
}

// Method error (code 0)
{
    "error": "Type of security not identified. Please contact Support.",
    "code": 0
}

// Permission error (code 12)
{
    "code": 12,
    "errorMsg": "This order may only be cancelled by a Cancellation Order or through traders",
    "error": "This order may only be cancelled by a Cancellation Order or through traders"
}
```

### Error Codes:

- **Code 2**: Unsupported query method (common error)
- **Code 0**: Type of security not identified (method error)
- **Code 12**: Insufficient rights - order can only be cancelled through special procedures

### Our Implementation Review:

✅ **Implementation:** `internal/clients/tradernet/sdk/methods.go:408`

```go
func (c *Client) Cancel(orderID int) (interface{}, error) {
    params := map[string]interface{}{
        "order_id": orderID,
    }
    return c.authorizedRequest("delTradeOrder", params)
}
```

**Analysis:**
✅ **Parameter:** Correctly passes `order_id` as int
✅ **Command:** Uses correct command "delTradeOrder"
✅ **Authorization:** Uses `authorizedRequest()` (requires auth session)
✅ **Return Type:** Returns interface{} with raw response

**Usage Context:**
Our `Cancel()` method is used in two key scenarios:
1. **Direct order cancellation** - User/system wants to cancel a pending order
2. **IOC emulation** - For "immediate or cancel" orders, we place the order then immediately cancel it:
   ```go
   // Place order with duration="day"
   resp, err := Trade(...)
   orderID := extractOrderID(resp)
   // Immediately cancel to emulate IOC
   Cancel(orderID)
   ```

**Potential Issues:**

⚠️ **ERROR HANDLING:** We don't specifically handle error code 12 (permission denied)
- If an order can only be cancelled through traders, our Cancel will fail
- Should we add retry logic or user notification for code 12?

⚠️ **RESPONSE PARSING:** We return `interface{}` but don't validate that `order_id` is in response
- Should verify the canceled order_id matches what we requested
- Should check for presence of error codes

📝 **Recommendations:**
1. Add specific error handling for code 12 (permission errors)
2. Parse and validate response contains expected `order_id`
3. Add warning log if cancellation fails with code 0 or 12
4. Consider retry logic for common errors (code 2)
5. Document IOC emulation dependency on Cancel() in code comments

**Confidence:** HIGH ✅
- Parameters match exactly
- Command is correct
- Used successfully in production (IOC emulation works)
- Main gap is error handling sophistication

---

## 6. ✅ Sending Stop Loss and Take Profit - `putStopLoss`

**Command:** `putStopLoss`
**Method:** HTTPS POST (API V2)

### Request Parameters:

```json
{
    "instr_name": "SIE.EU",                     // string - Required
    "take_profit": 1,                           // null|float - Optional
    "stop_loss": 1,                             // null|float - Optional
    "stop_loss_percent": 1,                     // null|float - Optional
    "stoploss_trailing_percent": 1              // null|float - Optional
}
```

### Parameter Descriptions:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `instr_name` | string | Yes | Instrument ticker |
| `take_profit` | null\|float | Optional | Take profit price. If null, TP order doesn't change. |
| `stop_loss` | null\|float | Optional | Stop loss price. If null, SL order doesn't change. **Ignored for trailing stops.** |
| `stop_loss_percent` | null\|float | Optional | Stop loss percentage (for trailing stops) |
| `stoploss_trailing_percent` | null\|float | Optional | Trailing stop percentage |

### Important Behavior:

1. **If all stop loss parameters are null**, stop loss order does not change
2. **If take_profit is null**, take profit order does not change
3. **For trailing stops**: use `stop_loss_percent` + `stoploss_trailing_percent` (the `stop_loss` parameter is ignored)

### Response Structure:

```json
{
    "order_id": 192054709,
    "order": {
        "id": 192054709,
        "type": 6,  // 5=StopLoss, 6=TakeProfit
        "instr": "SIE.EU",
        "oper": 3,
        "p": "0.033925",
        "stop": "0.06",
        "stop_init_price": "0.036445",
        "trailing_price": null,
        // ... many more fields
    }
}
```

### Order Type Values:

- **5** = StopLoss order
- **6** = TakeProfit order

### Our Implementation Review:

**Files:**
- `internal/clients/tradernet/sdk/models.go:194` - PutStopLossParams struct
- `internal/clients/tradernet/sdk/methods.go:1132` - Stop() method
- `internal/clients/tradernet/sdk/methods.go:1167` - TrailingStop() method
- `internal/clients/tradernet/sdk/methods.go:1203` - TakeProfit() method

#### 1. Stop() Method

```go
func (c *Client) Stop(symbol string, price float64) (interface{}, error) {
    params := PutStopLossParams{
        InstrName: symbol,
        StopLoss:  &price,
    }
    return c.authorizedRequest("putStopLoss", params)
}
```

**Analysis:**
✅ **Correct:** Sets only `stop_loss` parameter
✅ **Correct:** Leaves other parameters nil (won't change if set)
✅ **Correct:** Uses pointer for nullable float

#### 2. TrailingStop() Method

```go
func (c *Client) TrailingStop(symbol string, percent int) (interface{}, error) {
    if percent == 0 {
        percent = 1 // Default
    }
    params := PutStopLossParams{
        InstrName:               symbol,
        StopLossPercent:         &percent,
        StoplossTrailingPercent: &percent,
    }
    return c.authorizedRequest("putStopLoss", params)
}
```

**Analysis:**
✅ **Correct:** Sets both `stop_loss_percent` and `stoploss_trailing_percent`
✅ **Correct:** Leaves `stop_loss` nil (correctly ignored for trailing stops)
✅ **Correct:** Default value of 1% if 0 provided
⚠️ **TYPE ISSUE:** Uses `*int` but API expects `null|float`

#### 3. TakeProfit() Method

```go
func (c *Client) TakeProfit(symbol string, price float64) (interface{}, error) {
    params := PutStopLossParams{
        InstrName:  symbol,
        TakeProfit: &price,
    }
    return c.authorizedRequest("putStopLoss", params)
}
```

**Analysis:**
✅ **Correct:** Sets only `take_profit` parameter
✅ **Correct:** Leaves stop loss parameters nil (won't change if set)
✅ **Correct:** Uses pointer for nullable float

#### 4. PutStopLossParams Struct

```go
type PutStopLossParams struct {
    InstrName               string   `json:"instr_name"`
    StopLoss                *float64 `json:"stop_loss,omitempty"`
    StopLossPercent         *int     `json:"stop_loss_percent,omitempty"`
    StoplossTrailingPercent *int     `json:"stoploss_trailing_percent,omitempty"`
    TakeProfit              *float64 `json:"take_profit,omitempty"`
}
```

**Analysis:**
✅ **Field names:** Match API exactly
✅ **Required field:** `InstrName` is non-nullable string
✅ **Optional fields:** Use pointers with `omitempty`
⚠️ **TYPE MISMATCH:** `StopLossPercent` and `StoplossTrailingPercent` are `*int`, but API spec says `null|float`

### Issues Found:

#### ⚠️ ISSUE #1: Percent fields use wrong type

**Severity:** LOW-MEDIUM
**Impact:** Percent fields use `*int` instead of `*float64`
- API documentation specifies `null|float` for both percent fields
- Our implementation uses `*int`
- This might work if API accepts integers, but not spec-compliant
- Prevents using fractional percentages (e.g., 0.5% trailing stop)

**Fix Required:**
```go
// Change from:
StopLossPercent         *int `json:"stop_loss_percent,omitempty"`
StoplossTrailingPercent *int `json:"stoploss_trailing_percent,omitempty"`

// To:
StopLossPercent         *float64 `json:"stop_loss_percent,omitempty"`
StoplossTrailingPercent *float64 `json:"stoploss_trailing_percent,omitempty"`
```

**Files Affected:**
- `internal/clients/tradernet/sdk/models.go:197-198`
- `internal/clients/tradernet/sdk/methods.go:1167-1176` (TrailingStop method signature and implementation)

#### ⚠️ ISSUE #2: Response not parsed

**Severity:** LOW
**Impact:** All three methods return `interface{}` without parsing
- Response contains rich order data (order_id, type, stop prices, etc.)
- We don't validate or extract any response fields
- Cannot verify order was created successfully
- Cannot get order_id for tracking

**Recommendations:**
1. Parse response to extract `order_id`
2. Validate `type` field (5=StopLoss, 6=TakeProfit)
3. Verify `stop` and `stop_init_price` match expected values
4. Return structured response instead of `interface{}`

### Summary:

**Overall Confidence:** MEDIUM ✅

**What Works:**
- ✅ Correct command ("putStopLoss")
- ✅ Correct parameter names
- ✅ Correct usage patterns (simple stop, trailing, take profit)
- ✅ Nullable parameters handled correctly
- ✅ Trailing stop ignores `stop_loss` parameter as expected
- ✅ Used successfully in production

**What Needs Fixing:**
- ⚠️ **BUG-005:** Percent fields should be `*float64` not `*int`
- ⚠️ Response parsing could be improved

**Priority:** MEDIUM - Works in practice but not spec-compliant for fractional percentages

---

## 7. ✅ Receive orders - `getNotifyOrderJson`

**Command:** `getNotifyOrderJson`
**Method:** HTTPS POST (API V2)

### Request Parameters:

```json
{
    "cmd": "getNotifyOrderJson",
    "params": {
        "active_only": 1  // int - Optional (1=active, 0=all)
    }
}
```

### Parameter Descriptions:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `active_only` | int | Optional | 1 = Show only active orders, 0 = Show all orders (default) |

### Response Structure:

Returns an **array** of order objects (not a single object):

```typescript
type OrderDataRow = {
    order_id: number,         // Unique order ID
    instr: string,            // Ticker symbol
    type: 1|2|3|4|5|6,       // Order type
    oper: 1|2|3|4,           // Action type
    p: number,                // Order price
    q: number,                // Quantity
    leaves_qty: number,       // Remaining quantity
    stat: number,             // Order status
    stat_d: string,           // Status modification date
    stop: number,             // Stop price
    stop_activated: 0|1,      // Stop activation flag
    trailing_price: number,   // Trailing percentage
    exp: 1|2|3,              // Expiration type
    date: string,             // Order date
    cur: string,              // Currency
    trade: Array<{...}>       // Executed trades
    // ... many more fields
}

type Response = OrderDataRow[]
```

### Our Implementation Review:

**Files:**
- `internal/clients/tradernet/sdk/models.go:30` - GetNotifyOrderJSONParams struct
- `internal/clients/tradernet/sdk/methods.go:252` - GetPlaced() method

#### GetPlaced() Method

```go
func (c *Client) GetPlaced(active bool) (interface{}, error) {
    // Convert boolean to int: True=1, False=0
    activeOnly := 0
    if active {
        activeOnly = 1
    }
    params := GetNotifyOrderJSONParams{
        ActiveOnly: activeOnly,
    }
    return c.authorizedRequest("getNotifyOrderJson", params)
}
```

**Analysis:**
✅ **Correct:** Uses boolean parameter (Go idiom) and converts to int
✅ **Correct:** Parameter name `active_only` matches API
✅ **Correct:** Conversion logic (true→1, false→0) matches API expectation
✅ **Correct:** Uses correct command "getNotifyOrderJson"

#### GetNotifyOrderJSONParams Struct

```go
type GetNotifyOrderJSONParams struct {
    ActiveOnly int `json:"active_only"` // Boolean converted to int: True=1, False=0
}
```

**Analysis:**
✅ **Correct:** Field name matches API
✅ **Correct:** Uses int type (API expects int, not bool)
✅ **Correct:** Well-documented conversion behavior

### Additional Features:

**WebSocket Support:** The API also supports real-time order updates via WebSocket
- WebSocket URL: `wss://wss.tradernet.com/`
- Subscribe by sending: `JSON.stringify(['orders'])`
- Server sends `'orders'` event with order updates
- **NOT IMPLEMENTED** in our SDK (we only use REST API polling)

### Potential Enhancements:

#### 💡 ENHANCEMENT #1: WebSocket real-time order updates

**Current:** We poll using `GetPlaced()` periodically
**Available:** WebSocket subscription for instant order updates

**Benefits:**
- Instant notification when orders fill/cancel
- No polling overhead
- Lower latency for order status changes

**Complexity:** MEDIUM
- Requires WebSocket client implementation
- Connection management (reconnection logic)
- Event dispatching to appropriate handlers

**Priority:** MEDIUM
- Nice-to-have for real-time trading
- Current polling works adequately for retirement fund management

#### ⚠️ ISSUE #1: Response not parsed

**Severity:** LOW
**Impact:** Returns `interface{}` without structured parsing
- Cannot easily extract order fields
- No validation of response structure
- Difficult to work with in calling code

**Recommendation:**
- Create `Order` struct matching OrderDataRow
- Parse response into `[]Order`
- Return typed result

### Summary:

**Overall Confidence:** HIGH ✅

**What Works:**
- ✅ Correct command and parameter name
- ✅ Correct boolean→int conversion
- ✅ Clean Go API (boolean instead of int)
- ✅ Used successfully in production

**Missing/Enhancements:**
- 💡 WebSocket support not implemented (optional enhancement)
- ⚠️ Response parsing could be improved

**Priority:** No critical issues - works as expected

---

## 8. ✅ Get orders list for period - `getOrdersHistory`

**Command:** `getOrdersHistory`
**Method:** HTTPS GET

### Request Parameters:

```json
{
    "cmd": "getOrdersHistory",
    "SID": "[SID by authorization]",
    "params": {
        "from": "2020-03-23T00:00:00",  // datetime - Required
        "till": "2020-04-03T23:59:59"   // datetime - Required
    }
}
```

### Parameter Descriptions:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `from` | datetime | Yes | Period start date, ISO 8601 format: YYYY-MM-DD\Thh:mm:ss |
| `till` | datetime | Yes | Period end date, ISO 8601 format: YYYY-MM-DD\Thh:mm:ss |

**Date Format:** ISO 8601 - `2006-01-02T15:04:05` in Go time format

### Response Structure:

```json
{
    "orders": {
        "key": "USER",
        "order": [
            {
                "id": 1111222222112,
                "date": "2020-03-23T10:00:28.853",
                "stat": 31,
                "instr": "AAPL.US",
                "oper": 1,
                "type": 2,
                "p": 100,
                "q": 10,
                "leaves_qty": 10,
                "trade": [...],
                // ... many more fields
            }
        ]
    }
}
```

### Our Implementation Review:

**Files:**
- `internal/clients/tradernet/sdk/models.go:204` - GetOrdersHistoryParams struct
- `internal/clients/tradernet/sdk/methods.go:1337` - GetHistorical() method

#### GetHistorical() Method

```go
func (c *Client) GetHistorical(start, end time.Time) (interface{}, error) {
    // Date format: "2011-01-11T00:00:00"
    dateFrom := start.Format("2006-01-02T15:04:05")
    dateTo := end.Format("2006-01-02T15:04:05")

    params := GetOrdersHistoryParams{
        From: dateFrom,
        Till: dateTo,
    }
    return c.authorizedRequest("getOrdersHistory", params)
}
```

**Analysis:**
✅ **Correct:** Uses Go time.Time types (Go idiom)
✅ **Correct:** Formats dates to ISO 8601 format "2006-01-02T15:04:05"
✅ **Correct:** Parameter names `from` and `till` match API
✅ **Correct:** Uses correct command "getOrdersHistory"

#### GetOrdersHistoryParams Struct

```go
type GetOrdersHistoryParams struct {
    From string `json:"from"` // Format: "2011-01-11T00:00:00"
    Till string `json:"till"` // Format: "2024-01-01T00:00:00"
}
```

**Analysis:**
✅ **Correct:** Field names match API exactly
✅ **Correct:** Uses string type for formatted datetime
✅ **Correct:** Well-documented expected format

### Response Structure Notes:

The response has a **nested structure**:
- Top level: `orders` object
- Contains: `key` (user login) and `order` array
- The `order` array can be **null** if no orders found
- Each order has comprehensive fields including `trade` array for executions

### Summary:

**Overall Confidence:** HIGH ✅

**What Works:**
- ✅ Correct command and parameter names
- ✅ Correct date format (ISO 8601)
- ✅ Clean Go API (time.Time instead of strings)
- ✅ Proper date formatting

**Missing/Enhancements:**
- ⚠️ Response not parsed (returns `interface{}`)
- ⚠️ No handling of nested structure (orders.order)
- ⚠️ No handling of null order array

**Priority:** No critical issues - works as expected

**Recommendations:**
1. Parse response to extract `orders.order` array
2. Handle null order array gracefully
3. Create Order struct matching HistoricalOrderDataRow
4. Return typed `[]Order` instead of `interface{}`

---

## 9. ✅ Retrieving trades history - `getTradesHistory`

**Command:** `getTradesHistory`
**Method:** HTTPS GET

### Request Parameters:

```json
{
    "cmd": "getTradesHistory",
    "SID": "[SID by authorization]",
    "params": {
        "beginDate": "2020-03-23",      // date|string - Required (YYYY-MM-DD)
        "endDate": "2020-04-08",        // date|string - Required (YYYY-MM-DD)
        "tradeId": 232327727,           // int|null - Optional
        "max": 100,                     // int|null - Optional
        "nt_ticker": "AAPL.US",         // string|null - Optional
        "curr": "USD",                  // string|null - Optional
        "reception": 1                  // int|null - Optional
    }
}
```

### Parameter Descriptions:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `beginDate` | date\|string | Yes | Period start date (YYYY-MM-DD) |
| `endDate` | date\|string | Yes | Period end date (YYYY-MM-DD) |
| `tradeId` | int\|null | Optional | From which Trade ID to start retrieving |
| `max` | int\|null | Optional | Number of trades (0 or null = all) |
| `nt_ticker` | string\|null | Optional | Instrument ticker filter |
| `curr` | string\|null | Optional | Currency filter |
| `reception` | int\|null | Optional | Office ID filter |

**Date Format:** ISO 8601 date - `YYYY-MM-DD` (e.g., "2020-03-23")

### Response Structure:

```json
{
    "trades": {
        "max_trade_id": [{"@text": "40975888"}],
        "trade": [
            {
                "id": 2229992229292,
                "order_id": 299998887727,
                "instr_nm": "AAPL.US",
                "type": 1,  // 1=Buy, 2=Sell
                "p": 141.4,
                "q": 20,
                "v": 2828,
                "date": "2019-08-15T10:10:22",
                "curr_c": "USD",
                // ... many more fields
            }
        ]
    }
}
```

### Our Implementation Review:

**Files:**
- `internal/clients/tradernet/sdk/models.go:37` - GetTradesHistoryParams struct
- `internal/clients/tradernet/sdk/methods.go:266` - GetTradesHistory() method

#### GetTradesHistory() Method

```go
func (c *Client) GetTradesHistory(start, end string, tradeID, limit *int, symbol, currency *string) (interface{}, error) {
    params := GetTradesHistoryParams{
        BeginDate: start,
        EndDate:   end,
        TradeID:   tradeID,
        Max:       limit,
        NtTicker:  symbol,
        Curr:      currency,
    }
    return c.authorizedRequest("getTradesHistory", params)
}
```

**Analysis:**
✅ **Correct:** Parameter names match API (beginDate, endDate, tradeId, max, nt_ticker, curr)
✅ **Correct:** Uses string for dates (caller provides formatted date)
✅ **Correct:** Uses pointers for optional parameters
✅ **Correct:** Uses correct command "getTradesHistory"
❌ **MISSING:** `reception` parameter (Office ID filter)

#### GetTradesHistoryParams Struct

```go
type GetTradesHistoryParams struct {
    BeginDate string  `json:"beginDate"`           // Field 1 - ISO format YYYY-MM-DD
    EndDate   string  `json:"endDate"`             // Field 2 - ISO format YYYY-MM-DD
    TradeID   *int    `json:"tradeId,omitempty"`   // Field 3 - optional
    Max       *int    `json:"max,omitempty"`       // Field 4 - optional
    NtTicker  *string `json:"nt_ticker,omitempty"` // Field 5 - optional
    Curr      *string `json:"curr,omitempty"`      // Field 6 - optional
}
```

**Analysis:**
✅ **Correct:** Field names match API exactly
✅ **Correct:** Required fields (BeginDate, EndDate) are non-nullable strings
✅ **Correct:** Optional fields use pointers with `omitempty`
✅ **Correct:** Well-documented expected format
❌ **MISSING:** `Reception *int` field for office ID filter

### Issues Found:

#### ⚠️ ISSUE #1: Missing `reception` parameter

**Severity:** LOW
**Impact:** Cannot filter trades by office ID
- API supports `reception` (int|null) parameter for filtering by office
- Our implementation doesn't include this parameter
- Prevents filtering trades by specific office/branch

**Use Case:**
- If broker has multiple offices/branches
- Need to filter trades for specific office
- Currently: no way to specify office filter

**Fix Required:**
```go
// Add to GetTradesHistoryParams:
Reception *int `json:"reception,omitempty"` // Field 7 - optional

// Add to GetTradesHistory() signature:
func (c *Client) GetTradesHistory(start, end string, tradeID, limit *int, symbol, currency *string, reception *int) (interface{}, error) {
    params := GetTradesHistoryParams{
        BeginDate: start,
        EndDate:   end,
        TradeID:   tradeID,
        Max:       limit,
        NtTicker:  symbol,
        Curr:      currency,
        Reception: reception,  // Add this
    }
    return c.authorizedRequest("getTradesHistory", params)
}
```

**Files Affected:**
- `internal/clients/tradernet/sdk/models.go:37-44` (GetTradesHistoryParams struct)
- `internal/clients/tradernet/sdk/methods.go:266-276` (GetTradesHistory method)

**Priority:** LOW
- Most users have single office
- Filter typically not needed
- Can be added if multi-office support needed

### Response Structure Notes:

The response has a **nested structure**:
- Top level: `trades` object
- Contains: `max_trade_id` array (with last trade ID) and `trade` array
- Trade fields use cryptic names: `instr_nm`, `curr_c`, `type` (1=Buy, 2=Sell)
- Many numeric fields returned as strings (need parsing)

### Summary:

**Overall Confidence:** HIGH ✅

**What Works:**
- ✅ Correct command and all main parameter names
- ✅ Correct date format expectation (YYYY-MM-DD)
- ✅ Proper optional parameter handling (pointers)
- ✅ Comprehensive filtering (ticker, currency, limit, tradeID)
- ✅ Used successfully in production

**Missing/Enhancements:**
- ⚠️ **BUG-006:** Missing `reception` parameter (LOW severity)
- ⚠️ Response not parsed (returns `interface{}`)
- ⚠️ No handling of nested structure (trades.trade)

**Priority:** No critical issues - works as expected

**Recommendations:**
1. Add `reception` parameter for completeness
2. Parse response to extract `trades.trade` array
3. Handle `max_trade_id` for pagination
4. Create Trade struct matching TradeRow
5. Return typed `[]Trade` instead of `interface{}`

---

## 10. ✅ Get stock ticker data - `getStockQuotesJson`

**Command:** `getStockQuotesJson`
**Method:** HTTPS POST (API V2)
**Alternative:** `/securities/export` REST endpoint (may be recommended - needs testing)

### Request Parameters:

```json
{
    "cmd": "getStockQuotesJson",
    "params": {
        "tickers": "AAPL.US"  // string OR array
    }
}
```

### Parameter Descriptions:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `tickers` | string OR array | Yes | Single ticker: `"AAPL.US"` OR Multiple tickers: `["AAPL.US", "T.US"]` |

**PHP SDK Examples:**
```php
// Single ticker (string)
$response = $client->sendRequest('getStockQuotesJson', ['tickers' => "AAPL.US"]);

// Multiple tickers (array)
$response = $client->sendRequest('getStockQuotesJson', ['tickers' => ["AAPL.US", "T.US"]]);
```

**Alternative REST Endpoint:**
```
/securities/export?params=ltp+ltt+bbp+chg&tickers=USD/KZT+EUR/KZT
```
Note: Uses `+` to separate tickers and parameters

### Response Structure:

```php
[
    'result' => [
        'q' => [
            '0' => [
                'c' => 'AAPL.US',              // Ticker
                'ltr' => '',                    // Exchange of latest trade
                'name' => 'Apple Inc.',         // Name of security
                'name2' => 'Apple Inc.',        // Security name in Latin
                'mrg' => 'M',                   // Margin flag
                'bbp' => '147,39',             // Best bid price
                'bbc' => '',                    // Best bid change ('' = no change, 'D' = down, 'U' = up)
                'bbs' => '89170',              // Best bid size
                'bbf' => '0',                   // Best bid volume
                'bap' => '147,45',             // Best ask/offer price
                'bac' => 'U',                   // Best ask change ('' = no change, 'D' = down, 'U' = up)
                'bas' => '46250',              // Best ask size
                'baf' => '0',                   // Best ask volume
                'pp' => '147,95',              // Previous closing
                'op' => '147,95',              // Opening price
                'ltp' => '147,42',             // Last trade price
                'lts' => '4380',               // Last trade size
                'ltt' => '2016-10-05T17:04:43', // Time of last trade
                'ltc' => 'U',                   // Last trade change ('' = no change, 'D' = down, 'U' = up)
                'chg' => '-0,53',              // Change in points
                'pcp' => '-0,36',              // Percentage change
                'mintp' => '144',              // Minimum trade price per day
                'maxtp' => '151,9',            // Maximum trade price per day
                'vol' => '129617290',          // Trade volume per day (pieces)
                'vlt' => '3816983,97',         // Trading volume per day (currency)
                'yld' => '0',                   // Yield to maturity (bonds)
                'acd' => '0',                   // Accumulated coupon interest (ACI)
                'fv' => '',                     // Face value
                'mtd' => '',                    // Maturity date
                'cpn' => '0',                   // Coupon (currency)
                'cpp' => '0',                   // Coupon period (days)
                'ncd' => '',                    // Next coupon date
                'ncp' => '0',                   // Latest coupon date
                'dpb' => '0',                   // Purchase margin
                'dps' => '0',                   // Short sale margin
                'trades' => '0',                // Number of trades
                'min_step' => '0,01',          // Minimum price increment
                'step_price' => '0,01',        // Price increment
                'strike_price' => '',           // Option strike
                'kind' => '1',                  // Security kind
                'type' => '1'                   // Security type
            ]
        ]
    ]
]
```

### Response Fields Reference:

| Field | Description |
|-------|-------------|
| `c` | Ticker symbol |
| `ltr` | Exchange of the latest trade |
| `name` | Name of security |
| `name2` | Security name in Latin |
| `bbp` | Best bid price |
| `bbc` | Best bid change indicator ('' = no change, 'D' = down, 'U' = up) |
| `bbs` | Best bid size |
| `bbf` | Best bid volume |
| `bap` | Best ask/offer price |
| `bac` | Best ask change indicator |
| `bas` | Value (size) of the best offer |
| `baf` | Volume of the best offer |
| `pp` | Previous closing price |
| `op` | Opening price of the current trading session |
| `ltp` | Last trade price |
| `lts` | Last trade size |
| `ltt` | Time of last trade |
| `chg` | Change in the price of the last trade in points |
| `pcp` | Percentage change relative to previous close |
| `ltc` | Last trade price change indicator |
| `mintp` | Minimum trade price per day |
| `maxtp` | Maximum trade price per day |
| `vol` | Trade volume per day (in pieces) |
| `vlt` | Trading volume per day (in currency) |
| `yld` | Yield to maturity (for bonds) |
| `acd` | Accumulated coupon interest (ACI) |
| `fv` | Face value |
| `mtd` | Maturity date |
| `cpn` | Coupon (in the currency) |
| `cpp` | Coupon period (in days) |
| `ncd` | Next coupon date |
| `ncp` | Latest coupon date |
| `dpb` | Purchase margin |
| `dps` | Short sale margin |
| `trades` | Number of trades |
| `min_step` | Minimum price increment |
| `step_price` | Price increment |
| `strike_price` | Option strike |

### Our Implementation Review:

**Files:**
- `internal/clients/tradernet/sdk/models.go:46-50` - GetStockQuotesJSONParams struct
- `internal/clients/tradernet/sdk/methods.go:280-287` - GetQuotes() method
- `internal/clients/tradernet/transformers.go:463-473` - transformQuote() function

#### GetQuotes() Method

```go
func (c *Client) GetQuotes(symbols []string) (interface{}, error) {
    // Comma-separated string
    tickers := strings.Join(symbols, ",")
    params := GetStockQuotesJSONParams{
        Tickers: tickers,
    }
    return c.authorizedRequest("getStockQuotesJson", params)
}
```

**Analysis:**
✅ **Correct:** Uses correct command "getStockQuotesJson"
✅ **Correct:** Parameter name `tickers` matches API
✅ **Clean API:** Accepts `[]string` (Go idiom) instead of string
⚠️ **FORMAT QUESTION:** Joins with commas - API might expect array or different separator

#### GetStockQuotesJSONParams Struct

```go
type GetStockQuotesJSONParams struct {
    Tickers string `json:"tickers"` // Comma-separated string: "AAPL.US,MSFT.US"
}
```

**Analysis:**
✅ **Field name:** Matches API parameter
⚠️ **Type:** Uses `string` (comma-separated), but PHP SDK shows API accepts both `string` OR `array`
⚠️ **Separator:** Uses comma, but REST endpoint uses `+` sign

#### Response Transformer

```go
func transformQuote(sdkResult interface{}, symbol string) (*Quote, error) {
    // Handles both array and map response formats
    // Extracts: p, change, change_pct, volume, timestamp
}
```

**Analysis:**
✅ **Handles both formats:** Map (keyed by symbol) and array (with symbol field)
⚠️ **Limited fields:** Extracts only basic fields (price, change, volume)
⚠️ **Missing fields:** Doesn't extract bid/ask, OHLC, margin info, bond data

### Issues Found:

#### ⚠️ ISSUE #1: Parameter format uncertainty

**Severity:** MEDIUM
**Impact:** Unclear if comma-separated string works or if array is required

**Evidence:**
- Our SDK: Sends comma-separated string `"AAPL.US,MSFT.US"`
- PHP SDK: Accepts both string `"AAPL.US"` OR array `["AAPL.US", "T.US"]`
- REST endpoint: Uses plus signs `USD/KZT+EUR/KZT`
- Documentation: No clarity on which format is correct for API V2

**Current Status:** Works in production (comma-separated)
**Risk:** May not be using optimal format

**Recommendation:**
- Test if API accepts array format
- Verify comma separator works for multiple tickers
- Consider using `/securities/export` REST endpoint instead

#### ⚠️ ISSUE #2: Limited field extraction

**Severity:** LOW
**Impact:** Extracts only basic quote fields, ignores comprehensive data

**Missing Fields:**
- ❌ Bid/Ask data (bbp, bap, bbs, bas)
- ❌ OHLC data (op, mintp, maxtp, pp)
- ❌ Trade details (lts, ltt, trades)
- ❌ Bond data (yld, acd, cpn, mtd)
- ❌ Margin data (dpb, dps, mrg)
- ❌ Change indicators (bbc, bac, ltc)

**Current Extraction:** Only price (ltp/p), change, changePct, volume, timestamp

**Recommendation:**
- Expand Quote struct to include bid/ask spreads
- Add OHLC fields for charting
- Consider separate BondQuote type for bond-specific fields

#### 📝 ENHANCEMENT: Alternative endpoint

**Alternative:** `/securities/export?params=ltp+ltt+bbp&tickers=AAPL.US+MSFT.US`

**Advantages:**
- REST endpoint (simpler HTTP GET)
- Explicit field selection (params parameter)
- Documented separator (`+`)
- No authentication needed? (needs verification)

**Recommendation:** Test `/securities/export` endpoint as potential replacement

### Summary:

**Overall Confidence:** MEDIUM ✅

**What Works:**
- ✅ Correct command name
- ✅ Used successfully in production
- ✅ Handles both map and array response formats
- ✅ Clean Go API (accepts slice of symbols)

**Questions/Uncertainties:**
- ❓ Is comma separator correct or should we use array/plus sign?
- ❓ Should we use `/securities/export` REST endpoint instead?
- ❓ Are we missing important quote fields?

**Priority:** MEDIUM - Works but may not be optimal

**Recommendations:**
1. **Test parameter formats:** Verify comma-separated string works for multiple tickers
2. **Evaluate REST endpoint:** Test `/securities/export` as potential alternative
3. **Expand field extraction:** Add bid/ask, OHLC, margin fields
4. **Add response validation:** Parse response structure properly
5. **Consider typed response:** Return structured Quote instead of `interface{}`

**Testing Needed:**
- [ ] Verify comma-separated tickers work: `"AAPL.US,MSFT.US"`
- [ ] Test array format: `["AAPL.US", "MSFT.US"]`
- [ ] Test `/securities/export` REST endpoint
- [ ] Verify all response fields are available

---

## 11. ✅ Get quote historical data (candlesticks) - `getHloc`

**Command:** `getHloc`
**Method:** HTTPS GET / HTTPS POST (API V2)

### Request Parameters:

```json
{
    "cmd": "getHloc",
    "params": {
        "userId": null,                      // int|null - Optional (API v1 only)
        "id": "FB.US",                       // string - Required
        "count": -1,                         // signed int - Required
        "timeframe": 1440,                   // int - Required (minutes)
        "date_from": "15.08.2020 00:00",    // string|datetime - Required
        "date_to": "16.08.2020 00:00",      // string|datetime - Required
        "intervalMode": "ClosedRay"          // string - Required (single value)
    }
}
```

### Parameter Descriptions:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `userId` | int\|null | Optional | User ID for registered users (API v1 only) |
| `id` | string | Yes | Ticker name (can specify multiple separated by commas) |
| `count` | signed int | Yes | Number of extra candlesticks beyond interval. -1 = none. Negative values get candlesticks BEFORE interval (e.g., -100 = get 100 candles before start date) |
| `timeframe` | int | Yes | Interval in minutes. Valid values: [1, 5, 15, 60, 1440] |
| `date_from` | string\|datetime | Yes | Start date of interval in format: DD.MM.YYYY hh:mm |
| `date_to` | string\|datetime | Yes | End date of interval in format: DD.MM.YYYY hh:mm |
| `intervalMode` | string | Yes | **Must be "ClosedRay"** (single required value) |

### Response Structure (Success):

```typescript
type HlocResponse = {
    hloc: {
        [ticker: string]: number[][]  // Array of [high, low, open, close]
    },
    vl: {
        [ticker: string]: number[]    // Array of volumes
    },
    xSeries: {
        [ticker: string]: number[]    // Array of timestamps (in SECONDS!!!)
    },
    maxSeries: number,                // Timestamp of most recent candlestick
    info: {
        [ticker: string]: {
            id: string,
            nt_ticker: string,
            short_name: string,
            default_ticker: string,
            code_nm: string,
            currency: string,
            min_step: string,
            lot: string
        }
    },
    took: number                      // Execution time in milliseconds
}
```

### Example Response:

```json
{
    "hloc": {
        "FB.US": [
            [107.25, 106.1603, 106.96, 106.26],
            [106.45, 104.62, 106.45, 104.75],
            [103.33, 100.25, 103.33, 102.32],
            [103.71, 101.41, 102.33, 102.82],
            [103.7301, 100.89, 101.05, 102.97]
        ]
    },
    "vl": {
        "FB.US": [
            7588957,
            8812260,
            5941541,
            6529607,
            5905857
        ]
    },
    "xSeries": {
        "FB.US": [
            1451422800,
            1451509200,
            1451854800,
            1451941200,
            1452027600
        ]
    },
    "maxSeries": 1452027600,
    "info": {
        "FB.US": {
            "id": "FB.US",
            "nt_ticker": "FB.US",
            "short_name": "Facebook, Inc.",
            "default_ticker": "FB",
            "code_nm": "FB",
            "currency": "USD",
            "min_step": "0.01000000",
            "lot": "1.00000000"
        }
    },
    "took": 26.685
}
```

**Response Notes:**
- **HLOC format:** Each candlestick is `[high, low, open, close]`
- **Timestamps:** In SECONDS (not milliseconds!) - xSeries values
- **Multiple tickers:** Can request multiple tickers separated by commas in `id` parameter
- **Alignment:** All arrays (hloc, vl, xSeries) have same length and align by index

### Response (Error):

```json
// Common error
{
    "errMsg": "Bad json",
    "code": 2
}

// Method error
{
    "error": "User is not found",
    "code": 7
}
```

### Our Implementation Review:

**Files:**
- `internal/clients/tradernet/sdk/models.go:52-62` - GetHlocParams struct
- `internal/clients/tradernet/sdk/methods.go:289-308` - GetCandles() method

#### GetCandles() Method

```go
func (c *Client) GetCandles(symbol string, start, end time.Time, timeframeSeconds int) (interface{}, error) {
    // Date format: "01.01.2020 00:00"
    dateFrom := start.Format("02.01.2006 15:04")
    dateTo := end.Format("02.01.2006 15:04")

    // Timeframe: convert seconds to minutes
    timeframeMinutes := timeframeSeconds / 60

    params := GetHlocParams{
        ID:           symbol,
        Count:        -1,
        Timeframe:    timeframeMinutes,
        DateFrom:     dateFrom,
        DateTo:       dateTo,
        IntervalMode: "OpenRay",  // ❌ BUG: Should be "ClosedRay"
    }
    return c.authorizedRequest("getHloc", params)
}
```

**Analysis:**
✅ **Correct:** Uses correct command "getHloc"
✅ **Correct:** Parameter names match API
✅ **Correct:** Clean Go API (accepts time.Time and converts)
✅ **Correct:** Converts timeframe from seconds to minutes
✅ **Correct:** Date format "02.01.2006 15:04" → "DD.MM.YYYY hh:mm"
✅ **Correct:** Count set to -1 (no extra candles)
❌ **BUG:** IntervalMode uses "OpenRay" but API requires "ClosedRay"
⚠️ **Missing:** userId parameter (optional, API v1 only)

#### GetHlocParams Struct

```go
type GetHlocParams struct {
    ID           string `json:"id"`           // Field 1 - Symbol
    Count        int    `json:"count"`        // Field 2 - -1 for all
    Timeframe    int    `json:"timeframe"`    // Field 3 - Minutes (seconds / 60)
    DateFrom     string `json:"date_from"`    // Field 4 - Format: "01.01.2020 00:00"
    DateTo       string `json:"date_to"`      // Field 5 - Format: "01.01.2020 00:00"
    IntervalMode string `json:"intervalMode"` // Field 6 - "OpenRay"  ❌ WRONG
}
```

**Analysis:**
✅ **Correct:** All field names match API
✅ **Correct:** All field types match API
✅ **Correct:** Date format documented correctly
❌ **BUG:** Comment says "OpenRay" but API requires "ClosedRay"
⚠️ **Missing:** userId field (optional)

### Issues Found:

#### 🔴 BUG-007: IntervalMode uses "OpenRay" instead of "ClosedRay"

**Severity:** CRITICAL
**Impact:** API calls may fail or return unexpected data

**Current Implementation:**
```go
IntervalMode: "OpenRay",  // ❌ WRONG
```

**API Requirement:**
```json
"intervalMode": "ClosedRay"  // Required parameter, single value
```

**Evidence:**
- API docs state: "required parameter, a single value ClosedRay"
- Our code uses "OpenRay" (incorrect)
- Comment in struct also says "OpenRay" (propagated error)

**Fix Required:**
```go
// Change in GetCandles() method (line 305):
IntervalMode: "ClosedRay",  // ✅ CORRECT

// Change comment in GetHlocParams struct:
IntervalMode string `json:"intervalMode"` // Field 6 - "ClosedRay"
```

**Files Affected:**
- `internal/clients/tradernet/sdk/models.go:62` (struct comment)
- `internal/clients/tradernet/sdk/methods.go:305` (hardcoded value)

**Priority:** CRITICAL
- May cause API failures
- Data integrity issue (wrong interval mode)
- Simple one-line fix

**Testing Impact:**
- If endpoint currently works, API may accept both values (undocumented)
- Should verify with API team or test both values
- Change may affect existing candlestick data

#### ⚠️ ISSUE #1: Missing userId parameter

**Severity:** LOW
**Impact:** Cannot request data for specific users (API v1 only)

**Current:** No userId parameter
**API Spec:** Optional userId parameter for API v1

**Recommendation:**
- Add `UserID *int` field to GetHlocParams (optional)
- Note: Only relevant for API v1, may not be needed for V2
- Low priority unless multi-user support needed

#### ⚠️ ISSUE #2: Single ticker only

**Severity:** LOW
**Impact:** Cannot request multiple tickers in one call

**Current:** Accepts single symbol string
**API Spec:** id parameter can be "multiple tickers separated by commas"

**Recommendation:**
- Accept `[]string` symbols parameter
- Join with commas like GetQuotes does
- Would reduce API calls for bulk candle requests

#### ⚠️ ISSUE #3: Fixed count value

**Severity:** LOW
**Impact:** Cannot get extra candles before/after interval

**Current:** Always uses `Count: -1` (no extra candles)
**API Spec:** Supports negative values to get candles BEFORE interval

**Use Case:**
- Get 100 extra candles before start date: `count: -100`
- Useful for calculating indicators that need historical data

**Recommendation:**
- Add count parameter to GetCandles() signature
- Default to -1 if not specified
- Document the negative value behavior

#### ⚠️ ISSUE #4: Response not parsed

**Severity:** LOW
**Impact:** Returns unstructured data

**Missing:**
- No parsing of hloc array
- No extraction of volumes (vl)
- No timestamp conversion (xSeries in seconds)
- No ticker info extraction
- No typed Candle struct

**Recommendation:**
1. Create Candle struct with High, Low, Open, Close, Volume, Timestamp
2. Parse response to extract all data
3. Convert timestamps from seconds to time.Time
4. Return typed []Candle instead of interface{}

### Summary:

**Overall Confidence:** MEDIUM ⚠️

**What Works:**
- ✅ Correct command name
- ✅ Correct parameter names
- ✅ Correct date format
- ✅ Correct timeframe conversion (seconds → minutes)
- ✅ Clean Go API (time.Time parameters)

**Critical Issues:**
- 🔴 **BUG-007:** Uses "OpenRay" instead of "ClosedRay" (CRITICAL FIX NEEDED)

**Missing/Enhancements:**
- ⚠️ userId parameter (optional, low priority)
- ⚠️ Multiple tickers support (low priority)
- ⚠️ Configurable count parameter (low priority)
- ⚠️ Response parsing (low priority)

**Priority:** HIGH - Fix intervalMode bug immediately

**Recommendations:**
1. **URGENT:** Change "OpenRay" to "ClosedRay" in both files
2. Test if API accepts both values or only "ClosedRay"
3. Add count parameter for flexible candle retrieval
4. Parse response to return typed Candle data
5. Add multi-ticker support if needed

---

## 12. ✅ Stock ticker search - `tickerFinder`

**Command:** `tickerFinder`
**Method:** HTTPS GET (plain request, no authentication required)

### Request Parameters:

```json
{
    "cmd": "tickerFinder",
    "params": {
        "text": "AAPL.US"  // string - search phrase
    }
}
```

### Parameter Descriptions:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `text` | string | Yes | Search phrase (lowercase recommended). Can use format `<ticker>@<market>` to filter by exchange |

### Exchange Filtering:

Search on specific exchange using format: `"<ticker>@<market>"`

**Market codes:**
- `MCX` - MICEX
- `FORTS` - MICEX Derivatives
- `FIX` - NYSE/NASDAQ
- `UFORTS` - Ukrainian Derivatives Exchange
- `UFOUND` - Ukrainian Exchange
- `EU` - Europe
- `KASE` - Kazakhstan

**Examples:**
- `"AAPL@FIX"` - Search for AAPL on NYSE/NASDAQ
- `"EUR@EU"` - Search for EUR on European exchanges
- `"AAPL.US"` - General search (no exchange filter)

### Response Structure (Success):

```typescript
type TickerFinderDataRow = {
    instr_id: number,    // Unique ticker ID
    nm: string,          // Name (too long)
    n: string,           // Name
    ln: string,          // English name
    t: string,           // Ticker in Tradernet's system
    isin: string,        // Ticker ISIN code
    type: number,        // Instrument type
    kind: number,        // Instrument sub-type
    tn: string,          // Ticker plus name
    code_nm: string,     // Exchange ticker
    mkt_id: number,      // Market code
    mkt: string          // Market
}

type TickerFinderResult = {
    found: TickerFinderDataRow[]  // List of found tickers (max 30)
}
```

### Instrument Type/Kind Classification:

| Type | Kind | Description |
|------|------|-------------|
| 1 | 1 | Regular stock |
| 1 | 2 | Preferred stock |
| 1 | 7 | Investment units |
| 2 | - | Bonds |
| 3 | - | Futures |
| 5 | - | Exchange index |
| 6 | 1 | Cash |
| 6 | 8 | Crypto |
| 8, 9, 10 | - | Repo |

### Response Limits:

- **Maximum results:** 30 securities per search
- **Empty results:** Returns empty array if no matches found

### Our Implementation Review:

**Files:**
- `internal/clients/tradernet/sdk/methods.go:310-322` - FindSymbol() method
- `internal/clients/tradernet/transformers.go:345-460` - transformSecurityInfo() function

#### FindSymbol() Method

```go
func (c *Client) FindSymbol(symbol string, exchange *string) (interface{}, error) {
    text := symbol
    if exchange != nil {
        text = fmt.Sprintf("%s@%s", symbol, *exchange)
    }
    params := map[string]interface{}{
        "text": text,
    }
    return c.plainRequest("tickerFinder", params)
}
```

**Analysis:**
✅ **Correct:** Uses correct command "tickerFinder"
✅ **Correct:** Parameter name "text" matches API
✅ **Correct:** Uses plainRequest (no authentication needed)
✅ **Correct:** Handles exchange filtering with @ format
✅ **Clean API:** Optional exchange parameter (nullable pointer)
✅ **Correct:** Converts symbol to lowercase? (No - but docs recommend lowercase)

#### transformSecurityInfo() Function

```go
func transformSecurityInfo(sdkResult interface{}) ([]SecurityInfo, error) {
    // Handles both "found" (raw API) and "result" (normalized) formats
    // Maps short field names to full names with priority-based fallback:
    // - Symbol: "t" (short) → "symbol" (full)
    // - Name: "nm" (short) → "name" (full)
    // - Currency: "x_curr" (short) → "currency" (full)
    // - Market: "mkt" (short) → "market" (full)
    // - ExchangeCode: "codesub" (short) → "exchange_code" (full)
    // - ISIN: same in both formats
}
```

**Analysis:**
✅ **Handles both formats:** "found" (tickerFinder) and "result" (normalized)
✅ **Comprehensive fallbacks:** Tries short names first, falls back to full names
✅ **Defensive:** Skips items without symbol
✅ **Safe:** All optional fields use pointers, check for nil/empty
✅ **Tested:** Multiple test cases verify field mapping
⚠️ **Missing fields:** Doesn't extract instr_id, type, kind, code_nm, mkt_id

#### Field Mapping Summary:

| API Field | Short Form | Extracted As | Notes |
|-----------|------------|--------------|-------|
| Ticker | `t` | Symbol | ✅ Required field |
| Name | `nm`, `n`, `ln`, `tn` | Name | ✅ Extracts `nm` |
| ISIN | `isin` | ISIN | ✅ Extracted |
| Currency | `x_curr` | Currency | ✅ Extracted |
| Market | `mkt` | Market | ✅ Extracted |
| Exchange Code | `codesub` | ExchangeCode | ✅ Extracted |
| Unique ID | `instr_id` | - | ❌ NOT extracted |
| Instrument Type | `type` | - | ❌ NOT extracted |
| Instrument Kind | `kind` | - | ❌ NOT extracted |
| Exchange Ticker | `code_nm` | - | ❌ NOT extracted |
| Market ID | `mkt_id` | - | ❌ NOT extracted |

### Issues Found:

#### ⚠️ ISSUE #1: Missing instrument classification fields

**Severity:** LOW
**Impact:** Cannot filter or categorize securities by type (stock vs bond vs future, etc.)

**Missing Fields:**
- `type` - Instrument type (1=stock, 2=bond, 3=future, 5=index, 6=cash/crypto, 8-10=repo)
- `kind` - Instrument sub-type (1=regular, 2=preferred, 7=investment units, 8=crypto)

**Use Case:**
- Filtering search results by security type
- Categorizing results in UI (stocks vs bonds vs futures)
- Validation (ensure trading only stocks, not futures)

**Recommendation:**
- Add Type and Kind fields to SecurityInfo struct
- Extract from API response
- Add helper function to get type description
- Low priority unless categorization needed

#### ⚠️ ISSUE #2: Missing unique identifier

**Severity:** LOW
**Impact:** Cannot reference securities by unique ID

**Missing Field:**
- `instr_id` - Unique ticker ID in Tradernet system

**Use Case:**
- Stable identifier across ticker symbol changes
- Linking related data by ID instead of symbol

**Recommendation:**
- Add InstrumentID field if stable IDs needed
- Currently using symbol as identifier (works for most cases)
- Low priority

#### ⚠️ ISSUE #3: Missing exchange ticker

**Severity:** LOW
**Impact:** Cannot get official exchange ticker (different from Tradernet ticker)

**Missing Field:**
- `code_nm` - Exchange ticker (may differ from Tradernet's `t` field)

**Example:**
- Tradernet ticker: `AAPL.US`
- Exchange ticker (code_nm): `AAPL`

**Recommendation:**
- Add ExchangeTicker field if needed for regulatory reporting
- Low priority

#### 📝 ENHANCEMENT: Lowercase conversion

**Observation:** API docs recommend lowercase for search text
**Current:** We don't convert symbol to lowercase
**PHP Example:** `phrase.toLowerCase()`

**Recommendation:**
- Consider adding `.toLowerCase()` conversion in FindSymbol()
- Test if API is case-sensitive
- May not be necessary if API handles both cases

### Summary:

**Overall Confidence:** HIGH ✅

**What Works:**
- ✅ Correct command name
- ✅ Correct parameter name and format
- ✅ Exchange filtering with @ syntax
- ✅ Uses plainRequest (no auth needed)
- ✅ Handles both response formats (found/result)
- ✅ Comprehensive field name fallbacks
- ✅ Defensive programming (skips invalid items)
- ✅ All optional fields properly handled
- ✅ Well-tested (multiple test cases)

**Missing/Enhancements:**
- ⚠️ Missing instrument type/kind classification (low priority)
- ⚠️ Missing unique instr_id (low priority)
- ⚠️ Missing code_nm (exchange ticker) (low priority)
- 📝 Could add lowercase conversion for search text

**Priority:** No critical issues - works excellently as-is

**Recommendations:**
1. Add Type and Kind fields if security classification needed
2. Add lowercase conversion if API is case-sensitive
3. Consider extracting InstrumentID for stable references
4. Current implementation is production-ready and well-designed

---

## 13. ✅ Get market status - `getMarketStatus`

**Command:** `getMarketStatus`
**Method:** HTTPS GET

### Request Parameters:

```json
{
    "cmd": "getMarketStatus",
    "params": {
        "market": "*",         // string - Market identifier (briefName)
        "mode": "demo"         // string|null - Optional request mode
    }
}
```

### Parameter Descriptions:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `market` | string | Yes | Market identifier (briefName). Use "*" for all markets. See market list below. |
| `mode` | string\|null | Optional | Request mode: "demo". If not specified, shows market statuses for real users. |

### Market List (Partial - 60+ markets available):

| Market Code (briefName) | Full Title | Abbreviated Name |
|------------------------|------------|------------------|
| `*` | All markets | - |
| `FIX` | NYSE/NASDAQ | FIX |
| `MOEX` | MICEX Stock market | MCX |
| `FORTS` | FORTS Market FORTS | FORTS |
| `EU` | EU Europe | EU |
| `KASE` | Kazakhstan Stock Exchange | KASE |
| `SPBEX` | SPB Russian securities | SPBEX |
| `SPBFOR` | SPB Foreign securities | SPBFOR |
| `HKEX` | Hong Kong Stock Exchange | HKEX |
| `CME` | Chicago Mercantile Exchange | CME |
| `CRPT` | Cryptocurrency market | CRPT |
| ... | (60+ total markets) | ... |

### Response Structure (Success):

```typescript
type MarketInfoRow = {
    n: string,   // Full market name
    n2: string,  // Market abbreviation
    s: string,   // Current market status (e.g., "CLOSE", "OPEN")
    o: string,   // Market opening time (MSK format: "HH:MM:SS")
    c: string,   // Market closing time (MSK format: "HH:MM:SS")
    dt: string   // Time difference vs MSK time (in minutes, e.g., "-180")
}

type MarketStatusResponse = {
    result: {
        markets: {
            t: string,              // Current request time
            m: MarketInfoRow[]      // Array of market statuses
        }
    }
}
```

### Example Response:

```json
{
    "result": {
        "markets": {
            "t": "2020-11-18 19:29:27",
            "m": [
                {
                    "n": "KASE",
                    "n2": "KASE",
                    "s": "CLOSE",
                    "o": "08:20:00",
                    "c": "14:00:00",
                    "dt": "-180"
                }
            ]
        }
    }
}
```

### Response Fields:

| Field | Description |
|-------|-------------|
| `t` | Current request time (timestamp) |
| `m` | Array of market status objects |
| `n` | Full market name |
| `n2` | Market abbreviation |
| `s` | Current market status (OPEN, CLOSE, etc.) |
| `o` | Market opening time in MSK timezone (HH:MM:SS) |
| `c` | Market closing time in MSK timezone (HH:MM:SS) |
| `dt` | Time difference relative to MSK in minutes (negative = behind, positive = ahead) |

### Response (Error):

```json
// Common error
{
    "errMsg": "Bad json",
    "code": 2
}

// Method error
{
    "error": "Something wrong, service unavailable",
    "code": 14
}
```

**Error Codes:**
- **Code 2:** Common error (bad JSON, unsupported method)
- **Code 14:** Service unavailable

### Our Implementation Status:

**Our method:** `GetMarketStatus(market string, mode *string)`

**Status:** ⏳ PENDING VERIFICATION

---

## 14. ⏳ News on securities - `getNews`

**STATUS:** AWAITING DOCUMENTATION

---

## 15. ✅ Getting most traded securities - `getTopSecurities`

**Command:** `getTopSecurities`
**Method:** HTTPS POST

### Request Parameters:

```json
{
    "cmd": "getTopSecurities",
    "params": {
        "type": "stocks",       // string - Instrument type
        "exchange": "europe",   // string - Stock exchange
        "gainers": 1,          // int (boolean) - List type
        "limit": 10            // int - Number of instruments (default: 10)
    }
}
```

### Parameter Descriptions:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `type` | string | Yes | Instrument type (see table below) |
| `exchange` | string | Yes | Stock exchange (see table below) |
| `gainers` | int (boolean) | Yes | List type: 1 = top fastest-growing, 0 = top by trading volume |
| `limit` | int | Yes | Number of instruments to display (default: 10) |

### Instrument Types:

| Value | Description |
|-------|-------------|
| `stocks` | Stocks |
| `bonds` | Bonds |
| `futures` | Futures |
| `funds` | Funds |
| `indexes` | Indices |

### Available Exchanges:

| Value | Description |
|-------|-------------|
| `kazakhstan` | Kazakhstan |
| `europe` | Europe |
| `usa` | USA |
| `ukraine` | Ukraine |
| `currencies` | Currency |

### List Types:

| Value | Description |
|-------|-------------|
| `1` | Top fastest-growing (by year) |
| `0` | Top by trading volume |

**Note:** Fastest-growing list (gainers=1) is only available for stocks.

### Response Structure (Success):

```json
{
    "tickers": [
        "AAPL.US",
        "T.US",
        "F.US"
    ]
}
```

**Response:**
- Returns array of ticker symbols
- Maximum length determined by `limit` parameter

### Response (Error):

```json
{
    "code": 5,
    "error": "Exchange is missing",
    "errMsg": "Exchange is missing"
}
```

**Error Codes:**
- **Code 5:** Exchange is missing (invalid or missing exchange parameter)

### Our Implementation Status:

**Our method:** `GetMostTraded(secType string, exchange string, gainers bool, limit int)`

**Status:** ⏳ PENDING VERIFICATION

---

## 16. ⏳ Options - `getOptionsByMktNameAndBaseAsset`

**STATUS:** AWAITING DOCUMENTATION

---

## 17. ⏳ Get price alerts - `getAlertsList`

**STATUS:** AWAITING DOCUMENTATION

---

## 18. ⏳ Add price alert - `addPriceAlert`

**STATUS:** AWAITING DOCUMENTATION

---

## 19. ⏳ Delete price alert

**STATUS:** AWAITING DOCUMENTATION

---

## 20. ⏳ Receiving broker report - `getBrokerReport`

**STATUS:** AWAITING DOCUMENTATION

---

## 21. ⏳ Receiving order files - `getCpsFiles`

**STATUS:** AWAITING DOCUMENTATION

---

## 22. ⏳ Exchange rate by date

**STATUS:** AWAITING DOCUMENTATION

---

## 23. ⏳ List of currencies

**STATUS:** AWAITING DOCUMENTATION

---

## 24. ✅ Initial user data - `getOPQ`

**Command:** `getOPQ`
**Method:** HTTPS GET
**Authentication:** Required (SID parameter)

### Description:

Returns **complete initial user data** from server in a single request, including:
- Portfolio positions and cash balances
- Active orders
- Real-time quotes for user's securities
- Market statuses
- Open sessions
- User account information
- User preferences/options
- User watchlists

This is a comprehensive initialization endpoint for loading complete user state.

### Request Parameters:

```json
{
    "cmd": "getOPQ",
    "SID": "[SID by authorization]",
    "params": {}
}
```

### Parameter Descriptions:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `cmd` | string | Yes | Request execution command |
| `SID` | string | Yes | SID received during user's authorization |
| `params` | array | Yes | Request execution parameters (empty object) |

### Response Structure (Success):

**Note:** This is an extremely comprehensive response with nested objects. Key sections shown below.

```typescript
type OPQResponse = {
    OPQ: {
        rev: number,              // Revision number
        init_margin: number,      // Initial margin
        brief_nm: string,         // Brief name
        reception: number,        // Reception ID
        active: number,           // Active status
        quotes: {                 // Real-time quotes
            q: QuoteObject[]
        },
        ps: {                     // Portfolio & positions
            loaded: boolean,
            acc: AccountBalance[],
            pos: Position[]
        },
        orders: {                 // Active orders
            loaded: boolean,
            order: Order[]
        },
        sess: any[],             // Sessions
        markets: {               // Market statuses
            markets: {
                t: string,
                m: MarketStatus[]
            }
        },
        source: string,
        offbalance: {            // Off-balance positions
            net_assets: number,
            pos: any[],
            acc: any[]
        },
        homeCurrency: string,    // Home currency (e.g., "USD")
        userLists: {             // User watchlists
            userStockLists: {
                default: string[]
            },
            userStockListSelected: string,
            stocksArray: string[]
        },
        NO_ORDER_GROWLS: string|null,
        userInfo: UserInfo,      // Complete user profile
        userOptions: UserOptions // User preferences
    }
}
```

### Key Response Sections:

#### 1. Quotes (`quotes.q[]`)
Real-time quote data for all user's securities. Each quote contains 80+ fields including:
- Price data: `ltp`, `bbp`, `bap`, `op`, `pp`, `mintp`, `maxtp`
- Volume: `vol`, `vlt`, `trades`
- Change: `chg`, `pcp`, `chg5`, `chg22`, `chg110`, `chg220`
- Security details: `name`, `issue_nb`, `x_curr`, `min_step`
- Market status: `marketStatus`, `ltt`
- And 70+ additional fields

#### 2. Portfolio (`ps`)
**Account balances** (`ps.acc[]`):
```typescript
{
    curr: string,         // Currency
    currval: number,      // Exchange rate
    s: number,           // Available funds
    forecast_in: number,
    forecast_out: number,
    t2_in: number,
    t2_out: number
}
```

**Positions** (`ps.pos[]`):
```typescript
{
    i: string,            // Ticker
    q: number,           // Quantity
    mkt_price: number,   // Market price
    bal_price_a: number, // Average price
    profit_close: number, // Closed profit
    profit_price: number, // Current profit
    market_value: number,
    curr: string,        // Currency
    // ... 20+ additional fields
}
```

#### 3. Orders (`orders.order[]`)
Array of active orders (same structure as `getNotifyOrderJson`)

#### 4. Markets (`markets.markets.m[]`)
Market status information for all markets:
```typescript
{
    n: string,    // Market name
    n2: string,   // Market abbreviation
    s: string,    // Status (OPEN/CLOSE)
    o: string,    // Opening time
    c: string,    // Closing time
    dt: number,   // Time difference from MSK
    date: Array<{from, to, dayoff, desc}>, // Holidays
    ev: Array<{id, t, next}>               // Market events
}
```

#### 5. User Info (`userInfo`)
Complete user profile with 60+ fields:
- Personal: `id`, `login`, `email`, `firstname`, `lastname`
- Status: `status`, `type`, `f_active`, `f_demo`
- Account: `client_id`, `trader_systems_id`, `reception`
- Documents: `numdoc`, `docseries`, `inn`
- Dates: `birthday`, `date_open_real`, `rec_tmstmp`
- Details: Nested object with additional metadata
- Tariff: `tariffDetails` with plan information
- Messages: `messages_counts` with unread count

#### 6. User Options (`userOptions`)
User preferences and settings:
```typescript
{
    cost_open: number,
    cost_last: number,
    graphic_type: number,
    graphic_format: string,      // "Candlestick"
    period: string,              // "Y1"
    interval: string,            // "D1"
    theme: string,               // "light"
    showPortfolioBlock: number,
    gridPortfolio: string[],     // Column preferences
    // ... 20+ preference fields
}
```

#### 7. User Lists (`userLists`)
Watchlists and selected securities:
```typescript
{
    userStockLists: {
        default: string[]        // Array of tickers
    },
    userStockListSelected: string, // Selected list name
    stocksArray: string[]         // Combined array
}
```

### Response (Error):

```json
// Common error
{
    "errMsg": "Bad json",
    "code": 2
}

// Method error
{
    "error": "User is not found",
    "code": 7
}
```

**Error Codes:**
- **Code 2:** Common error (bad JSON, unsupported method)
- **Code 7:** User not found

### Important Notes:

⚠️ **Massive response** - This endpoint returns everything in one call (can be several MB)
⚠️ **Requires authentication** - Must have valid SID from login
⚠️ **Initial load only** - Typically called once at app startup
⚠️ **Use subscriptions** - For real-time updates, use WebSocket subscriptions instead of polling this endpoint

### Typical Usage Pattern:

1. **App startup:** Call `getOPQ` to load initial state
2. **Real-time updates:** Subscribe to WebSocket for incremental updates
3. **Avoid polling:** Don't repeatedly call this endpoint - it's expensive

### Our Implementation Status:

**Our method:** `GetUserData()`

**Status:** ⏳ PENDING VERIFICATION

---

## 25. ✅ Instruments details - `getSecurityInfo`

**Command:** `getSecurityInfo`
**Method:** HTTPS GET / HTTPS POST (API V2)

### Request Parameters:

```json
{
    "cmd": "getSecurityInfo",
    "params": {
        "ticker": "AAPL.US",  // string - Required
        "sup": true           // bool - Required
    }
}
```

### Parameter Descriptions:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `ticker` | string | Yes | The ticker symbol to retrieve information for |
| `sup` | bool | Yes | IMS and trading system format (boolean, NOT int!) |

### Response (Success):

```json
{
    "id": 2772,
    "short_name": "Apple Inc.",
    "default_ticker": "AAPL",
    "nt_ticker": "AAPL.US",
    "firstDate": "02.01.1990",
    "currency": "USD",
    "min_step": 0.01000,
    "code": 0
}
```

### Response Fields:

| Field | Type | Description |
|-------|------|-------------|
| `id` | number | Unique ticker ID in Tradernet system |
| `short_name` | string | Short ticker name (company name) |
| `default_ticker` | string | Ticker name on the Exchange |
| `nt_ticker` | string | Ticker name in Tradernet system |
| `firstDate` | string | Company registration date in stock exchange (DD.MM.YYYY) |
| `currency` | string | Currency of the security |
| `min_step` | number | Minimum price increment of the security |
| `code` | number | Error code (0 = success) |

### Response (Error):

```json
// Common error
{
    "errMsg": "Bad json",
    "code": 2
}

// Method error
{
    "error": "User is not found",
    "code": 7
}
```

### Our Implementation Review:

**Files:**
- `internal/clients/tradernet/sdk/models.go:64-70` - GetSecurityInfoParams struct
- `internal/clients/tradernet/sdk/methods.go:327-333` - SecurityInfo() method

#### SecurityInfo() Method

```go
func (c *Client) SecurityInfo(symbol string, sup bool) (interface{}, error) {
    params := GetSecurityInfoParams{
        Ticker: symbol,
        Sup:    sup, // Boolean, NOT int!
    }
    return c.authorizedRequest("getSecurityInfo", params)
}
```

**Analysis:**
✅ **Correct:** Uses correct command "getSecurityInfo"
✅ **Correct:** Parameter names match API (ticker, sup)
✅ **Correct:** `sup` parameter stays as `bool` (NOT converted to int) - this is documented in code
✅ **Correct:** Clean Go API (uses "symbol" parameter name, maps to "ticker" in JSON)
✅ **Correct:** Uses `authorizedRequest()` (requires auth session)

#### GetSecurityInfoParams Struct

```go
type GetSecurityInfoParams struct {
    Ticker string `json:"ticker"` // Field 1
    Sup    bool   `json:"sup"`    // Field 2 - Boolean (NOT int!)
}
```

**Analysis:**
✅ **Correct:** Field names match API exactly
✅ **Correct:** Both fields required (non-nullable)
✅ **Correct:** `Sup` is `bool` type (API explicitly expects boolean)
✅ **Correct:** Well-documented that boolean stays boolean

### Summary:

**Overall Confidence:** HIGH ✅

**What Works:**
- ✅ Correct command and parameter names
- ✅ Correct parameter types (ticker=string, sup=bool)
- ✅ Boolean stays boolean (NOT converted to int like other endpoints)
- ✅ Clean Go API with semantic parameter name
- ✅ Used successfully in production

**Missing/Enhancements:**
- ⚠️ Response not parsed (returns `interface{}`)
- ⚠️ No typed response struct for security info fields
- ⚠️ No validation of response code field

**Priority:** No critical issues - works as expected

**Recommendations:**
1. Create SecurityInfo struct matching response structure
2. Parse response to extract all fields
3. Validate `code` field (0 = success)
4. Return typed result instead of `interface{}`

---

---

## 26. ✅ Directory of securities - `getAllSecurities` / `getReadyList`

**Command:** `getAllSecurities` (also possibly `getReadyList`)
**Method:** HTTPS POST (required - GET not supported)
**Alternative REST endpoint:** `/securities/ajax-get-all-securities/?take=20&skip=0`

### Request Parameters:

```json
{
    "cmd": "getAllSecurities",
    "params": {
        "take": 10,          // int - Number of securities to retrieve
        "skip": 0,           // int - Pagination offset
        "sort": [{           // array - Optional sorting
            "field": "ticker",
            "dir": "ASC"
        }],
        "filter": {          // object - Optional filtering
            "filters": [{
                "field": "ticker",
                "operator": "eq",
                "value": "AAPL.US"
            }]
        }
    }
}
```

### Parameter Descriptions:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `take` | int | Optional | Number of securities to retrieve |
| `skip` | int | Optional | Pagination offset (for paging) |
| `sort` | array | Optional | Array of sort objects with `field` and `dir` properties |
| `filter` | object | Optional | Filter object with `filters` array |

### Filterable/Sortable Fields:

| Field | Type | Description | Notes |
|-------|------|-------------|-------|
| `ticker` | string | Ticker in Tradernet system | |
| `instr_type` | string | Instrument type | |
| `instr_kind` | string | Instrument sub-type | See Instruments details |
| `instr_kind_c` | integer | Instrument subtype code | See Instruments details |
| `code_sec` | string | Day trading code | |
| `code_rep` | string | Night trading code | |
| `code_nm` | string | Instrument ticker on exchange | |
| `name` | string | Name | For sorting only |
| `reg_nb` | string | Registration number | |
| `issue_nb` | string | Issue number (ISIN) | |
| `face_curr_c` | string | Currency | See Currency reference |
| `mkt_id` | integer | Market ID | See Trading platforms |
| `mkt_name` | string | Market code | See Trading platforms |
| `mkt_short_code` | string | Market short code | See Trading platforms |
| `fv` | numeric | Bond body value | |
| `step_price` | numeric | Price increment | |
| `x_short` | boolean | Available for short selling | |

### Filter Operators:

| Operator | Description |
|----------|-------------|
| `eq` | Equal (default) |
| `neq` | Is not equal |
| `eqormore` | Greater than or equal to |
| `eqorless` | Less than or equal to |
| `isempty` | For empty data |
| `contains` | Pattern matching anywhere in string (ILIKE %%) |
| `doesnotcontain` | Exclude pattern matching (NOT ILIKE %%) |
| `startswith` | String-start pattern matching (ILIKE %) |
| `endswith` | String-end pattern matching (ILIKE %) |
| `in` | Multiple values separated by commas (IN ()) |

### Response Structure (Success):

**Note:** Response is extremely comprehensive with many fields. Key fields shown below.

```json
{
    "securities": [{
        "ticker": "AAPL.US",
        "instr_type_c": 1,
        "instr_type": "Ordinary stock",
        "instr_kind_c": 1,
        "instr_kind": "Share",
        "instr_id": "40000001",
        "code_nm": "AAPL",
        "name": "Apple Inc.",
        "name_alt": "Apple Inc.",
        "issue_nb": "US0378331005",
        "face_curr_c": "USD",
        "mkt_id": "30000000001",
        "mkt_name": "FIX",
        "lot_size_q": "1.00000000",
        "istrade": 1,
        "fv": "100.00000000",
        "x_short": 1,
        "step_price": "0.01000000",
        "min_step": "0.01000000",
        "quotes": "{...}",        // JSON string with quote data
        "attributes": "{...}",    // JSON string with attributes
        "rate": 10893,
        "id": 113
    }],
    "total": 1
}
```

### Key Response Fields:

| Field | Type | Description |
|-------|------|-------------|
| `ticker` | string | Internal ticker in Tradernet system |
| `instr_type` / `instr_type_c` | string/int | Instrument type |
| `instr_kind` / `instr_kind_c` | string/int | Instrument sub-type |
| `code_nm` | string | Stock exchange ticker |
| `name` / `name_alt` | string | Full name / Short name |
| `issue_nb` | string | ISIN code |
| `face_curr_c` | string | Currency code |
| `mkt_name` / `mkt_short_code` | string | Market identifier |
| `x_short` | int | Short selling permitted (1=yes, 0=no) |
| `step_price` / `min_step` | string | Price increment |
| `quotes` | JSON string | Embedded quote data (must be parsed) |
| `attributes` | JSON string | Embedded attributes (must be parsed) |
| `total` | int | Total number of matching securities |

### Embedded Quotes Object:

The `quotes` field contains a JSON string with comprehensive quote data including:
- Price data: `ltp`, `bbp`, `bap`, `op`, `pp`, `mintp`, `maxtp`
- Volume: `vol`, `vlt`, `trades`
- Change: `chg`, `pcp`, `chg5`, `chg22`, `chg110`, `chg220`
- Bond data: `acd`, `cpn`, `yld`, `yld_ytm_ask`, `yld_ytm_bid`
- And 50+ additional fields

### Usage Examples:

**Complete list (no pagination):**
```json
{
    "cmd": "getAllSecurities",
    "params": {}
}
```

**Paginated request:**
```json
{
    "cmd": "getAllSecurities",
    "params": {
        "take": 10,
        "skip": 0
    }
}
```

**With sorting:**
```json
{
    "cmd": "getAllSecurities",
    "params": {
        "take": 10,
        "skip": 0,
        "sort": [{
            "field": "ticker",
            "dir": "ASC"
        }]
    }
}
```

**With filtering:**
```json
{
    "cmd": "getAllSecurities",
    "params": {
        "take": 10,
        "skip": 0,
        "filter": {
            "filters": [{
                "field": "ticker",
                "operator": "eq",
                "value": "AAPL.US"
            }]
        }
    }
}
```

### Important Notes:

⚠️ **POST method required** - GET is NOT supported for this endpoint
⚠️ **JSON strings in response** - `quotes` and `attributes` fields are JSON strings that must be parsed
⚠️ **Large response** - Requesting all securities without pagination can return massive datasets

### Our Implementation Status:

**Our method:** `Symbols()` uses `getReadyList` command (may be different from `getAllSecurities`)

**Status:** ⏳ PENDING VERIFICATION - Need to determine if `getAllSecurities` and `getReadyList` are the same endpoint or different

---

## 27. ⏳ API key authentication

**STATUS:** AWAITING DOCUMENTATION

---

---

# BUGS & ISSUES TRACKING

## Critical Bugs Found

### 🔴 BUG-001: Missing stop_price parameter in putTradeOrder
- **Severity:** CRITICAL
- **Impact:** Cannot place stop orders or stop-limit orders
- **Fix Required:** Add `stop_price` parameter to `PutTradeOrderParams`
- **Files Affected:**
  - `internal/clients/tradernet/sdk/models.go`
  - `internal/clients/tradernet/sdk/methods.go`

### 🔴 BUG-002: Incomplete order_type_id support
- **Severity:** HIGH
- **Impact:** Only supports Market (1) and Limit (2), missing Stop (3) and Stop Limit (4)
- **Fix Required:** Update order type logic to support all 4 types
- **Files Affected:**
  - `internal/clients/tradernet/sdk/methods.go` (Trade method)

### ⚠️ BUG-003: limit_price should be nullable
- **Severity:** MEDIUM
- **Impact:** Sends 0 instead of null for market orders (may work but not clean)
- **Fix Required:** Change `LimitPrice float64` to `LimitPrice *float64`
- **Files Affected:**
  - `internal/clients/tradernet/sdk/models.go`

### ⚠️ BUG-004: Cancel() lacks error code handling
- **Severity:** LOW-MEDIUM
- **Impact:** Doesn't handle specific error codes from delTradeOrder
- **Error Codes:**
  - Code 12: Permission denied (order can only be cancelled through traders)
  - Code 0: Security type not identified
  - Code 2: Unsupported query method
- **Fix Required:** Add error parsing and specific handling for code 12
- **Files Affected:**
  - `internal/clients/tradernet/sdk/methods.go` (Cancel method)
- **Recommendations:**
  1. Parse error response to extract error code
  2. Add specific handling for code 12 (permission errors)
  3. Add warning logs for code 0 and 12
  4. Consider retry logic for code 2
  5. Validate response contains expected `order_id`

### ⚠️ BUG-005: putStopLoss percent fields use wrong type
- **Severity:** LOW-MEDIUM
- **Impact:** Percent fields use `*int` instead of `*float64` as API spec requires
- **Current:** `StopLossPercent *int` and `StoplossTrailingPercent *int`
- **Expected:** Both should be `*float64` (API spec: `null|float`)
- **Consequence:** Cannot use fractional percentages (e.g., 0.5% trailing stop)
- **Fix Required:** Change both fields to `*float64`
- **Files Affected:**
  - `internal/clients/tradernet/sdk/models.go:197-198` (PutStopLossParams struct)
  - `internal/clients/tradernet/sdk/methods.go:1167-1176` (TrailingStop method signature)

### ⚠️ BUG-006: getTradesHistory missing reception parameter
- **Severity:** LOW
- **Impact:** Cannot filter trades by office ID
- **Current:** No `reception` field in GetTradesHistoryParams
- **Expected:** `Reception *int` field (API spec: optional office ID filter)
- **Consequence:** Cannot filter trades by specific office/branch (multi-office setups)
- **Fix Required:** Add `Reception *int` field to struct and method signature
- **Files Affected:**
  - `internal/clients/tradernet/sdk/models.go:37-44` (GetTradesHistoryParams struct)
  - `internal/clients/tradernet/sdk/methods.go:266-276` (GetTradesHistory method)

### 🔴 BUG-007: getHloc uses "OpenRay" instead of "ClosedRay"
- **Severity:** CRITICAL
- **Impact:** API calls may fail or return unexpected/incorrect candlestick data
- **Current:** `IntervalMode: "OpenRay"`
- **Expected:** `IntervalMode: "ClosedRay"` (API spec: required parameter, single value)
- **Evidence:** API documentation explicitly states "required parameter, a single value ClosedRay"
- **Consequence:**
  - May cause API failures
  - May return incorrect interval data
  - Data integrity issue for historical charts
- **Fix Required:**
  ```go
  // Change in methods.go:305
  IntervalMode: "ClosedRay",  // ✅ CORRECT

  // Change comment in models.go:62
  IntervalMode string `json:"intervalMode"` // Field 6 - "ClosedRay"
  ```
- **Files Affected:**
  - `internal/clients/tradernet/sdk/models.go:62` (GetHlocParams struct comment)
  - `internal/clients/tradernet/sdk/methods.go:305` (GetCandles method - hardcoded value)
- **Priority:** CRITICAL - Fix immediately before deploying
- **Testing Note:** If currently working, API may accept both values (undocumented tolerance). Test both to confirm.

## Questions Pending Docs

### ❓ QUESTION-001: Which profit field to use?
- `profit_close` (previous day) vs `profit_price` (current)
- Awaiting clarification from additional docs

### ❓ QUESTION-002: Use API-provided exchange rates?
- API provides `currval` field
- We currently fetch rates separately
- Which is more accurate/reliable?

---

# VERIFICATION PROGRESS

- [x] putTradeOrder - Documented
- [x] getPositionJson - Documented
- [x] getClientCpsHistory - Documented
- [x] getUserCashFlows - Documented (NOT IMPLEMENTED in our SDK)
- [x] delTradeOrder - Documented
- [x] putStopLoss - Documented
- [x] getNotifyOrderJson - Documented
- [x] getOrdersHistory - Documented
- [x] getTradesHistory - Documented
- [x] getStockQuotesJson - Documented (needs testing)
- [x] getHloc - Documented (CRITICAL BUG FOUND)
- [x] tickerFinder - Documented
- [x] getMarketStatus - Documented
- [ ] getNews
- [x] getTopSecurities - Documented
- [ ] getOptionsByMktNameAndBaseAsset
- [ ] getAlertsList
- [ ] addPriceAlert
- [ ] Delete price alert
- [ ] getBrokerReport
- [ ] getCpsFiles
- [ ] Exchange rate by date
- [ ] List of currencies
- [ ] getOPQ
- [x] getSecurityInfo - Documented
- [x] getAllSecurities/getReadyList - Documented
- [ ] API authentication

**Progress:** 16/27 (59%)
