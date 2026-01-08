# Tradernet API - Official Documentation Reference

**Purpose:** Complete reference of Tradernet API endpoints (official documentation)
**Source:** https://tradernet.com/tradernet-api/
**Last Updated:** 2026-01-08
**Status:** Collecting comprehensive documentation

---

## TABLE OF CONTENTS

### AUTHENTICATION & SESSION
- [Login/Password](#loginpassword)
- [API key](#api-key)
- [Initial user data](#initial-user-data---getopq)
- [User's current authorization session information](#users-current-authorization-session-information)
- [Public API client](#public-api-client)
- [Tradernet Python SDK](#tradernet-python-sdk)

### SECURITY SESSION
- [Get a list of open security sessions and subscribe for changes](#get-list-of-open-security-sessions)
- [Opening the security session](#opening-the-security-session)

### SET UP THE LIST OF SECURITIES
- [Receiving the lists of securities](#receiving-lists-of-securities)
- [Adding the list of securities](#adding-list-of-securities)
- [Changing the list of securities](#changing-list-of-securities)
- [Deleting the saved list of securities](#deleting-saved-list-of-securities)
- [Setting the selected list of securities](#setting-selected-list-of-securities)
- [Adding the ticker to the list](#adding-ticker-to-list)
- [Deleting the ticker from the list](#deleting-ticker-from-list)

### QUOTES AND TICKERS
- [Get updates on market status](#get-updates-on-market-status---getmarketstatus)
- [Get stock ticker data](#get-stock-ticker-data---getstockquotesjson)
- [Options demonstration](#options-demonstration---getoptionsbymktnameandbaseasset)
- [Getting the most traded securities](#getting-most-traded-securities---gettopsecurities)
- [Subscribe to stock quotes updates](#subscribe-to-stock-quotes-updates)
- [Get a quote](#get-a-quote)
- [Subscribe to market depth data](#subscribe-to-market-depth-data)
- [Get quote historical data (candlesticks)](#get-quote-historical-data---gethloc)
- [Get trades](#get-trades)
- [Retrieving trades history](#retrieving-trades-history---gettradeshistory)
- [Stock ticker search](#stock-ticker-search---tickerfinder)
- [News on securities](#news-on-securities---getnews)
- [Directory of securities](#directory-of-securities---getreadylist)
- [Checking the instruments allowed for trading](#checking-instruments-allowed-for-trading)

### PORTFOLIO
- [Getting information on a portfolio and subscribing for changes](#getting-portfolio-information---getpositionjson)

### ORDERS
- [Receive orders in the current period and subscribe for changes](#receive-orders-in-current-period---getnotifyorderjson)
- [Get orders list for the period](#get-orders-list-for-period---getordershistory)
- [Sending of order for execution](#sending-order-for-execution---puttradeorder)
- [Sending Stop Loss and Take Profit losses](#sending-stop-loss-and-take-profit---putstoploss)
- [Cancel the order](#cancel-the-order---deltradeorder)

### PRICE ALERTS
- [Get current price alerts](#get-current-price-alerts---getalertslist)
- [Add price alert](#add-price-alert---addpricealert)
- [Delete price alert](#delete-price-alert)

### REQUESTS
- [Receiving clients' requests history](#receiving-clients-requests-history---getclientcpshistory)
- [Receiving order files](#receiving-order-files---getcpsfiles)

### BROKER REPORT
- [Receiving broker report](#receiving-broker-report---getbrokerreport)
- [Getting the broker's report via a direct link](#getting-broker-report-via-direct-link)
- [Obtain a depository report](#obtain-depository-report)
- [Obtain a depository report via direct link](#obtain-depository-report-via-direct-link)
- [Money funds movement](#money-funds-movement)

### CURRENCIES
- [Exchange rate by date](#exchange-rate-by-date)
- [List of currencies](#list-of-currencies)

### WEBSOCKET - REAL-TIME DATA
- [Connecting to a websocket server](#connecting-to-websocket-server)
- [Subscribe to stock quotes updates (WebSocket)](#subscribe-to-quotes-websocket)
- [Subscribe to market depth data (WebSocket)](#subscribe-to-market-depth-websocket)
- [Subscribing to changes in security sessions](#subscribing-to-security-sessions-websocket)
- [Subscribe to portfolio updates (WebSocket)](#subscribe-to-portfolio-websocket)
- [Subscribe to orders updates (WebSocket)](#subscribe-to-orders-websocket)
- [Subscribing to changes in market statuses](#subscribing-to-market-statuses-websocket)

### VARIOUS / REFERENCE DATA
- [List of existing offices](#list-of-existing-offices)
- [Name list of the system files](#name-list-of-system-files)
- [Trading platforms](#trading-platforms)
- [Instruments details](#instruments-details---getsecurityinfo)
- [List of request types](#list-of-request-types)
- [A list of the user's profile fields](#list-of-users-profile-fields)
- [Types of documents for the application](#types-of-documents)
- [Orders statuses](#orders-statuses)
- [Types of signatures](#types-of-signatures)
- [Types of valid codes](#types-of-valid-codes)

---

---

# DOCUMENTATION

## ORDERS

### Sending order for execution - `putTradeOrder`

Send an order to execute.

**Description:** A method that allows you to work with the submission of orders for execution.

**Prerequisites:** First, open the session using the method `getAuthInfo`

**HTTP Method:** POST (for API V2)
**Command:** `putTradeOrder`

#### Request Parameters:

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

#### Parameter Details:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `instr_name` | string | Yes | The instrument used to issue an order |
| `action_id` | int | Yes | Action:<br>1 - Purchase (Buy)<br>2 - Purchase when making trades with margin (Buy on Margin)<br>3 - Sale (Sell)<br>4 - Sale when making trades with margin (Sell Short)* |
| `order_type_id` | int | Yes | Type of order:<br>1 - Market Order<br>2 - Order at a set price (Limit)<br>3 - Market Stop-order (Stop)<br>4 - Stop-order at a set price (Stop Limit) |
| `qty` | int | Yes | Quantity in the order |
| `limit_price` | null\|float | Optional | Limit price |
| `stop_price` | null\|float | Optional | Stop price |
| `expiration_id` | int | Yes | Order expiration:<br>1 - Until end of current trading session (Day)<br>2 - Day/night or night/day (Day + Ext)<br>3 - Before cancellation (GTC, before cancellation with participation in night sessions) |
| `user_order_id` | null\|int | Optional | Custom order ID |

**Important Note:** *Tradernet allows using margin at all times. The check is only carried out in terms of adequacy of the portfolio and orders collateral. The action_id field values equal to 3 or 4 are now the same in the system.

#### Response (Success):

```json
{
    "order_id": 4982349829328
}
```

#### Response (Error):

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

#### Response Fields:

| Field | Type | Description |
|-------|------|-------------|
| `order_id` | int | Order ID of the created order |

---

## PORTFOLIO

### Getting portfolio information - `getPositionJson`

Getting information on a portfolio and subscribing for changes.

**Description:** A method for obtaining information on a portfolio with the subscription to changes.

**HTTP Method:** POST (for API V2)
**Command:** `getPositionJson`

#### Request Parameters:

```json
{
    "cmd": "getPositionJson",
    "params": {}
}
```

No parameters required (empty params object).

#### Response Structure (TypeScript definitions):

```typescript
/**
 * Account Info Row (cash balance)
 */
type AccountInfoRow = {
    curr: string,          // Account currency
    currval: number,       // Account currency exchange rate
    forecast_in: number,   // Forecast incoming
    forecast_out: number,  // Forecast outgoing
    s: number,            // Available funds
    t2_in: string,        // T+2 incoming
    t2_out: string        // T+2 outgoing
}

/**
 * Position Info Row
 */
type PositionInfoRow = {
    acc_pos_id: number,      // Unique identifier of an open position in the Tradernet system
    accruedint_a: number,    // (ACI) accrued coupon income
    curr: string,            // Open position currency
    currval: number,         // Account currency exchange rate
    fv: number,              // Coefficient to calculate initial margin
    go: number,              // Initial margin per position
    i: string,               // Open position ticker
    k: number,
    q: number,               // Number of securities in the position
    s: number,
    t: number,
    t2_in: string,
    t2_out: string,
    vm: number,              // Variable margin of a position
    name: string,            // Issuer name
    name2: string,           // Issuer alternative name
    mkt_price: number,       // Open position market value
    market_value: number,    // Asset value
    bal_price_a: number,     // Open position book value
    open_bal: number,        // Position book value
    price_a: number,         // Book value of the position when opened
    profit_close: number,    // Previous day positions profit
    profit_price: number,    // Current position profit
    close_price: number,     // Position closing price
    trade: {trade_count: number}[]
}

/**
 * Portfolio Response
 */
type PortfolioResponse = {
    key: string,
    acc: AccountInfoRow[],   // Cash accounts/balances
    pos: PositionInfoRow[]   // Open positions
}
```

#### Example Response:

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
            "t": 1,
            "k": 1,
            "s": 22.4,
            "q": 100,
            "fv": 100,
            "curr": "USD",
            "currval": 1,
            "name": "Apple Inc.",
            "name2": "Apple Inc.",
            "open_bal": 299.4,
            "mkt_price": 23.81,
            "vm": ".00000000",
            "go": ".00000000",
            "profit_close": -2.4,
            "acc_pos_id": 85600002,
            "accruedint_a": ".00000000",
            "acd": ".00000000",
            "bal_price_a": 29.924,
            "price_a": 29.924,
            "base_currency": "USD",
            "face_val_a": 3,
            "scheme_calc": "T2",
            "instr_id": 10000007229,
            "Yield": ".00000000",
            "issue_nb": "US0000040",
            "profit_price": 2.83,
            "market_value": 2020,
            "close_price": 20.83
        }
    ]
}
```

#### Response (Error):

```json
// Common error
{
    "errMsg": "Unsupported query method",
    "code": 2
}
```

#### Key Response Fields:

**Account (acc) Fields:**
- `curr` - Currency code
- `currval` - Exchange rate for this currency
- `s` - Available cash balance
- `forecast_in` / `forecast_out` - Forecasted cash movements
- `t2_in` / `t2_out` - T+2 settlement cash movements

**Position (pos) Fields:**
- `i` - Ticker symbol
- `q` - Quantity held
- `curr` - Position currency
- `mkt_price` - Current market price
- `bal_price_a` - Average book price
- `profit_close` - Previous day's profit/loss
- `profit_price` - Current unrealized profit/loss
- `market_value` - Total market value of position
- `name` / `name2` - Security name

---

## REQUESTS

### Receiving clients' requests history - `getClientCpsHistory`

Receiving clients' requests (CPS - Client Payment System) history.

**Description:** A method for obtaining the history of client requests to the broker, including deposits, withdrawals, transfers, and other financial operations.

**HTTP Method:** POST (for API V2)
**Command:** `getClientCpsHistory`

#### Request Parameters:

```json
{
    "cmd": "getClientCpsHistory",
    "params": {
        "cpsDocId"   : 181,          // null|int - Optional
        "id"         : 123123123,    // null|int - Optional
        "date_from"  : "2020-04-10", // null|string|date - Optional (YYYY-MM-DD or ISO format)
        "date_to"    : "2020-05-10", // null|string|date - Optional (YYYY-MM-DD or ISO format)
        "limit"      : 100,          // null|int - Optional (pagination limit)
        "offset"     : 20,           // null|int - Optional (pagination offset)
        "cps_status" : 1             // null|int - Optional (filter by status: 0-3)
    }
}
```

#### Parameter Details:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `cpsDocId` | null\|int | Optional | Specific CPS document ID to retrieve |
| `id` | null\|int | Optional | Specific request ID to retrieve |
| `date_from` | null\|string\|date | Optional | Start date for filtering (YYYY-MM-DD or ISO format like "2011-01-11T00:00:00") |
| `date_to` | null\|string\|date | Optional | End date for filtering (YYYY-MM-DD or ISO format like "2024-01-01T00:00:00") |
| `limit` | null\|int | Optional | Maximum number of records to return (pagination) |
| `offset` | null\|int | Optional | Number of records to skip (pagination) |
| `cps_status` | null\|int | Optional | Filter by request status:<br>0 - Draft request<br>1 - In process of execution<br>2 - Request is rejected<br>3 - Request is executed |

#### CPS Status Values:

- **0** - Draft request (not yet submitted)
- **1** - In process of execution (pending)
- **2** - Request is rejected (failed)
- **3** - Request is executed (completed successfully)

#### Response Structure:

The response contains an array of CPS history records. Each record may contain the following fields:

**Common Fields:**
- `id` - Request ID
- `transaction_id` - External transaction reference
- `type_doc_id` - Document type identifier
- `type` - Transaction type description
- `transaction_type` - Alternative transaction type field
- `dt` - Transaction date (short form)
- `date` - Transaction date (full form)
- `sm` - Amount (short form, Russian "сумма" = sum/amount)
- `amount` - Amount (full form)
- `curr` - Currency code (short form)
- `currency` - Currency code (full form)
- `sm_eur` - Amount in EUR (short form)
- `amount_eur` - Amount in EUR (full form)
- `status` - Status code
- `status_c` - Status description
- `description` - Human-readable description
- `params` - Additional parameters (object/map)

**Note:** The API may return different field names for the same data (short vs full forms). Applications should handle both variants with fallback logic.

#### Example Response:

```json
[
    {
        "id": 123456,
        "transaction_id": "TXN789",
        "type_doc_id": 1,
        "type": "deposit",
        "date": "2024-01-15",
        "amount": 1000.00,
        "currency": "USD",
        "amount_eur": 920.00,
        "status": 3,
        "status_c": "executed",
        "description": "Bank transfer deposit"
    },
    {
        "id": 123457,
        "transaction_id": "TXN790",
        "type_doc_id": 2,
        "type": "withdrawal",
        "date": "2024-01-20",
        "amount": -500.00,
        "currency": "USD",
        "amount_eur": -460.00,
        "status": 1,
        "status_c": "pending",
        "description": "Withdrawal to bank account"
    }
]
```

#### Response (Error):

```json
// Common error
{
    "errMsg": "Unsupported query method",
    "code": 2
}
```

#### Key Response Fields:

- `id` - Unique request identifier
- `type` / `transaction_type` - Type of financial operation (deposit, withdrawal, dividend, etc.)
- `date` / `dt` - When the request was created/executed
- `amount` / `sm` - Transaction amount in original currency
- `currency` / `curr` - Currency code (USD, EUR, etc.)
- `amount_eur` / `sm_eur` - Amount converted to EUR for reporting
- `status` - Current status of the request (0-3)
- `description` - Human-readable description of the operation

---

## BROKER REPORT

### Money funds movement - `getUserCashFlows`

Obtaining data on the client's cash flow.

**Description:** A method for receiving detailed cash flow data including commissions, trade settlements, dividends, and other monetary movements. Provides advanced filtering, sorting, and grouping capabilities.

**HTTP Method:** GET
**Command:** `getUserCashFlows`

**Authorization:** Required. The SID received during authorization must be passed in the header of the SID cookie request, or in the request parameter.

#### Request Parameters:

```json
{
    "cmd": "getUserCashFlows",
    "SID": "[SID by authorization]",
    "params": {
        "user_id"        : null,    // int|null - Optional
        "groupByType"    : 1,       // int|null - Optional (1=group by type, 0=no grouping)
        "cash_totals"    : 1,       // int|null - Optional (1=show trade amounts by day, 0=hide)
        "hide_limits"    : 0,       // int|null - Optional (1=hide available limits, 0=show)
        "take"           : 10,      // int|null - Optional (pagination limit)
        "skip"           : 5,       // int|null - Optional (pagination offset)
        "without_refund" : 1,       // int|null - Optional (1=exclude refunds, 0=include)
        "filters"        : [        // array|null - Optional (filter conditions)
            {
                "field"    : "type_code",  // Filter field
                "operator" : "neq",        // Filter statement
                "value"    : "your value"  // Filter value
            }
        ],
        "sort"           : [        // array|null - Optional (sort conditions)
            {
                "field" : "type_code",     // Sorting field
                "dir"   : "DESC"           // Sorting order (ASC or DESC)
            }
        ]
    }
}
```

#### Parameter Details:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `cmd` | string | Yes | Request execution command ("getUserCashFlows") |
| `SID` | string | Yes | Session ID received during authorization |
| `params` | object | Yes | Request execution parameters |
| `params.user_id` | int\|null | Optional | Client ID to find the report |
| `params.groupByType` | int\|null | Optional | 1=Group by type, 0=No grouping |
| `params.cash_totals` | int\|null | Optional | 1=Show trade amounts by day, 0=Hide |
| `params.hide_limits` | int\|null | Optional | 1=Hide available limits, 0=Show limits |
| `params.take` | int\|null | Optional | Output data amount (pagination limit) |
| `params.skip` | int\|null | Optional | Output data offset (pagination skip) |
| `params.without_refund` | int\|null | Optional | 1=Data without refunds, 0=Include refunds |
| `params.filters` | array\|null | Optional | Output data filter (see filter structure below) |
| `params.sort` | array\|null | Optional | Output data sorting (see sort structure below) |

#### Filter Structure:

**Filterable Fields:**
- `date` - Transaction date
- `sum` - Transaction amount
- `currency` - Currency code
- `comment` - Transaction comment
- `type_code` - Type code (see "Types of valid codes" reference)

**Filter Operators:**
- `eq` - is equal to
- `neq` - not equal to
- `more` - the value is greater than the desired one
- `eqormore` - the value is equal to or greater than the desired one
- `eqorless` - the value is smaller or equal to the desired one
- `contains` - finding a value in the middle of the string
- `doesnotcontain` - values are missing in the middle of the string
- `startswith` - values are being searched for at the beginning of the string
- `endswith` - values are being searched for at the end of the string
- `in` - search for any of the transferred values

#### Sort Structure:

**Sortable Fields:** Same as filterable fields (date, sum, currency, comment, type_code)

**Sort Directions:**
- `ASC` - Ascending order
- `DESC` - Descending order

#### Response Structure (Success):

```json
{
    "total": 10,
    "cashflow": [
        {
            "id": "9f0a11cc61",
            "type_code": "commission_for_trades",
            "icon": "commission",
            "date": "2021-06-28",
            "sum": "-3.00",
            "comment": "(Trade 1515 2021-06-28 12:21:35)",
            "currency": "USD",
            "type_code_name": "Commission for trades"
        }
    ],
    "limits": {
        "USD": {
            "minimum": 50.0,
            "multiplicity": 1.0,
            "maximum": 100.0
        }
    },
    "cash_totals": {
        "currency": "USD",
        "list": [
            {
                "date": "2021-06-28",
                "sum": 500.62
            },
            {
                "date": "2021-06-29",
                "sum": 1080.97
            }
        ]
    }
}
```

#### Response (Error):

```json
// Common error
{
    "errMsg": "Bad json",
    "code": 2
}
```

#### Response Fields:

**Top Level:**
- `total` (int) - Total number of records
- `cashflow` (array) - Array of cash flow records
- `limits` (object) - Available limits by currency (optional, hidden if `hide_limits=1`)
- `cash_totals` (object) - Daily cash totals (optional, shown if `cash_totals=1`)

**Cash Flow Record:**
- `id` (string) - Unique identifier
- `type_code` (string) - Type code for this cash flow (e.g., "commission_for_trades", "dividend", etc.)
- `icon` (string) - Icon identifier for UI
- `date` (string|datetime) - Transaction date
- `sum` (string) - Transaction amount (negative for outflows, positive for inflows)
- `comment` (string) - Description with trade reference
- `currency` (string) - Currency code
- `type_code_name` (string) - Human-readable type name

**Limits Object (by currency):**
- `minimum` (float) - Minimum transaction amount
- `multiplicity` (float) - Transaction amount multiplicity
- `maximum` (float) - Maximum transaction amount

**Cash Totals:**
- `currency` (string) - Reporting currency
- `list` (array) - Daily totals
  - `date` (datetime) - Date
  - `sum` (float) - Total amount for that day

#### Example (jQuery):

```javascript
var exampleParams = {
    "cmd": "getUserCashFlows",
    "SID": "[SID by authorization]",
    "params": {
        "user_id": null,
        "groupByType": 1,
        "take": 10,
        "skip": 5,
        "filters": [
            {
                'field': 'type_code',
                'operator': 'neq',
                'value': 'commission_for_trades'
            }
        ],
        "sort": [
            {
                'field': 'type_code',
                'dir': 'DESC'
            }
        ]
    }
};

function getUserCashFlows(callback) {
    $.getJSON("https://tradernet.com/api/", {q: JSON.stringify(exampleParams)}, callback);
}

getUserCashFlows(function(json){
    console.log(json);
});
```

---

### Cancel the order - `delTradeOrder`

Delete/Cancel an order.

**Description:** A method that allows to cancel submitted orders.

**Prerequisites:** First, open the session using the method `getAuthInfo`

**HTTP Method:** POST (for API V2)
**Command:** `delTradeOrder`

#### Request Parameters:

```json
{
    "order_id": 2929292929  // int - Required
}
```

#### Parameter Details:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `order_id` | int | Yes | ID of the order that we want to cancel |

#### Response (Success):

```json
{
    "order_id": 2929292929
}
```

Returns the ID of the canceled order on success.

#### Response (Error):

```json
// Common error
{
    "errMsg": "Unsupported query method",
    "code": 2
}

// Method error
{
    "error": "Type of security not identified. Please contact Support.",
    "code": 0
}

// Insufficient rights error
{
    "code": 12,
    "errorMsg": "This order may only be cancelled by a Cancellation Order or through traders",
    "error": "This order may only be cancelled by a Cancellation Order or through traders"
}
```

#### Error Codes:

| Code | Type | Description |
|------|------|-------------|
| 2 | Common Error | Unsupported query method |
| 0 | Method Error | Type of security not identified |
| 12 | Permission Error | Order may only be cancelled by a Cancellation Order or through traders |

**Note:** Error code 12 indicates that you have insufficient rights to cancel this specific order. Some orders can only be cancelled through special procedures or by traders.

#### Response Fields:

| Field | Type | Description |
|-------|------|-------------|
| `order_id` | int | Order ID of the canceled order |

#### Example (jQuery):

```javascript
$(document).ready(function () {
    var settings = {
        url: 'https://tradernet.com/api/v2',
        post: 'POST',
        apiKey: '‹YOUR_API_KEY›',
        apiSecretKey: '‹YOUR_API_SECRET_KEY›',
        nonce: (new Date().getTime() * 10000),
        sign: '‹hash›'
    };

    delTradeOrder();

    function delTradeOrder() {
        getAuthInfo(function (responseText) {
            if (responseText.sess_id) {
                ajaxSender(
                    {
                        "cmd": "delTradeOrder",
                        "apiKey": settings.apiKey,
                        "nonce": settings.nonce,
                        "params": {
                            "order_id": 2929292929
                        }
                    },
                    settings.url + '/cmd/delTradeOrder',
                    function (responseText) {
                        console.log(responseText);
                    }
                );
            }
        });
    }

    function getAuthInfo(callback) {
        ajaxSender(
            {
                "cmd": "getAuthInfo",
                "apiKey": settings.apiKey,
                "nonce": settings.nonce,
            },
            settings.url + '/cmd/getAuthInfo',
            callback
        );
    }

    function ajaxSender(data, url, callback) {
        url = (typeof url === 'undefined') ? settings.url : url;

        $.ajaxSetup({
            headers: {
                'X_REQUESTED_WITH': 'XMLHttpRequest',
                'X-NtApi-Sig': settings.sign,
                'Nt-Jqp': true
            },
            xhrFields: {withCredentials: true}
        });

        $.ajax({
            url: url,
            method: settings.post,
            dataType: 'json',
            data: data,
            success: function (responseText) {
                if (callback) {
                    callback(responseText);
                } else {
                    console.log(responseText);
                }
            },
            error: function (err) {
                console.log(err)
            }
        });
    }
});
```

---

### Sending Stop Loss and Take Profit - `putStopLoss`

Sending Stop Loss and Take Profit commands for execution.

**Description:** A method that allows you to work with the submission of stop loss and take profit orders for open positions.

**Important Behavior:**
- If **all** stop loss parameters are null, the stop loss order does not change
- If `take_profit` is null, the take profit order does not change
- When setting trailing stop: specify `stop_loss_percent` and `stoploss_trailing_percent` (the `stop_loss` parameter is ignored)

**HTTP Method:** POST (for API V2)
**Command:** `putStopLoss`

#### Request Parameters:

```json
{
    "instr_name": "SIE.EU",                     // string - Required
    "take_profit": 1,                           // null|float - Optional
    "stop_loss": 1,                             // null|float - Optional
    "stop_loss_percent": 1,                     // null|float - Optional
    "stoploss_trailing_percent": 1              // null|float - Optional
}
```

#### Parameter Details:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `instr_name` | string | Yes | The instrument ticker (e.g., "AAPL.US") |
| `take_profit` | null\|float | Optional | Take profit price level. If null, take profit order does not change. |
| `stop_loss` | null\|float | Optional | Stop loss price level. If null, stop loss order does not change. **Ignored when using trailing stops.** |
| `stop_loss_percent` | null\|float | Optional | Stop loss percentage. Used for trailing stops. |
| `stoploss_trailing_percent` | null\|float | Optional | Trailing stop loss percentage. Used with `stop_loss_percent` for trailing stops. |

**Usage Patterns:**

1. **Simple Stop Loss:**
   - Set `stop_loss` to desired price
   - Leave other parameters null

2. **Simple Take Profit:**
   - Set `take_profit` to desired price
   - Leave other parameters null

3. **Trailing Stop:**
   - Set `stop_loss_percent` and `stoploss_trailing_percent`
   - The `stop_loss` parameter is ignored in this case

4. **Combined Stop Loss + Take Profit:**
   - Set both `stop_loss` and `take_profit`
   - Creates bracket order

#### Response Structure (Success):

```typescript
/**
 * Order Data Row
 */
type OrderDataRow = {
    date: string,                    // Order date
    market_time: string,             // Date of securities purchase
    changetime: number,              // Order change time (timestamp)
    last_checked_datetime: string,   // Last verification date
    exp: number,                     // Expiration: 1=Day, 2=Day+Ext, 3=GTC
    id: number,                      // Tradernet unique order ID
    order_id: number,                // Tradernet unique order ID (same as id)
    instr_type: number,              // Instrument type
    instr: string,                   // Ticker symbol
    leaves_qty: number,              // Number of remaining securities
    auth_login: string,              // Login of the client who sent the order
    creator_login: string,           // Login of the client who sent the order
    owner_login: string,             // Login of the user for which order was created
    user_id: string,                 // ID of the user who placed the order
    oper: number,                    // Action: 1=Buy, 2=Buy on Margin, 3=Sell, 4=Sell Short
    p: number,                       // Order price
    q: number,                       // Quantity in order
    curr_q: number,                  // Quantity before the transaction
    profit: number,                  // Trade profit
    cur: string,                     // Order currency
    stat: number,                    // Order status (see "Orders statuses")
    stat_d: string,                  // Order status modification date
    stat_orig: number,               // Initial order status (same as stat for API)
    stat_prev: number,               // Previous order status
    stop: number,                    // Order stop-price
    stop_activated: number,          // 1|0 indicator of activated stop order
    stop_init_price: number,         // Price to activate a stop order
    trailing_price: number | null,   // Trailing order variance percentage
    type: number,                    // Order type: 1=Market, 2=Limit, 3=Stop, 4=Stop Limit, 5=StopLoss, 6=TakeProfit
    user_order_id: string,           // Order ID assigned by user at order placing
    trades: OrderTradesInfo,         // Trade list for an order
    trades_json: string,             // Trade list for an order (JSON)
    error: string,                   // Ordering error (JSON)
    safety_type_id: string,          // Security session opening type
    order_nb: string | null          // Exchange ID
}

/**
 * Order Trades Info
 */
type OrderTradesInfo = {
    acd: number,      // Accumulated coupon interest
    date: string,     // Trade date
    fv: number,       // Coefficient for relative currencies (futures)
    go_sum: number,   // Initial margin per trade
    id: number,       // Tradernet unique trade ID
    p: number,        // Trade price
    profit: number,   // Trade profit
    q: number,        // Number of securities in trade
    v: number         // Trade amount
}

/**
 * Response
 */
type PutStopLossResponse = {
    order_id: number,
    order: OrderDataRow
}
```

#### Example Response:

```json
{
    "order_id": 192054709,
    "order": {
        "id": 192054709,
        "order_id": 192054709,
        "auth_login": "user@test.com",
        "user_id": 1088278273826,
        "date": "2020-10-06 12:41:58",
        "stat": 10,
        "stat_orig": 10,
        "stat_d": "2020-11-20 17:11:43",
        "instr": "SIE.EU",
        "oper": 3,
        "type": 6,
        "cur": "EUR",
        "p": "0.033925",
        "stop": "0.06",
        "stop_init_price": "0.036445",
        "stop_activated": 0,
        "q": "20000",
        "leaves_qty": "20000",
        "exp": 3,
        "stat_prev": 1,
        "user_order_id": "apiv2:160810000002",
        "trailing_price": null,
        "changetime": 160555333445000,
        "trades": "{}",
        "profit": "0.00",
        "curr_q": "20000",
        "trades_json": "[]",
        "error": "",
        "market_time": "2020-10-06 12:41:58",
        "owner_login": "user@test.com",
        "creator_login": "user@test.com",
        "safety_type_id": 3,
        "repo_start_date": null,
        "repo_end_date": null,
        "repo_start_cash": null,
        "repo_end_cash": null,
        "instr_type": 1,
        "order_nb": null,
        "last_checked_datetime": "2020-11-20 17:11:47"
    }
}
```

#### Response (Error):

```json
// Common error
{
    "errMsg": "Unsupported query method",
    "code": 2
}

// Method error
{
    "error": "You are trying to submit a request for client CLIENT, but you are working under client Real CLIENT",
    "code": 1
}
```

#### Response Fields:

| Field | Type | Description |
|-------|------|-------------|
| `order_id` | int | Order ID of the created order |
| `order` | object | Complete order data (see OrderDataRow type above) |

#### Order Type Values:

| Value | Type |
|-------|------|
| 1 | Market Order |
| 2 | Limit Order |
| 3 | Stop Order |
| 4 | Stop Limit Order |
| 5 | StopLoss |
| 6 | TakeProfit |

#### Example (jQuery):

```javascript
$(document).ready(function () {
    var settings = {
        url: 'https://tradernet.com/api/v2',
        post: 'POST',
        apiKey: '‹YOUR_API_KEY›',
        apiSecretKey: '‹YOUR_API_SECRET_KEY›',
        nonce: (new Date().getTime() * 10000),
        sign: '‹hash›'
    };

    putStopLoss();

    function putStopLoss() {
        getAuthInfo(function (responseText) {
            if (responseText.sess_id) {
                ajaxSender(
                    {
                        "cmd": "putStopLoss",
                        "apiKey": settings.apiKey,
                        "nonce": settings.nonce,
                        "params": {
                            "instr_name": "SIE.EU",
                            "take_profit": 1,
                            "stop_loss_percent": 1,
                            "stoploss_trailing_percent": 1
                        }
                    },
                    settings.url + '/cmd/putStopLoss',
                    function (responseText) {
                        console.log(responseText);
                    }
                );
            }
        });
    }

    function getAuthInfo(callback) {
        ajaxSender(
            {
                "cmd": "getAuthInfo",
                "apiKey": settings.apiKey,
                "nonce": settings.nonce,
            },
            settings.url + '/cmd/getAuthInfo',
            callback
        );
    }

    function ajaxSender(data, url, callback) {
        url = (typeof url === 'undefined') ? settings.url : url;

        $.ajaxSetup({
            headers: {
                'X_REQUESTED_WITH': 'XMLHttpRequest',
                'X-NtApi-Sig': settings.sign,
                'Nt-Jqp': true
            },
            xhrFields: {withCredentials: true}
        });

        $.ajax({
            url: url,
            method: settings.post,
            dataType: 'json',
            data: data,
            success: function (responseText) {
                if (callback) {
                    callback(responseText);
                } else {
                    console.log(responseText);
                }
            },
            error: function (err) {
                console.log(err)
            }
        });
    }
});
```

**References:**
- Order statuses available at "Orders statuses"
- Types of opening a security session at "Types of signatures"
- Instrument types and type name available at "Instruments details"

---

### Receive orders in current period - `getNotifyOrderJson`

Receive orders in the current period and subscribe for changes.

**Description:** A method that allows you to get a list of orders in the current period with a subscription to changes.

**HTTP Method:** POST (for API V2)
**Command:** `getNotifyOrderJson`

#### Request Parameters:

```json
{
    "cmd": "getNotifyOrderJson",
    "params": {
        "active_only": 1  // int - Optional (1=active only, 0=all orders)
    }
}
```

#### Parameter Details:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `cmd` | string | Yes | Request execution command ("getNotifyOrderJson") |
| `params` | object | Yes | Request execution parameters |
| `params.active_only` | int | Optional | 1 = Show only active orders, 0 = Show all orders (active + completed) |

#### Response Structure (Success):

```typescript
/**
 * Order Data Row
 */
type OrderDataRow = {
    aon: 0 | 1,                  // All or Nothing: 0=can be partially executed, 1=cannot be partially executed
    cur: string,                 // Order currency
    curr_q: number,              // Current quantity
    date: string,                // Order date (ISO format)
    exp: 1 | 2 | 3,             // Expiration: 1=Day, 2=Day+Ext, 3=GTC
    fv: number,                  // Coefficient for relative currencies (futures)
    order_id: number,            // Tradernet unique order ID
    instr: string,               // Ticker symbol
    leaves_qty: number,          // Remaining securities number
    auth_login: string,          // Login of the client who sent the order
    creator_login: string,       // Login of the client who sent the order
    owner_login: string,         // Login of the user for which order was created
    mkt_id: number,              // Market unique trade ID
    name: string,                // Name of company issuing security
    name2: string,               // Alternative name of the issuer
    oper: 1 | 2 | 3 | 4,        // Action: 1=Buy, 2=Buy on Margin, 3=Sell, 4=Sell Short
    p: number,                   // Order price
    q: number,                   // Quantity in order
    rep: number,                 // (Field purpose unclear)
    stat: number,                // Order status (see "Orders statuses")
    stat_d: string,              // Order status modification date
    stat_orig: number,           // Initial order status (equals stat for API)
    stat_prev: number,           // Previous order status
    stop: number,                // Order stop-price
    stop_activated: 0 | 1,       // Stop order activation indicator
    stop_init_price: number,     // Price to activate a stop order
    trailing_price: number,      // Trailing order variance percentage
    type: 1 | 2 | 3 | 4 | 5 | 6, // Order type: 1=Market, 2=Limit, 3=Stop, 4=Stop Limit, 5=StopLoss, 6=TakeProfit
    user_order_id: number,       // Order ID assigned by user at order placing
    trade: OrderTradeInfo[]      // Trade list for an order
}

/**
 * Order Trade Info
 */
type OrderTradeInfo = {
    acd: number,      // Accumulated coupon interest
    date: string,     // Trade date
    fv: number,       // Coefficient for relative currencies (futures)
    go_sum: number,   // Initial margin per trade
    id: number,       // Tradernet unique trade ID
    p: number,        // Trade price
    profit: number,   // Trade profit
    q: number,        // Number of securities in trade
    v: number         // Trade amount
}

/**
 * Response (array of orders)
 */
type GetNotifyOrderJsonResponse = OrderDataRow[]
```

#### Example Response:

```json
[
    {
        "aon": 0,
        "cur": "USD",
        "curr_q": 0,
        "date": "2015-12-23T17:05:02.133",
        "exp": 1,
        "fv": 0,
        "order_id": 8757875,
        "instr": "FCX.US",
        "leaves_qty": 0,
        "auth_login": "virtual@virtual.com",
        "creator_login": "virtual@virtual.com",
        "owner_login": "virtual@virtual.com",
        "mkt_id": 30000000001,
        "name": "Freeport-McMoran Cp & Gld",
        "name2": "Freeport-McMoran Cp & Gld",
        "oper": 2,
        "p": 6.5611,
        "q": 2625,
        "rep": 0,
        "stat": 21,
        "stat_d": "2015-12-23T17:05:03.283",
        "stat_orig": 21,
        "stat_prev": 10,
        "stop": 0,
        "stop_activated": 1,
        "stop_init_price": 6.36,
        "trailing_price": 0,
        "type": 1,
        "user_order_id": 1450879514204,
        "trade": [
            {
                "acd": 0,
                "date": "2015-12-23T17:05:03",
                "fv": 100,
                "go_sum": 0,
                "id": 13446624,
                "p": 6.37,
                "profit": 0,
                "q": 2625,
                "v": 16721.25
            }
        ]
    }
]
```

#### Response (Error):

```json
// Common error
{
    "errMsg": "Unsupported query method",
    "code": 2
}
```

#### Response Fields:

| Field | Type | Description |
|-------|------|-------------|
| (root) | array | Array of order objects |

**Key Order Fields:**
- `order_id` - Unique order identifier
- `instr` - Ticker symbol
- `type` - Order type (1-6)
- `oper` - Action type (1-4)
- `stat` - Order status
- `leaves_qty` - Remaining quantity
- `p` - Order price
- `q` - Original quantity
- `trade` - Array of executed trades for this order

#### WebSocket Support:

The server sends the `'orders'` event with order updates via WebSocket:

```javascript
const WS_SOCKETURL = 'wss://wss.tradernet.com/';
const ws = new WebSocket(WS_SOCKETURL);

ws.onopen = function () {
    // Subscribe to orders updates
    ws.send(JSON.stringify(['orders']));
};

ws.onmessage = function ({ data }) {
    // Server message handler
    const [event, messageData] = JSON.parse(data);
    if (event === 'orders') {
        console.info(messageData); // OrderDataRow[]
    }
};
```

**References:**
- Order statuses available at "Orders statuses"

---

### Get orders list for period - `getOrdersHistory`

Retrieving orders history for the period.

**Description:** Request to get the user's order history for the selected period.

**Authorization:** Required. The SID received during authorization must be passed in the header of the SID cookie request, or in the request parameter.

**HTTP Method:** GET
**Command:** `getOrdersHistory`

#### Request Parameters:

```json
{
    "cmd": "getOrdersHistory",
    "SID": "[SID by authorization]",
    "params": {
        "from": "2020-03-23T00:00:00",  // datetime - Required (ISO 8601 format)
        "till": "2020-04-03T23:59:59"   // datetime - Required (ISO 8601 format)
    }
}
```

#### Parameter Details:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `cmd` | string | Yes | Request execution command ("getOrdersHistory") |
| `SID` | string | Yes | Session ID received during authorization |
| `params` | object | Yes | Request execution parameters |
| `params.from` | datetime | Yes | Period start date, format ISO 8601: YYYY-MM-DD\Thh:mm:ss |
| `params.till` | datetime | Yes | Period end date, format ISO 8601: YYYY-MM-DD\Thh:mm:ss |

**Date Format:** ISO 8601 format `YYYY-MM-DD\Thh:mm:ss` (e.g., "2020-03-23T00:00:00")

#### Response Structure (Success):

```typescript
/**
 * Order Data Row (Historical)
 */
type HistoricalOrderDataRow = {
    id: number,                  // Order ID
    date: string | Date,         // Order date (ISO 8601)
    stat: number,                // Order status
    stat_orig: number,           // Initial order status (equals stat for API)
    stat_d: string | Date,       // Order status modification date
    instr: string,               // Ticker symbol
    oper: number,                // Action: 1=Buy, 2=Buy on Margin, 3=Sell, 4=Sell Short
    type: number,                // Order type: 1=Market, 2=Limit, 3=Stop, 4=Stop Limit, 5=StopLoss, 6=TakeProfit
    cur: string,                 // Currency
    p: number,                   // Order price
    stop: string,                // Stop price
    stop_init_price: number,     // Price to activate stop order
    stop_activated: number,      // Stop activation indicator
    q: number,                   // Quantity
    leaves_qty: number,          // Remaining quantity
    aon: number,                 // All or Nothing: 0=can be partially executed, 1=cannot
    exp: number,                 // Expiration: 1=Day, 2=Day+Ext, 3=GTC
    rep: string,                 // (Field purpose unclear)
    fv: string,                  // Coefficient for relative currencies
    name: string,                // Company name
    name2: string,               // Alternative company name
    stat_prev: number,           // Previous order status
    userOrderId: string,         // User-assigned order ID
    trailing: string,            // Trailing percentage
    login: string,               // User login
    instr_type: number,          // Instrument type
    curr_q: number,              // Current quantity
    mkt_id: number,              // Market unique trade ID
    owner_login: string,         // Owner login
    comp_login: string,          // Company login
    safety_type_id: number,      // Security session opening type
    condition: string,           // Order condition
    text: string,                // Order text/message
    "@text": string,             // Alternative text field
    OrigClOrdID: string | null,  // Original client order ID
    trade: TradeData[]           // Array of executed trades
}

/**
 * Trade Data
 */
type TradeData = {
    id: number,                  // Trade ID
    p: number,                   // Trade price
    q: number,                   // Trade quantity
    v: number,                   // Trade value
    date: string | Date,         // Trade date
    profit: string,              // Trade profit
    acd: string,                 // Accumulated coupon interest
    pay_d: string | Date,        // Payment date
    before_q: string,            // Quantity before trade
    after_q: number,             // Quantity after trade
    details: string              // Trade details
}

/**
 * Response
 */
type GetOrdersHistoryResponse = {
    orders: {
        key: string,                          // User key (login)
        order: HistoricalOrderDataRow[] | null  // Array of orders (null if none)
    }
}
```

#### Example Response:

```json
{
    "orders": {
        "key": "USER",
        "order": [
            {
                "id": 1111222222112,
                "date": "2020-03-23T10:00:28.853",
                "stat": 31,
                "stat_orig": 31,
                "stat_d": "2020-03-23T10:00:33.620",
                "instr": "AAPL.US",
                "oper": 1,
                "type": 2,
                "cur": "USD",
                "p": 100,
                "stop": ".00000000",
                "stop_init_price": 112.82,
                "stop_activated": 1,
                "q": 10,
                "leaves_qty": 10,
                "aon": "0",
                "exp": 3,
                "rep": "0",
                "fv": "0",
                "name": "Apple Inc.",
                "name2": "Apple Inc.",
                "stat_prev": 2,
                "userOrderId": "cps_21112",
                "trailing": ".00000000",
                "login": "USER",
                "instr_type": 1,
                "curr_q": 30,
                "mkt_id": 95006833,
                "owner_login": "USER",
                "comp_login": "example@domain.com",
                "safety_type_id": 15,
                "condition": "",
                "text": "The order may not be accepted",
                "@text": "The order may not be accepted",
                "OrigClOrdID": "The order may not be accepted",
                "trade": [
                    {
                        "id": 40543041,
                        "p": 78.5129,
                        "q": 1,
                        "v": 78.51,
                        "date": "2020-04-02T21:19:47",
                        "profit": ".00000000",
                        "acd": ".00000000",
                        "pay_d": "2020-04-06T00:00:00",
                        "before_q": ".00000000",
                        "after_q": 1,
                        "details": ""
                    }
                ]
            }
        ]
    }
}
```

#### Response (Error):

```json
// Common error
{
    "errMsg": "Bad json",
    "code": 2
}

// Method error
{
    "error": "Exec wrong",
    "code": 18
}
```

#### Response Fields:

| Field | Type | Description |
|-------|------|-------------|
| `orders` | object | Orders container |
| `orders.key` | string | User key (login) |
| `orders.order` | array\|null | Array of order objects, null if no orders found |

**Key Order Fields:**
- `id` - Unique order identifier
- `instr` - Ticker symbol
- `type` - Order type (1-6)
- `oper` - Action type (1-4)
- `stat` - Order status
- `p` - Order price
- `q` - Original quantity
- `leaves_qty` - Remaining quantity
- `trade` - Array of executed trades for this order

#### Example (jQuery):

```javascript
var exampleParams = {
    "cmd": "getOrdersHistory",
    "SID": "[SID by authorization]",
    "params": {
        "from": "2020-03-23T00:00:00",
        "till": "2020-04-03T23:59:59"
    }
};

function getOrdersHistory(callback) {
    $.getJSON("https://tradernet.com/api/", {q: JSON.stringify(exampleParams)}, callback);
}

getOrdersHistory(function(json){
    console.log(json);
});
```

---

## QUOTES AND TICKERS

### Retrieving trades history - `getTradesHistory`

Retrieving trades history.

**Description:** Request for user's trades history.

**Authorization:** Required. The SID received during authorization must be passed in the header of the SID cookie request, or in the request parameter.

**HTTP Method:** GET
**Command:** `getTradesHistory`

#### Request Parameters:

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

#### Parameter Details:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `cmd` | string | Yes | Request execution command ("getTradesHistory") |
| `SID` | string | Yes | Session ID received during authorization |
| `params` | object | Yes | Request execution parameters |
| `params.beginDate` | date\|string | Yes | Period start date, format ISO 8601: YYYY-MM-DD |
| `params.endDate` | date\|string | Yes | Period end date, format ISO 8601: YYYY-MM-DD |
| `params.tradeId` | int\|null | Optional | From which Trade ID to start retrieving report data |
| `params.max` | int\|null | Optional | Number of trades. If 0 or not specified - returns all trades |
| `params.nt_ticker` | string\|null | Optional | Instrument ticker (e.g., "AAPL.US") |
| `params.curr` | string\|null | Optional | Base currency or quote currency |
| `params.reception` | int\|null | Optional | Office ID |

**Date Format:** ISO 8601 date format `YYYY-MM-DD` (e.g., "2020-03-23")

#### Response Structure (Success):

```typescript
/**
 * Max Trade ID Row
 */
type MaxTradeIdRow = {
    "@text": string  // Last Trade ID
}

/**
 * Trade Row
 */
type TradeRow = {
    id: string,                           // ID in the Tradernet system
    order_id: string,                     // Exchange order number
    p: string,                            // Trade price
    q: string,                            // Quantity
    v: string,                            // Trade amount
    date: string | Date,                  // Date of compilation
    profit: string,                       // Trade profit
    instr_nm: string,                     // Security ticker
    curr_c: string,                       // Current currency
    type: string,                         // Trade type: 1=Buy, 2=Sell
    reception: string,                    // Office
    login: string,                        // Client login
    summ: string,                         // Amount
    curr_q: string,                       // Quantity before the transaction
    instr_type_c: string,                 // Security type
    mkt_id: string,                       // Market
    instr_id: string,                     // Instrument ID
    comment: string,                      // Comment
    step_price: string,                   // Price increment
    min_step: string,                     // Minimum price increment
    rate_offer: string,                   // Cross rate - portfolio currency to security currency
    fv: string,                           // Coefficient for relative currencies (futures)
    acd: string,                          // Accumulated coupon interest (ACI)
    go_sum: string,                       // Initial margin per trade
    curr_price: string,                   // Currency price
    curr_price_money: string,             // Currency price
    curr_price_begin: string,             // Starting price in currency
    curr_price_begin_money: string,       // Starting price in currency
    pay_d: string,                        // Trade date
    trade_d_exch: string | Date | null,  // Trade time on the exchange
    T2_confirm: string,                   // Confirmation of settlement for a trade
    trade_nb: string,                     // Trade exchange number
    repo_close: string,                   // Repo closing
    StartCash: string,                    // The amount of repo opening
    EndCash: string,                      // Closing price of REPO
    commiss_exchange: string,             // Exchange commission for derivatives market
    otc: string,                          // OTC trade sign
    details: string,                      // Transaction details for TN
    OrigClOrdID: string | null            // Original client order ID
}

/**
 * Response
 */
type GetTradesHistoryResponse = {
    trades: {
        max_trade_id: MaxTradeIdRow[],  // Array with last trade ID
        trade: TradeRow[]                // Array of trade objects
    }
}
```

#### Example Response:

```json
{
    "trades": {
        "max_trade_id": [
            {
                "@text": "40975888"
            }
        ],
        "trade": [
            {
                "id": 2229992229292,
                "order_id": 299998887727,
                "p": 141.4,
                "q": 20,
                "v": 2828,
                "date": "2019-08-15T10:10:22",
                "profit": ".00000000",
                "instr_nm": "AAPL.US",
                "curr_c": "USD",
                "type": 1,
                "reception": 1,
                "login": "example@domain.com",
                "summ": 2828,
                "curr_q": ".000000000000",
                "instr_type_c": 1,
                "mkt_id": 95006833,
                "instr_id": 10000005775,
                "comment": "56896/",
                "step_price": ".02000000",
                "min_step": ".02000000",
                "rate_offer": 1,
                "fv": 100,
                "acd": ".0000000000",
                "go_sum": ".000000000000",
                "curr_price": ".000000000000",
                "curr_price_money": ".000000000000",
                "curr_price_begin": ".000000000000",
                "curr_price_begin_money": ".000000000000",
                "pay_d": "2019-08-19T00:00:00",
                "T2_confirm": "2019-08-19T00:00:00",
                "trade_nb": 299998887727,
                "repo_close": "0",
                "StartCash": ".000000",
                "EndCash": ".000000",
                "commiss_exchange": ".00000000",
                "otc": "0",
                "details": ""
            }
        ]
    }
}
```

#### Response (Error):

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

#### Response Fields:

| Field | Type | Description |
|-------|------|-------------|
| `trades` | object | Trades container |
| `trades.max_trade_id` | array | Array with last trade ID |
| `trades.trade` | array | Array of trade objects |

**Key Trade Fields:**
- `id` - Tradernet unique trade ID
- `order_id` - Exchange order number
- `instr_nm` - Security ticker
- `type` - Trade type (1=Buy, 2=Sell)
- `p` - Trade price
- `q` - Quantity
- `v` - Trade amount (value)
- `date` - Trade date/time
- `curr_c` - Currency

#### Example (jQuery):

```javascript
var exampleParams = {
    "cmd": "getTradesHistory",
    "SID": "[SID by authorization]",
    "params": {
        "beginDate": "2020-03-23",
        "endDate": "2020-04-08",
        "tradeId": 232327727,
        "max": 100,
        "nt_ticker": "AAPL.US",
        "curr": "USD",
        "reception": 1
    }
};

function getTradesHistory(callback) {
    $.getJSON("https://tradernet.com/api/", {q: JSON.stringify(exampleParams)}, callback);
}

getTradesHistory(function(json){
    console.log(json);
});
```

---

### Get stock ticker data - `getStockQuotesJson`

**Status:** ⏳ AWAITING DOCUMENTATION

---

### Get quote historical data - `getHloc`

Get historical candlestick (OHLC) data for securities.

**Description:** Method of obtaining historical information as per the listing (candlesticks).

**HTTP Method:** GET or POST (for API V2)
**Command:** `getHloc`

#### Request Parameters:

```json
{
    "cmd": "getHloc",
    "params": {
        "userId": null,                      // int|null - Optional
        "id": "FB.US",                       // string - Required
        "count": -1,                         // signed int - Required
        "timeframe": 1440,                   // int - Required
        "date_from": "15.08.2020 00:00",     // string|datetime - Required
        "date_to": "16.08.2020 00:00",       // string|datetime - Required
        "intervalMode": "ClosedRay"          // string - Required
    }
}
```

#### Parameter Details:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `cmd` | string | Yes | Request execution command ("getHloc") |
| `params` | object | Yes | Request execution parameters |
| `params.userId` | int\|null | Optional | User ID. Used to retrieve candlestick history data for a registered user. For GET request only (API v1) |
| `params.id` | string | Yes | Ticker name. You can specify multiple tickers separated by commas (e.g., "AAPL.US,MSFT.US") |
| `params.count` | signed int | Yes | Number of candlesticks to receive in addition to specified interval. Use `-1` if not required. For example: `-100` = get all candlesticks for the year + 100 candlesticks before the interval |
| `params.timeframe` | int | Yes | Interval in minutes. Valid values: `1`, `5`, `15`, `60`, `1440` |
| `params.date_from` | string\|datetime | Yes | Start date of the interval. Format: `DD.MM.YYYY hh:mm` (e.g., "15.08.2020 00:00") |
| `params.date_to` | string\|datetime | Yes | End date of the interval. Format: `DD.MM.YYYY hh:mm` (e.g., "16.08.2020 00:00") |
| `params.intervalMode` | string | Yes | **Required parameter, single value: "ClosedRay"** |

**Date Format:** `DD.MM.YYYY hh:mm` (e.g., "15.08.2020 00:00")

**Timeframe Values:**
- `1` - 1 minute
- `5` - 5 minutes
- `15` - 15 minutes
- `60` - 1 hour
- `1440` - 1 day (24 hours)

**CRITICAL:** `intervalMode` must be `"ClosedRay"` (not "OpenRay")

#### Response Structure (Success):

```typescript
/**
 * Candlestick Data Response
 */
type HlocDataResponse = {
    hloc: {
        [ticker: string]: number[][]  // [high, low, open, close] for each candlestick
    },
    vl: {
        [ticker: string]: number[]    // Volume for each candlestick
    },
    xSeries: {
        [ticker: string]: number[]    // Timestamps in seconds (Unix epoch)
    },
    maxSeries: number,                // Timestamp of the most recent candlestick
    info: {
        [ticker: string]: TickerInfo  // Information about the requested ticker
    },
    took: number                      // Request execution time in milliseconds
}

/**
 * Ticker Information
 */
type TickerInfo = {
    id: string,              // Ticker ID
    nt_ticker: string,       // Tradernet ticker name
    short_name: string,      // Company short name
    default_ticker: string,  // Default ticker symbol
    code_nm: string,         // Ticker code
    currency: string,        // Trading currency
    min_step: string,        // Minimum price increment
    lot: string             // Lot size
}
```

#### Example Response:

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

#### Response (Error):

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

#### Response Fields:

| Field | Type | Description |
|-------|------|-------------|
| `hloc` | Object | Candlestick OHLC data. Key: ticker name, Value: array of `[high, low, open, close]` arrays |
| `vl` | Object | Candlestick volume data. Key: ticker name, Value: array of volume numbers |
| `xSeries` | Object | Candlestick timestamps **in seconds** (Unix epoch). Key: ticker name, Value: array of timestamps |
| `maxSeries` | number | Timestamp of the most recent candlestick rendering |
| `info` | Object | Information about the requested ticker(s) |
| `took` | float | Request execution time in milliseconds |

**Important Notes:**
- **Timestamps are in seconds** (not milliseconds)
- HLOC format is: `[high, low, open, close]` (not OHLC)
- Multiple tickers can be requested by separating with commas in the `id` parameter
- Response will contain data for all requested tickers
- The `count` parameter allows fetching additional candlesticks before the specified date range

#### Example Usage (jQuery):

```javascript
/**
 * @type {GetHlocParams}
 */
var exampleParams = {
    "cmd": "getHloc",
    "params": {
        "id": "FB.US",
        "count": -1,
        "timeframe": 1440,
        "date_from": "15.08.2020 00:00",
        "date_to": "16.08.2020 00:00",
        "intervalMode": "ClosedRay"
    }
};

function getHloc(callback) {
    $.getJSON("https://tradernet.com/api/", {q: JSON.stringify(exampleParams)}, callback);
}

/**
 * Get the object
 **/
getHloc(function(json){
    console.log(json);
});
```

---

### Stock ticker search - `tickerFinder`

Search for securities by ticker symbol or company name.

**Description:** Search endpoint that returns a maximum of 30 securities matching the search query.

**HTTP Method:** GET
**Command:** `tickerFinder`

#### Request Parameters:

```json
{
    "cmd": "tickerFinder",
    "params": {
        "text": "AAPL.US"  // string - Required
    }
}
```

**Search Formats:**

1. **Simple search:** `"AAPL"` or `"Apple"` - searches across all markets
2. **Exchange-specific search:** `"<ticker>@<market>"` - searches on specific exchange

**Example:** `"AAPL@FIX"` - search for AAPL on NYSE/NASDAQ

#### Parameter Details:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `cmd` | string | Yes | Request execution command ("tickerFinder") |
| `params` | object | Yes | Request execution parameters |
| `params.text` | string | Yes | Search query (ticker symbol or company name). Use lowercase for best results. |

#### Market Codes (for Exchange-Specific Search):

| Market Code | Exchange |
|-------------|----------|
| `MCX` | MICEX (Moscow Exchange) |
| `FORTS` | MICEX Derivatives |
| `FIX` | NYSE/NASDAQ (US markets) |
| `UFORTS` | Ukrainian Derivatives Exchange |
| `UFOUND` | Ukrainian Exchange |
| `EU` | Europe |
| `KASE` | Kazakhstan |

**Format:** `"<ticker>@<market>"` (e.g., `"AAPL@FIX"`)

#### Response Structure (Success):

```typescript
/**
 * Ticker Finder Data Row
 */
type TickerFinderDataRow = {
    instr_id: number,     // Unique ticker ID
    nm: string,           // Name (too long format)
    n: string,            // Name (short format)
    ln: string,           // English name
    t: string,            // Ticker in Tradernet's system
    isin: string,         // Ticker ISIN code
    type: number,         // Instrument type
    kind: number,         // Instrument sub-type
    tn: string,           // Ticker plus name
    code_nm: string,      // Exchange ticker
    mkt_id: number,       // Market code (numeric)
    mkt: string          // Market name
}

/**
 * Ticker Finder Result
 */
type TickerFinderResult = {
    found: TickerFinderDataRow[]  // List of found tickers (max 30)
}
```

#### Instrument Type and Kind Combinations:

| Type | Kind | Description |
|------|------|-------------|
| 1 | 1 | Regular stock (common stock) |
| 1 | 2 | Preferred stock |
| 1 | 7 | Investment units (funds/ETFs) |
| 2 | - | Bonds (all kinds) |
| 3 | - | Futures (all kinds) |
| 5 | - | Exchange index |
| 6 | 1 | Cash (fiat currency) |
| 6 | 8 | Crypto (cryptocurrency) |
| 8, 9, 10 | - | Repo (repurchase agreements) |

#### Example Response:

```json
{
    "found": [
        {
            "instr_id": 10000007229,
            "nm": "Apple Inc.",
            "n": "Apple",
            "ln": "Apple Inc.",
            "t": "AAPL.US",
            "isin": "US0378331005",
            "type": 1,
            "kind": 1,
            "tn": "AAPL.US Apple Inc.",
            "code_nm": "AAPL",
            "mkt_id": 95006833,
            "mkt": "NASDAQ"
        }
    ]
}
```

#### Response Fields:

| Field | Type | Description |
|-------|------|-------------|
| `found` | array | Array of matching securities (maximum 30 results) |
| `found[].instr_id` | number | Unique instrument ID in Tradernet system |
| `found[].nm` | string | Full company/instrument name (long format) |
| `found[].n` | string | Short company/instrument name |
| `found[].ln` | string | English language name |
| `found[].t` | string | Ticker symbol in Tradernet format (e.g., "AAPL.US") |
| `found[].isin` | string | International Securities Identification Number |
| `found[].type` | number | Instrument type code (see type/kind combinations table) |
| `found[].kind` | number | Instrument sub-type/kind code |
| `found[].tn` | string | Combined ticker and name string |
| `found[].code_nm` | string | Ticker symbol on the exchange (without suffix) |
| `found[].mkt_id` | number | Numeric market identifier |
| `found[].mkt` | string | Market/exchange name |

**Important Notes:**
- **Maximum 30 results** returned per search
- Use lowercase search text for best results
- Exchange-specific searches use `@` symbol: `"ticker@market"`
- Response may vary based on whether field names are short or full format
- The `type` and `kind` fields classify the instrument (see combinations table)

#### Example Usage (jQuery):

```javascript
/**
 * @typedef {{
 *  search: string,
 *  q: {
 *      cmd: 'tickerFinder',
 *      params: {
 *          text: string
 *      }
 *  }
 * }} TickerFinderQueryParams
 */

/**
 * @param {string} phrase
 * @param {function} callback
 */
function findTickers(phrase, callback) {
    /**
     * @type {TickerFinderQueryParams}
     */
    var queryParams = {
        q: {
            cmd: 'tickerFinder',
            params: {
                text: phrase.toLowerCase()
            }
        }
    };

    $.getJSON('https://tradernet.com/api', queryParams, callback);
}

// Simple search
findTickers('AAPL.US',
    /**
     * @param {TickerFinderResults} data
     */
    function (data) {
        console.info(data);
    }
);

// Exchange-specific search
findTickers('AAPL@FIX',
    function (data) {
        console.info(data);  // Search AAPL on NYSE/NASDAQ only
    }
);
```

#### Alternative Request Format:

The endpoint also supports a simpler GET request format:

```javascript
// Using q parameter directly
$.getJSON('https://tradernet.com/api', {
    q: JSON.stringify({
        cmd: 'tickerFinder',
        params: {
            text: 'AAPL@FIX'
        }
    })
}, callback);
```

**Best Practices:**
- Convert search text to lowercase for consistent results
- Use exchange-specific search (`@market`) to narrow results and improve relevance
- Check `type` and `kind` fields to filter by instrument category
- Validate ISIN if you need globally unique identification

---

### Get updates on market status - `getMarketStatus`

Obtain information about market statuses and operation.

**Description:** Get current status (open/closed), opening times, and time zones for markets.

**HTTP Method:** GET
**Command:** `getMarketStatus`

#### Request Parameters:

```json
{
    "cmd": "getMarketStatus",
    "params": {
        "market": "*",       // string - Required
        "mode": "demo"       // string|null - Optional
    }
}
```

#### Parameter Details:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `cmd` | string | Yes | Request execution command ("getMarketStatus") |
| `params` | object | Yes | Request execution parameters |
| `params.market` | string | Yes | Market identifier (briefName). Use `"*"` for all markets. See Market List table below. |
| `params.mode` | string\|null | Optional | Request mode: `"demo"`. If not specified, displays market statuses for real users. |

#### Market List (briefName Codes)

Complete list of available markets:

| Market Code | Full Title | Abbreviated Name |
|-------------|-----------|------------------|
| `*` | All markets | - |
| `AMX` | Armenia Securities Exchange | AMX |
| `AIX` | Astana International Exchange | AIX |
| `ATHEX` | Athens Stock Exchange | ATHEX |
| `BEB.RUS` | BEB. The market for the conversions calendar | BEB.RUS |
| `BEX` | BEX Best Execution | BEX |
| `SBQ` | Broker Quote System (BQS) | SBQ |
| `US_OPT` | CBOE (US Options) | US_OPT |
| `CMX` | COMEX (Commodity Exchange) | CMX |
| `CBF` | Cboe Futures Exchange | CBF |
| `CBT` | Chicago Board of Trade | CBT |
| `CME` | Chicago Mercantile Exchange | CME |
| `SecForCrypto` | Crypto | SecForCrypto |
| `FINERY` | Crypto Finery Market | FINERY |
| `CRPT` | Cryptocurrency market | CRPT |
| `EU` | EU Europe | EU |
| `EXANTE` | EXANTE | EXANTE |
| `EASTE` | East Exchange | EASTE |
| `EUX` | Eurex | EUX |
| `EUROBOND` | Eurobonds | EUROBOND |
| `FORTS` | FORTS Market FORTS | FORTS |
| `FFSP` | Freedom Finance Structural Products | FFSP |
| `HKG` | Hong Kong Futures Exchange | HKG |
| `HKEX` | Hong Kong Stock Exchange | HKEX |
| `EDX` | ICE Endex | EDX |
| `NYB` | ICE Futures U.S. | NYB |
| `WCE` | ICE Futures US-Canadian Grains | WCE |
| `IMEX` | IMEX Crypto Market | IMEX |
| `ISF` | ISF: ICE Futures Europe S2F | ISF |
| `ITS` | ITS | ITS |
| `ITS_MONEY` | ITS Money Market | ITS_MONEY |
| `ICE` | Intercontinental Exchange | ICE |
| `KASE` | Kazakhstan Stock Exchange | KASE |
| `KASE.CUR` | Kazakhstan Stock Exchange. Currency section | KASE.CUR |
| `Kraken` | Kraken Crypto Exchange | Kraken |
| `LME` | LME: London Metal Exchange | LME |
| `LMAX` | Lmax currency | LMAX |
| `MCX.CUR` | MCX Currency. Currency exchange | MCX.CUR |
| `MCX.OTC` | MCX Over-The-Counter Market | MCX.OTC |
| `MCX.nottraded` | MCX.nottraded | MCX.nottraded |
| `MOEX` / `MCX` | MICEX. Stock market | MCX |
| `MONEY` | MONEY Foreign Exchange Market | MONEY |
| `OTC.xxxx.RUR` | Market for settlement of forwards on foreign stocks for Russian Rubles | OTC.xxxx.RUR |
| `MBANK_EU` | MayBank EU Instruments | MBANK_EU |
| `MBANK` | MayBank HKE Instruments | MBANK |
| `MBANK_US` | MayBank US Instruments | MBANK_US |
| `MGE` | Minneapolis Grain Exchange (MGEX) | MGE |
| `NGC` | NSE IFSC | NGC |
| `NYF` | NYF - ICE Futures US Indices | NYF |
| `FIX` | NYSE/NASDAQ | FIX |
| `NSE` | Natl Stock Exchange of India | NSE |
| `NYM` | New York Mercantile Exchange | NYM |
| `FIX.OTC` | OTC. Foreign securities. | FIX.OTC |
| `PFTS_OBL` | PFTS. Obligations | PFTS_OBL |
| `PFTS_SPOT` | PFTS. Spot | PFTS_SPOT |
| `PRSP_OBL` | Perspektiva market. Obligations | PRSP_OBL |
| `PRSP_SPOT` | Perspektiva market. Spot | PRSP_SPOT |
| `RTSBoard` | RTSBoard РТС board | RTSBoard |
| `UZSE` | Republican Stock Exchange "Toshkent" (UZSE) | UZSE |
| `SGC` / `RTS` | SGQ system of guaranteed quotes on RTS | RTS |
| `SGX` | SGX: Singapore Exchange | SGX |
| `SPBFOR` | SPB Foreign securities. | SPBFOR |
| `SPBEX` | SPB. Russian securities. | SPBEX |
| `KASE.OTC` | Store. Kazakhstan. F24 | KASE.OTC |
| `TABADUL` | Tabadul Exchange | TABADUL |
| `UB_OBL` | UB. Obligations | UB_OBL |
| `UKR_FORTS` / `UFORTS` | UKR_FORTS FORTS Ukraine | UFORTS |
| `UKR_FOUND` / `UFOUND` | UKR_FOUND Stock Ukraine | UFOUND |
| `UX.OTC` | UX Over-The-Counter Market | UX.OTC |

#### Response Structure (Success):

```typescript
/**
 * Market Info Row
 */
type MarketInfoRow = {
    n: string,   // Full market name
    n2: string,  // Market abbreviation
    s: string,   // Current market status (e.g., "OPEN", "CLOSE")
    o: string,   // Market opening time (MSK timezone) - Format: "HH:MM:SS"
    c: string,   // Market closing time (MSK timezone) - Format: "HH:MM:SS"
    dt: string  // Time difference relative to MSK time (in minutes as string)
}

/**
 * Market Status Response
 */
type MarketStatusResponse = {
    result: {
        markets: {
            t: string,              // Current request time (server time)
            m: MarketInfoRow[]     // Array of market info
        }
    }
}
```

#### Example Response:

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

#### Response (Error):

```json
// Common error
{
    "errMsg": "Bad json",
    "code": 2
}

// Method error (service unavailable)
{
    "error": "Something wrong, service unavailable",
    "code": 14
}
```

**Error Codes:**
- **Code 2:** Common error (bad JSON)
- **Code 14:** Service unavailable

#### Response Fields:

| Field | Type | Description |
|-------|------|-------------|
| `result.markets` | object | Markets container |
| `result.markets.t` | string | Current request time (server timestamp) |
| `result.markets.m` | array | Array of market information objects |
| `result.markets.m[].n` | string | Full market name |
| `result.markets.m[].n2` | string | Market abbreviation/code |
| `result.markets.m[].s` | string | Current market status (e.g., "OPEN", "CLOSE", "PRE_OPEN", "POST_CLOSE") |
| `result.markets.m[].o` | string | Market opening time in MSK timezone (HH:MM:SS format) |
| `result.markets.m[].c` | string | Market closing time in MSK timezone (HH:MM:SS format) |
| `result.markets.m[].dt` | string | Time difference relative to Moscow time in minutes (e.g., "-180" = 3 hours behind) |

**Market Status Values:**
- `"OPEN"` - Market is currently open for trading
- `"CLOSE"` - Market is currently closed
- `"PRE_OPEN"` - Pre-market session
- `"POST_CLOSE"` - After-hours session

**Time Zone Notes:**
- Opening and closing times (`o`, `c`) are in **MSK (Moscow Standard Time)**
- Use `dt` field to calculate local market time
- Positive `dt` = ahead of MSK, Negative `dt` = behind MSK

#### Example Usage (jQuery):

```javascript
/**
 * @type {getMarketStatus}
 */
var paramsToGetStatus = {
    "cmd": "getMarketStatus",
    "params": {
        "market": "*",    // Get all markets
        "mode": "demo"    // Demo mode
    }
};

/**
 * The request allows you to get updates on the market status directly from the server
 */
function getMarketStatuses(callback) {
    $.getJSON("https://tradernet.com/api/", {q: JSON.stringify(paramsToGetStatus)}, callback);
}

getMarketStatuses(function (json) {
    console.info(json);
});

// Get specific market
var paramsSpecific = {
    "cmd": "getMarketStatus",
    "params": {
        "market": "FIX"  // Get NYSE/NASDAQ status only
    }
};
```

#### WebSocket Support:

The market status can also be received via WebSocket for real-time updates. Subscribe to market status changes to receive push notifications when markets open/close.

**Usage Notes:**
- Use `"*"` to get all markets at once
- Use specific market code (e.g., `"FIX"`) to get single market
- `mode: "demo"` shows demo account market statuses
- Omit `mode` parameter for real account market statuses
- Times are always in MSK timezone - convert using `dt` field
- Status updates are useful for determining if orders can be placed

---

### News on securities - `getNews`

**Status:** ⏳ AWAITING DOCUMENTATION

---

### Getting most traded securities - `getTopSecurities`

**Status:** ⏳ AWAITING DOCUMENTATION

---

### Options demonstration - `getOptionsByMktNameAndBaseAsset`

**Status:** ⏳ AWAITING DOCUMENTATION

---

## PRICE ALERTS

### Get current price alerts - `getAlertsList`

**Status:** ⏳ AWAITING DOCUMENTATION

---

### Add price alert - `addPriceAlert`

**Status:** ⏳ AWAITING DOCUMENTATION

---

### Delete price alert

**Status:** ⏳ AWAITING DOCUMENTATION

---

## BROKER REPORT

### Receiving broker report - `getBrokerReport`

**Status:** ⏳ AWAITING DOCUMENTATION

---

### Receiving order files - `getCpsFiles`

**Status:** ⏳ AWAITING DOCUMENTATION

---

### Getting broker report via direct link

**Status:** ⏳ AWAITING DOCUMENTATION

---

### Obtain depository report

**Status:** ⏳ AWAITING DOCUMENTATION

---

### Obtain depository report via direct link

**Status:** ⏳ AWAITING DOCUMENTATION

---

## CURRENCIES

### Exchange rate by date - `getCrossRatesForDate`

Exchange rate by date.

**Description:** Get exchange rates for specific currencies relative to a base currency on a specific date.

**HTTP Method:** GET
**Command:** `getCrossRatesForDate`

**Authorization:** Required. The SID received during authorization must be passed in the header of the SID cookie request, or in the request parameter.

#### Request Parameters:

```json
{
    "cmd": "getCrossRatesForDate",
    "SID": "[SID by authorization]",
    "params": {
        "base_currency": "USD",         // string - Required
        "currencies": ["EUR", "HKD"],   // array - Required
        "date": "2024-05-01"            // null|string - Optional
    }
}
```

#### Parameter Details:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `cmd` | string | Yes | Request execution command ("getCrossRatesForDate") |
| `SID` | string | Yes | Session ID received during authorization |
| `params` | object | Yes | Request execution parameters |
| `params.base_currency` | string | Yes | Base currency (e.g., "USD") |
| `params.currencies` | array | Yes | List of currencies for which the rate is retrieved (e.g., ["EUR", "HKD"]) |
| `params.date` | string\|null | Optional | Date as of which the rate is requested (YYYY-MM-DD format). If missing, current date is used. |

**Date Format:** ISO 8601 date format `YYYY-MM-DD` (e.g., "2024-05-01")

#### Response Structure (Success):

```json
{
    "rates": {
        "EUR": 0.92261342533093,
        "HKD": 7.8070160113905
    }
}
```

**Response Fields:**

| Field | Type | Description |
|-------|------|-------------|
| `rates` | object | Object containing exchange rates to the base currency |
| `rates.[CURRENCY]` | float | Exchange rate for the specified currency relative to base currency |

#### Response (Error):

```json
// Common error
{
    "errMsg": "Bad parameters",
    "code": 2
}
```

#### Example Usage (jQuery):

```javascript
/**
 * @type {getCrossRatesForDate}
 */
var exampleParams = {
    "cmd": "getCrossRatesForDate",
    "params": {
        "base_currency": "USD",
        "currencies": ["EUR", "HKD"],
        "date": "2024-05-01"
    }
};

function getCrossRatesForDate(callback) {
    $.getJSON("https://tradernet.com/api/", {q: JSON.stringify(exampleParams)}, callback);
}

/**
 * Get the object
 **/
getCrossRatesForDate(function(json){
    console.log(json);
});
```

---

### List of currencies

**Status:** ⏳ AWAITING DOCUMENTATION

---

## AUTHENTICATION & SESSION

### Login/Password

**Status:** ⏳ AWAITING DOCUMENTATION

---

### API key

**Status:** ⏳ AWAITING DOCUMENTATION

---

### Initial user data - `getOPQ`

**Status:** ⏳ AWAITING DOCUMENTATION

---

### User's current authorization session information

**Status:** ⏳ AWAITING DOCUMENTATION

---

### Public API client

**Status:** ⏳ AWAITING DOCUMENTATION

---

### Tradernet Python SDK

**Status:** ⏳ AWAITING DOCUMENTATION

---

## SECURITY SESSION

### Get list of open security sessions

**Status:** ⏳ AWAITING DOCUMENTATION

---

### Opening the security session

**Status:** ⏳ AWAITING DOCUMENTATION

---

## SET UP THE LIST OF SECURITIES

### Receiving lists of securities

**Status:** ⏳ AWAITING DOCUMENTATION

---

### Adding list of securities

**Status:** ⏳ AWAITING DOCUMENTATION

---

### Changing list of securities

**Status:** ⏳ AWAITING DOCUMENTATION

---

### Deleting saved list of securities

**Status:** ⏳ AWAITING DOCUMENTATION

---

### Setting selected list of securities

**Status:** ⏳ AWAITING DOCUMENTATION

---

### Adding ticker to list - `addStockListTicker`

Add a ticker to a user's securities list (watchlist).

**Description:** Add a specific ticker symbol to one of the user's custom securities lists.

**HTTP Method:** POST
**Command:** `addStockListTicker`

**Authorization:** Required. The SID received during authorization must be passed.

#### Request Parameters:

```json
{
    "cmd": "addStockListTicker",
    "SID": "<SID>",
    "params": {
        "id": 2,              // integer - Required
        "ticker": "AAPL.US",  // string - Required
        "index": 2            // integer - Required
    }
}
```

#### Parameter Details:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `cmd` | string | Yes | Request execution command ("addStockListTicker") |
| `SID` | string | Yes | Session ID received during authorization |
| `params` | object | Yes | Request execution parameters |
| `params.id` | integer | Yes | List ID (the ID of the securities list to add the ticker to) |
| `params.ticker` | string | Yes | Ticker symbol (e.g., "AAPL.US") |
| `params.index` | integer | Yes | Ticker item number (position in the list) |

#### Response Structure (Success):

```typescript
/**
 * Stock List
 */
type StockList = {
    id: number,           // List ID
    userId: number,       // User ID who owns the list
    name: string,         // List name
    tickers: string[],    // Array of ticker symbols in the list
    picture: string | null // Emoji or image for the list (optional)
}

/**
 * Add Stock List Ticker Response
 */
type AddStockListTickerResponse = {
    userStockLists: StockList[],  // All user's stock lists (updated)
    selectedId: number,            // Currently selected list ID
    defaultId: number             // Default list ID
}
```

#### Example Response:

```json
{
    "userStockLists": [
        {
            "id": 1,
            "userId": 123456,
            "name": "default",
            "tickers": [],
            "picture": null
        },
        {
            "id": 2,
            "userId": 123456,
            "name": "etf",
            "tickers": [
                "AAAU.US",
                "ACES.US",
                "AAPL.US",
                "ACIO.US",
                "AFIF.US"
            ],
            "picture": "🙂"
        }
    ],
    "selectedId": 1,
    "defaultId": 1
}
```

#### Response (Error):

```json
// Common error
{
    "errMsg": "Bad json",
    "code": 2
}
```

#### Response Fields:

| Field | Type | Description |
|-------|------|-------------|
| `userStockLists` | array | Complete array of all user's securities lists (updated after adding ticker) |
| `userStockLists[].id` | number | Unique list identifier |
| `userStockLists[].userId` | number | ID of the user who owns the list |
| `userStockLists[].name` | string | Display name of the list |
| `userStockLists[].tickers` | array | Array of ticker symbols in this list (in order) |
| `userStockLists[].picture` | string\|null | Emoji or image representing the list (optional) |
| `selectedId` | number | ID of the currently selected/active list |
| `defaultId` | number | ID of the user's default list |

**Important Notes:**
- Response returns **all** user's lists, not just the modified one
- The `tickers` array shows the updated list including the newly added ticker
- `index` parameter determines the position of the ticker in the list
- If ticker already exists in the list, it may be moved to the new index
- The list specified by `id` must exist and belong to the authenticated user

#### Example Usage (jQuery):

```javascript
/**
 * @type {addStockListTicker}
 */
var exampleParams = {
    "cmd": "addStockListTicker",
    "SID": "<SID>",
    "params": {
        "id": 2,              // Add to list ID 2
        "ticker": "AAPL.US",  // Add Apple stock
        "index": 2            // Insert at position 2
    }
};

function addStockListTicker(callback) {
    $.getJSON("https://tradernet.com/api/", {q: JSON.stringify(exampleParams)}, callback);
}

/**
 * Get the object
 **/
addStockListTicker(function(json){
    console.log(json);
    console.log("Ticker added. Updated lists:", json.userStockLists);
});
```

**Use Cases:**
- Building custom watchlists
- Organizing securities by strategy, sector, or theme
- Creating favorites lists for quick access
- Managing multiple portfolios or trading ideas

**Related Endpoints:**
- `addSecuritiesList` - Create a new securities list
- `deleteTickerFromList` - Remove a ticker from a list
- `getSecuritiesLists` - Get all user's securities lists
- `updateSecuritiesList` - Modify list properties (name, picture)
- `deleteSecuritiesList` - Delete a securities list

---

### Deleting ticker from list - `deleteStockListTicker`

Remove a ticker from a user's securities list (watchlist).

**Description:** Delete a specific ticker symbol from one of the user's custom securities lists.

**HTTP Method:** POST
**Command:** `deleteStockListTicker`

**Authorization:** Required. The SID received during authorization must be passed.

#### Request Parameters:

```json
{
    "cmd": "deleteStockListTicker",
    "SID": "<SID>",
    "params": {
        "id": 2,              // integer - Required
        "ticker": "AAPL.US"   // string - Required
    }
}
```

#### Parameter Details:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `cmd` | string | Yes | Request execution command ("deleteStockListTicker") |
| `SID` | string | Yes | Session ID received during authorization |
| `params` | object | Yes | Request execution parameters |
| `params.id` | integer | Yes | List ID (the ID of the securities list to remove the ticker from) |
| `params.ticker` | string | Yes | Ticker symbol to remove (e.g., "AAPL.US") |

#### Response Structure (Success):

```typescript
/**
 * Stock List
 */
type StockList = {
    id: number,           // List ID
    userId: number,       // User ID who owns the list
    name: string,         // List name
    tickers: string[],    // Array of ticker symbols in the list
    picture: string | null // Emoji or image for the list (optional)
}

/**
 * Delete Stock List Ticker Response
 */
type DeleteStockListTickerResponse = {
    userStockLists: StockList[],  // All user's stock lists (updated)
    selectedId: number,            // Currently selected list ID
    defaultId: number             // Default list ID
}
```

#### Example Response:

```json
{
    "userStockLists": [
        {
            "id": 1,
            "userId": 123456,
            "name": "default",
            "tickers": [],
            "picture": null
        },
        {
            "id": 2,
            "userId": 123456,
            "name": "etf",
            "tickers": [
                "AAAU.US",
                "ACES.US",
                "ACIO.US",
                "AFIF.US"
            ],
            "picture": "🙂"
        }
    ],
    "selectedId": 1,
    "defaultId": 1
}
```

#### Response (Error):

```json
// Common error
{
    "errMsg": "Bad json",
    "code": 2
}
```

#### Response Fields:

| Field | Type | Description |
|-------|------|-------------|
| `userStockLists` | array | Complete array of all user's securities lists (updated after removing ticker) |
| `userStockLists[].id` | number | Unique list identifier |
| `userStockLists[].userId` | number | ID of the user who owns the list |
| `userStockLists[].name` | string | Display name of the list |
| `userStockLists[].tickers` | array | Array of ticker symbols in this list (ticker removed) |
| `userStockLists[].picture` | string\|null | Emoji or image representing the list (optional) |
| `selectedId` | number | ID of the currently selected/active list |
| `defaultId` | number | ID of the user's default list |

**Important Notes:**
- Response returns **all** user's lists, not just the modified one
- The `tickers` array shows the updated list with the ticker removed
- If the ticker doesn't exist in the list, the operation succeeds silently (no error)
- The list specified by `id` must exist and belong to the authenticated user
- The ticker is removed from the specified list only, not from other lists

#### Example Usage (jQuery):

```javascript
/**
 * @type {deleteStockListTicker}
 */
var exampleParams = {
    "cmd": "deleteStockListTicker",
    "SID": "<SID>",
    "params": {
        "id": 2,              // Remove from list ID 2
        "ticker": "AAPL.US"   // Remove Apple stock
    }
};

function deleteStockListTicker(callback) {
    $.getJSON("https://tradernet.com/api/", {q: JSON.stringify(exampleParams)}, callback);
}

/**
 * Get the object
 **/
deleteStockListTicker(function(json){
    console.log(json);
    console.log("Ticker removed. Updated lists:", json.userStockLists);
});
```

**Use Cases:**
- Cleaning up watchlists
- Removing securities that no longer match list criteria
- Managing list organization
- Removing duplicates across multiple lists

**Related Endpoints:**
- `addStockListTicker` - Add a ticker to a list
- `addSecuritiesList` - Create a new securities list
- `getSecuritiesLists` - Get all user's securities lists
- `updateSecuritiesList` - Modify list properties (name, picture)
- `deleteSecuritiesList` - Delete an entire securities list

---

## QUOTES AND TICKERS (continued)

### Subscribe to stock quotes updates

**Status:** ⏳ AWAITING DOCUMENTATION

---

### Get a quote

**Status:** ⏳ AWAITING DOCUMENTATION

---

### Subscribe to market depth data

**Status:** ⏳ AWAITING DOCUMENTATION

---

### Get trades

**Status:** ⏳ AWAITING DOCUMENTATION

---

### Directory of securities - `getReadyList`

**Status:** ⏳ AWAITING DOCUMENTATION

---

### Checking instruments allowed for trading

**Status:** ⏳ AWAITING DOCUMENTATION

---

## WEBSOCKET - REAL-TIME DATA

### Connecting to websocket server

Connect to Tradernet's WebSocket server for real-time data streaming.

**Description:** WebSocket connection provides real-time updates for quotes, portfolio, orders, and market data.

**Server Address:** `wss://wss.tradernet.com/`

**Authorization:** Optional (SID parameter)
- **With SID:** Access to live account data (portfolio, orders, quotes)
- **Without SID:** Demo data only (quotes and market depth)

#### Connection Format

All WebSocket requests and responses use the format:

```javascript
[event, data]  // event: string, data: optional JSON object
```

- `event` - Event name (string)
- `data` - Request/response data in JSON format (optional)

#### Connection Methods

**1. Connect without SID (Demo Mode):**
```javascript
const ws = new WebSocket('wss://wss.tradernet.com/');
```

**2. Connect with SID (Live Account):**
```javascript
const ws = new WebSocket('wss://wss.tradernet.com/?SID=<your-sid>');
```

**Note:** To obtain a SID, see the 'Authorization' section.

#### Basic Connection Example

```javascript
const WebSocketsURL = "wss://wss.tradernet.com/";

const ws = new WebSocket(WebSocketsURL);

// Connection event
ws.onopen = function () {
    console.log('Connected to WS');
};

// Incoming message processing
ws.onmessage = function (m) {
    const [event, data] = JSON.parse(m.data);
    console.log(event, data);
};

// Connection closure processing
ws.onclose = function (e) {
    console.log('sockets closed', e);
};

// Error processing
ws.onerror = function (error) {
    console.log("Sockets.error: ", error);
    ws.close();
};
```

**Learn more:** [WebSocket object's methods and events](https://developer.mozilla.org/en-US/docs/Web/API/WebSocket)

#### Reconnection Logic

**Important:** Out-of-the-box WebSocket doesn't support automatic reconnection. Implement reconnection logic for production use:

```javascript
(function websocketStart() {
    const ws = new WebSocket(WebSocketsURL);

    ws.onopen = function () {
        console.log('Connected to WS');
    };

    ws.onmessage = function (m) {
        const [event, data] = JSON.parse(m.data);
        console.log(event, data);
    };

    ws.onclose = function (e) {
        console.log('sockets closed', e);
        setTimeout(function () {
            websocketStart();  // Reconnect after 5 seconds
        }, 5000);
    };

    ws.onerror = function (error) {
        console.log("Sockets.error: ", error);
        ws.close();
    };
})();
```

#### Initial Response: `userData` Event

When connecting, the server sends a `userData` event with user information:

```javascript
[
    "userData",
    {
        "isDemo": false,
        "mode": "prod",
        "authLogin": "user@example.com",
        "login": "user1",
        "clientLogin": "user@example.com"
    }
]
```

#### `userData` Event Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `isDemo` | bool | Demo mode indicator. `true` if: (1) No SID provided, (2) SID authorization failed, or (3) User has no live account |
| `mode` | string | Mode string representation: `"prod"` or `"demo"` |
| `authLogin` | string | Authentication login (email used for login) |
| `login` | string | User login (username) |
| `clientLogin` | string | Username under which the user logged in |

#### Subscribing to Events

To subscribe to an event, send the event name and parameters (if required):

```javascript
// Subscribing to quotes
ws.onopen = function() {
    // Wait for connection to open
    ws.send(JSON.stringify(['quotes', ['AAPL.SPB']]));
};

// Incoming data processing with event handlers
function quotesHandler(data) {
    console.log(data);
}

const handlers = {
    q: quotesHandler  // 'q' event for quotes
};

ws.onmessage = function (m) {
    const [event, data] = JSON.parse(m.data);
    if (handlers[event]) {
        handlers[event](data);
    }
};
```

#### Connection States

WebSocket connection states:
- `CONNECTING (0)` - Connection is being established
- `OPEN (1)` - Connection is open and ready to communicate
- `CLOSING (2)` - Connection is being closed
- `CLOSED (3)` - Connection is closed

Check connection state:
```javascript
console.log(ws.readyState);  // 0, 1, 2, or 3
```

#### Available Event Subscriptions

Once connected, you can subscribe to various events:
- **quotes** - Real-time stock quotes
- **portfolio** - Portfolio updates
- **orders** - Order status updates
- **orderbook** - Market depth data
- **markets** - Market status changes
- **sessions** - Security session changes

See individual event documentation for subscription formats and data structures.

#### Error Handling Best Practices

1. **Implement reconnection logic** - Network interruptions are common
2. **Handle authentication failures** - Check `isDemo` flag in `userData`
3. **Validate event data** - Check for required fields before processing
4. **Rate limit subscriptions** - Don't subscribe to too many symbols at once
5. **Clean up on disconnect** - Clear timers and event handlers

#### Connection with SID Example

```javascript
// Get SID from authentication
const sid = '<your-sid-from-auth>';

// Connect with live account access
const ws = new WebSocket(`wss://wss.tradernet.com/?SID=${sid}`);

ws.onopen = function() {
    console.log('Connected with live account');
};

ws.onmessage = function(m) {
    const [event, data] = JSON.parse(m.data);

    if (event === 'userData') {
        console.log('User data:', data);
        if (!data.isDemo) {
            // Live account connected - can subscribe to portfolio/orders
            ws.send(JSON.stringify(['portfolio']));
            ws.send(JSON.stringify(['orders']));
        }
    }
};
```

#### Demo Mode Limitations

When connected **without SID** or with **demo account**:
- ✅ Real-time quotes available
- ✅ Market depth data available
- ❌ Portfolio subscription unavailable
- ❌ Order subscription unavailable
- ❌ Account-specific data unavailable

**Usage Notes:**
- WebSocket provides **push updates** (more efficient than polling)
- Use for **real-time features** (live quotes, order status)
- Complement REST API (use REST for initial data, WebSocket for updates)
- **Reconnection is critical** for production reliability
- Monitor `userData.isDemo` to ensure live account access

---

### Subscribe to quotes (WebSocket)

Subscribe to real-time quote updates via WebSocket.

**Description:** Receive live price updates, bid/ask spreads, volume, and other market data for specified securities.

**Availability:** WebSocket only (not available via REST API)

**Event Name:** `q` (quotes)

**Authorization:** Optional (works in both demo and live modes)

#### Subscription Format

```javascript
// Subscribe to one or more tickers
ws.send(JSON.stringify(['quotes', ['AAPL.US', 'MSFT.US', 'TSLA.US']]));
```

#### Server Response Event

The server sends `q` events with quote updates:

```javascript
["q", [
    {
        "c": "AAPL.US",
        "ltp": 150.25,
        "chg": 2.50,
        // ... other fields
    }
]]
```

#### Quote Data Fields

Complete list of fields in quote updates:

| Field | Name | Type | Description |
|-------|------|------|-------------|
| `c` | Ticker | string | Ticker symbol |
| `ltr` | Exchange | string | Exchange of the latest trade |
| `name` | Name | string | Name of security |
| `name2` | Name Latin | string | Security name in Latin characters |
| `bbp` | Best Bid Price | float | Best bid price |
| `bbc` | Best Bid Change | string | Bid change indicator: `''` (no change), `'D'` (down), `'U'` (up) |
| `bbs` | Best Bid Size | int | Best bid size (quantity) |
| `bbf` | Best Bid Volume | float | Best bid volume (value) |
| `bap` | Best Ask Price | float | Best offer/ask price |
| `bac` | Best Ask Change | string | Ask change indicator: `''` (no change), `'D'` (down), `'U'` (up) |
| `bas` | Best Ask Size | int | Size of the best offer |
| `baf` | Best Ask Volume | float | Volume of the best offer |
| `pp` | Previous Close | float | Previous closing price |
| `op` | Open | float | Opening price of current trading session |
| `ltp` | Last Trade Price | float | Last trade price |
| `lts` | Last Trade Size | int | Last trade size (quantity) |
| `ltt` | Last Trade Time | string | Time of last trade |
| `chg` | Change | float | Price change in points relative to previous close |
| `pcp` | Change Percent | float | Percentage change relative to previous close |
| `ltc` | Last Trade Change | string | Price change indicator: `''` (no change), `'D'` (down), `'U'` (up) |
| `mintp` | Day Low | float | Minimum trade price for the day |
| `maxtp` | Day High | float | Maximum trade price for the day |
| `vol` | Volume | int | Trade volume for the day (in shares/units) |
| `vlt` | Volume Currency | float | Trading volume for the day in currency |
| `yld` | Yield | float | Yield to maturity (for bonds) |
| `acd` | Accrued Interest | float | Accumulated coupon interest (ACI) |
| `fv` | Face Value | float | Face value (for bonds) |
| `mtd` | Maturity Date | string | Maturity date (for bonds) |
| `cpn` | Coupon | float | Coupon in currency (for bonds) |
| `cpp` | Coupon Period | int | Coupon period in days (for bonds) |
| `ncd` | Next Coupon Date | string | Next coupon date (for bonds) |
| `ncp` | Latest Coupon Date | string | Latest coupon date (for bonds) |
| `dpd` | Margin Long | float | Purchase margin (long position) |
| `dps` | Margin Short | float | Short sale margin |
| `trades` | Trades Count | int | Number of trades for the day |
| `min_step` | Min Step | float | Minimum price increment |
| `step_price` | Step Price | float | Price increment value |

#### Example Implementation

```javascript
var WebSocketsURL = "wss://wss.tradernet.com/";
var ws = new WebSocket(WebSocketsURL);

var tickersToWatchChanges = ["AAPL.US", "MSFT.US", "TSLA.US"];

/**
 * Process quote updates
 * @param {QuoteUpdate[]} data
 */
function updateWatcher(data) {
    data.forEach(function(quote) {
        console.log(`${quote.c}: ${quote.ltp} (${quote.chg >= 0 ? '+' : ''}${quote.chg})`);
    });
}

ws.onmessage = function (m) {
    const [event, data] = JSON.parse(m.data);

    if (event === 'q') {
        updateWatcher(data);
    }
};

ws.onopen = function() {
    console.log('Connected, subscribing to quotes...');
    ws.send(JSON.stringify(['quotes', tickersToWatchChanges]));
};
```

#### Change Indicators

The `bbc`, `bac`, and `ltc` fields indicate price movement:
- `''` (empty string) - No change
- `'D'` - Down (price decreased)
- `'U'` - Up (price increased)

**Important Notes:**
- Quotes are available in both **demo** and **live** modes
- **No SID required** for basic quote data
- Updates are **push-based** (no polling needed)

---

### Subscribe to market depth (WebSocket)

Subscribe to real-time market depth (order book) updates via WebSocket.

**Description:** Receive Level 2 market data showing bid/ask depth with prices and quantities at multiple levels.

**Availability:** WebSocket only (not available via REST API)

**Event Name:** `b` (book/depth)

**Authorization:** Optional (works in both demo and live modes)

#### Subscription Format

```javascript
// Subscribe to order book for one or more tickers
ws.send(JSON.stringify(["orderBook", ["AAPL.US"]]));
```

#### Server Response Event

The server sends `b` events with market depth updates:

```javascript
["b", {
    "n": 102,
    "i": "AAPL.US",
    "del": [],
    "ins": [],
    "upd": [
        {"p": 33.925, "s": "S", "q": 196100, "k": 3},
        {"p": 33.89, "s": "S", "q": 373700, "k": 6}
    ],
    "cnt": 21,
    "x": 11
}]
```

#### Market Depth Data Structure

```typescript
/**
 * Market Depth Row
 */
type DomInfoRow = {
    k: number,      // Position number in the market depth (level)
    p: number,      // Price at this level
    q: number,      // Quantity at this level
    s: 'S' | 'B'   // Side: 'S' = Sell (ask), 'B' = Buy (bid)
}

/**
 * Market Depth Update Block
 */
type DomInfoBlock = {
    i: string,           // Ticker symbol
    n: number,           // Update sequence number
    cnt: number,         // Depth of market data (number of levels)
    x: number,           // Additional metadata
    ins: DomInfoRow[],   // New rows to insert in market depth
    del: DomInfoRow[],   // Market depth rows to delete
    upd: DomInfoRow[]    // Market depth rows to update
}
```

#### Response Fields

| Field | Type | Description |
|-------|------|-------------|
| `i` | string | Ticker symbol for which market depth information was received |
| `n` | number | Update sequence number |
| `cnt` | number | Total depth of market data (number of price levels) |
| `x` | number | Additional metadata field |
| `ins` | array | New price levels to insert into the order book |
| `del` | array | Price levels to remove from the order book |
| `upd` | array | Price levels to update in the order book |

#### Row Fields

| Field | Type | Description |
|-------|------|-------------|
| `k` | number | Position/level number in the market depth (1 = closest to spread) |
| `p` | number | Price at this level |
| `q` | number | Total quantity available at this price level |
| `s` | string | Side indicator: `'S'` = Sell/Ask side, `'B'` = Buy/Bid side |

#### Example Implementation

```javascript
const ticker = 'AAPL.US';
const WebSocketsURL = "wss://wss.tradernet.com/";
const ws = new WebSocket(WebSocketsURL);

// Store order book state
const orderBook = {
    bids: {},  // { level: { price, quantity } }
    asks: {}   // { level: { price, quantity } }
};

function processDepthUpdate(data) {
    console.log(`Market Depth Update for ${data.i}`);

    // Process deletions
    data.del.forEach(row => {
        const side = row.s === 'B' ? 'bids' : 'asks';
        delete orderBook[side][row.k];
    });

    // Process insertions
    data.ins.forEach(row => {
        const side = row.s === 'B' ? 'bids' : 'asks';
        orderBook[side][row.k] = { price: row.p, quantity: row.q };
    });

    // Process updates
    data.upd.forEach(row => {
        const side = row.s === 'B' ? 'bids' : 'asks';
        orderBook[side][row.k] = { price: row.p, quantity: row.q };
    });

    // Display top 5 levels
    console.log('Top 5 Asks:');
    Object.keys(orderBook.asks).slice(0, 5).forEach(k => {
        const level = orderBook.asks[k];
        console.log(`  ${level.price} x ${level.quantity}`);
    });

    console.log('Top 5 Bids:');
    Object.keys(orderBook.bids).slice(0, 5).forEach(k => {
        const level = orderBook.bids[k];
        console.log(`  ${level.price} x ${level.quantity}`);
    });
}

ws.onmessage = function (m) {
    const [event, data] = JSON.parse(m.data);

    if (event === 'b') {
        processDepthUpdate(data);
    }
};

ws.onopen = function() {
    console.log('Connected, subscribing to order book...');
    ws.send(JSON.stringify(["orderBook", [ticker]]));
};
```

#### Update Types

Market depth updates come in three categories:

1. **Insertions (`ins`)** - New price levels added to the order book
2. **Deletions (`del`)** - Price levels removed from the order book
3. **Updates (`upd`)** - Existing price levels with quantity changes

#### Maintaining Order Book State

To maintain accurate order book state:

1. **Initialize** - Start with empty order book
2. **Apply deletions** - Remove specified levels
3. **Apply insertions** - Add new levels
4. **Apply updates** - Modify existing levels
5. **Sort** - Keep bids descending, asks ascending by price

#### Side Indicators

- `'B'` = **Bid** side (buy orders) - highest bid price closest to spread
- `'S'` = **Ask** side (sell/offer orders) - lowest ask price closest to spread

#### Level Numbers (`k`)

- Level 1 (`k=1`) is closest to the spread (best bid/ask)
- Higher levels are further from the spread
- Level numbers indicate depth position

#### Example Response

```json
{
    "n": 102,
    "i": "AAPL.US",
    "del": [],
    "ins": [],
    "upd": [
        {"p": 33.925, "s": "S", "q": 196100, "k": 3},
        {"p": 33.89, "s": "S", "q": 373700, "k": 6},
        {"p": 33.885, "s": "S", "q": 1198800, "k": 7},
        {"p": 33.88, "s": "S", "q": 251600, "k": 8}
    ],
    "cnt": 21,
    "x": 11
}
```

This update shows 4 ask levels being updated with new quantities.

#### Use Cases

- **Order placement** - See available liquidity at different price levels
- **Slippage estimation** - Calculate market impact of large orders
- **Market microstructure** - Analyze order book dynamics
- **Limit order positioning** - Place orders based on depth
- **Spread analysis** - Monitor bid-ask spread changes

#### Best Practices

1. **Maintain full state** - Apply all updates (ins/del/upd) sequentially
2. **Handle missing data** - Some updates may skip levels
3. **Validate sequence** - Use `n` field to detect missed updates
4. **Aggregate levels** - Sum quantities at same price if needed
5. **Display top levels** - Show most relevant depth (typically 5-10 levels)

**Important Notes:**
- Market depth available in both **demo** and **live** modes
- **No SID required** for depth data
- Updates are **incremental** - must maintain state
- High-frequency updates for liquid securities
- Useful for **advanced trading** strategies

---

### Subscribing to security sessions (WebSocket)

**Status:** ⏳ AWAITING DOCUMENTATION

---

### Subscribe to portfolio (WebSocket)

Subscribe to real-time portfolio updates via WebSocket.

**Description:** Receive live updates about your portfolio positions, account balances, and P&L changes.

**Availability:** WebSocket only (not available via REST API)

**Event Name:** `portfolio`

**Authorization:** **Required** (SID must be provided) - Does not work in demo mode

#### Subscription Format

```javascript
// Subscribe to portfolio updates (no parameters needed)
ws.send(JSON.stringify(["portfolio"]));
```

#### Server Response Event

The server sends `portfolio` events with portfolio updates:

```javascript
["portfolio", {
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
            "mkt_price": 23.81,
            "profit_price": 2.83,
            // ... other fields
        }
    ]
}]
```

#### Portfolio Data Structure

```typescript
/**
 * Account Balance Information
 */
type AccountInfoRow = {
    curr: string,        // Account currency
    currval: number,     // Account currency exchange rate
    forecast_in: number, // Forecasted incoming funds
    forecast_out: number, // Forecasted outgoing funds
    s: number,           // Available funds
    t2_in: string,       // T+2 incoming
    t2_out: string       // T+2 outgoing
}

/**
 * Position Information
 */
type PositionInfoRow = {
    acc_pos_id: number,      // Unique position ID in Tradernet system
    accruedint_a: number,    // Accumulated coupon income (ACI)
    curr: string,            // Position currency
    currval: number,         // Currency exchange rate
    fv: number,              // Coefficient to calculate initial margin
    go: number,              // Initial margin per position
    i: string,               // Position ticker symbol
    k: number,               // Kind/type
    q: number,               // Quantity of securities in position
    s: number,               // Size/value
    t: number,               // Type
    t2_in: string,           // T+2 settlement incoming
    t2_out: string,          // T+2 settlement outgoing
    vm: number,              // Variable margin of position
    name: string,            // Issuer name
    name2: string,           // Issuer alternative name
    mkt_price: number,       // Current market price
    market_value: number,    // Total market value of position
    bal_price_a: number,     // Book value (average price)
    open_bal: number,        // Position book value at open
    price_a: number,         // Book value when position opened
    profit_close: number,    // Previous day profit/loss
    profit_price: number,    // Current position profit/loss
    close_price: number,     // Position closing price
    trade: Array<{trade_count: number}>, // Trade information
    base_currency: string,   // Base currency
    face_val_a: number,      // Face value (for bonds)
    scheme_calc: string,     // Calculation scheme (e.g., "T2")
    instr_id: number,        // Instrument ID
    Yield: number,           // Yield (for bonds)
    issue_nb: string,        // ISIN code
    acd: number              // Accumulated coupon date
}

/**
 * Portfolio Response
 */
type SocketPortfolioResponse = {
    key: string,                  // User identifier
    acc: AccountInfoRow[],        // Account balances (by currency)
    pos: PositionInfoRow[]        // Open positions
}
```

#### Account Balance Fields (`acc`)

| Field | Type | Description |
|-------|------|-------------|
| `curr` | string | Account currency code (USD, EUR, etc.) |
| `currval` | number | Currency exchange rate to base currency |
| `s` | number | Available funds in this currency |
| `forecast_in` | number | Forecasted incoming funds |
| `forecast_out` | number | Forecasted outgoing funds |
| `t2_in` | string | T+2 settlement incoming |
| `t2_out` | string | T+2 settlement outgoing |

#### Position Fields (`pos`)

| Field | Type | Description |
|-------|------|-------------|
| `i` | string | Ticker symbol |
| `q` | number | Quantity of shares/units held |
| `curr` | string | Position currency |
| `currval` | number | Currency exchange rate |
| `name` | string | Security name |
| `name2` | string | Alternative security name |
| `mkt_price` | number | Current market price per unit |
| `market_value` | number | Total market value (qty × price) |
| `bal_price_a` | number | Average purchase price (book value) |
| `price_a` | number | Book value when opened |
| `open_bal` | number | Opening book value |
| `profit_close` | number | Previous day P&L |
| `profit_price` | number | Current unrealized P&L |
| `close_price` | number | Closing price |
| `go` | number | Initial margin requirement |
| `vm` | number | Variable margin |
| `acc_pos_id` | number | Unique position identifier |
| `accruedint_a` | number | Accumulated coupon income (bonds) |
| `acd` | number | Accrued coupon date |
| `fv` | number | Face value coefficient |
| `base_currency` | string | Base currency |
| `scheme_calc` | string | Settlement calculation scheme |
| `instr_id` | number | Instrument ID |
| `Yield` | number | Yield (for bonds) |
| `issue_nb` | string | ISIN code |
| `t` | number | Type code |
| `k` | number | Kind code |
| `s` | number | Size/value |
| `trade` | array | Trade count information |

#### Example Implementation

```javascript
const WS_SOCKETURL = 'wss://wss.tradernet.com/?SID=<your-sid>';
const ws = new WebSocket(WS_SOCKETURL);

ws.onopen = function() {
    console.log('Connected, subscribing to portfolio...');
    // Subscribe to portfolio updates
    ws.send(JSON.stringify(["portfolio"]));
};

ws.onmessage = function (m) {
    const [event, data] = JSON.parse(m.data);

    if (event === 'portfolio') {
        console.log('Portfolio Update:');

        // Process account balances
        data.acc.forEach(account => {
            console.log(`${account.curr}: ${account.s} available`);
        });

        // Process positions
        data.pos.forEach(position => {
            const totalValue = position.market_value;
            const pnl = position.profit_price;
            const pnlPercent = ((position.mkt_price - position.bal_price_a) / position.bal_price_a * 100).toFixed(2);

            console.log(`${position.i}: ${position.q} shares`);
            console.log(`  Current: ${position.mkt_price} | Avg Cost: ${position.bal_price_a}`);
            console.log(`  Value: ${totalValue} | P&L: ${pnl} (${pnlPercent}%)`);
        });
    }
};

ws.onerror = function(error) {
    console.error('WebSocket error:', error);
};

ws.onclose = function(e) {
    console.log('WebSocket closed:', e);
};
```

#### Calculating Position Metrics

**Total Market Value:**
```javascript
const marketValue = position.q * position.mkt_price;
// or use position.market_value directly
```

**Unrealized P&L:**
```javascript
const unrealizedPnL = (position.mkt_price - position.bal_price_a) * position.q;
// or use position.profit_price directly
```

**P&L Percentage:**
```javascript
const pnlPercent = ((position.mkt_price - position.bal_price_a) / position.bal_price_a) * 100;
```

**Cost Basis:**
```javascript
const costBasis = position.bal_price_a * position.q;
```

#### Update Frequency

- Portfolio updates are sent **in real-time** when:
  - Trades are executed (positions change)
  - Market prices change (affects unrealized P&L)
  - Cash balance changes (deposits, withdrawals, settlements)
  - Margin requirements update

#### Multi-Currency Support

- Each account balance (`acc`) represents one currency
- Positions can be in different currencies
- Use `currval` to convert to base currency
- `base_currency` field indicates the main currency

#### Settlement Information

T+2 settlement fields (`t2_in`, `t2_out`):
- Show pending settlements
- Funds not yet available
- Important for cash management

#### Use Cases

- **Portfolio monitoring** - Track positions and P&L in real-time
- **Risk management** - Monitor margin requirements
- **Performance tracking** - Calculate returns live
- **Rebalancing** - Detect when positions drift from targets
- **Alerts** - Trigger notifications on P&L thresholds
- **Dashboard displays** - Show live portfolio value

#### Best Practices

1. **Authenticate first** - SID is required for portfolio subscription
2. **Handle updates efficiently** - Updates can be frequent during market hours
3. **Calculate totals** - Sum across positions for portfolio-level metrics
4. **Track changes** - Compare with previous state to detect what changed
5. **Currency conversion** - Use `currval` for accurate multi-currency totals
6. **Handle empty portfolio** - `pos` array will be empty if no positions

**Important Notes:**
- **SID required** - Portfolio subscription only works with authenticated connection
- **Personal data** - Contains sensitive account information
- **Real-time updates** - Reflects market price changes immediately
- **Multi-currency** - Handle different currencies properly
- **Settlement timing** - Consider T+2 settlement for cash management
- **Margin data** - Includes margin requirements for leveraged positions

---

### Subscribe to orders (WebSocket)

Subscribe to real-time order updates via WebSocket.

**Description:** Receive live updates about order status changes, fills, cancellations, and other order lifecycle events.

**Availability:** WebSocket only (not available via REST API)

**Event Name:** `orders`

**Authorization:** **Required** (SID must be provided) - Does not work in demo mode

#### Subscription Format

```javascript
// Subscribe to order updates (no parameters needed)
ws.send(JSON.stringify(['orders']));
```

#### Server Response Event

The server sends `orders` events with order updates:

```javascript
["orders", [{
    "order_id": 8757875,
    "instr": "FCX.US",
    "oper": 2,
    "type": 1,
    "p": 6.5611,
    "q": 2625,
    "stat": 21,
    // ... other fields
}]]
```

#### Order Data Structure

```typescript
/**
 * Order Data Row
 */
type OrderDataRow = {
    aon: 0 | 1,              // All or Nothing: 0 = can be partially executed, 1 = cannot be partially executed
    cur: string,             // Order currency
    curr_q: number,          // Current quantity
    date: string,            // Order date (ISO 8601)
    exp: 1 | 2 | 3,          // Expiration: 1 = Day, 2 = Day+Ext, 3 = GTC
    fv: number,              // Coefficient for relative currencies (futures)
    order_id: number,        // Tradernet unique order ID
    instr: string,           // Ticker symbol
    leaves_qty: number,      // Remaining quantity (unfilled)
    auth_login: string,      // Login of client who sent the order
    creator_login: string,   // Login of order creator
    owner_login: string,     // Login of user for whom order was created
    mkt_id: number,          // Market unique trade ID
    name: string,            // Company name
    name2: string,           // Alternative company name
    oper: number,            // Operation: 1 = Buy, 2 = Buy on Margin, 3 = Sell, 4 = Sell Short
    p: number,               // Order price
    q: number,               // Order quantity
    rep: number,             // Report field
    stat: number,            // Order status (see Order Status codes)
    stat_d: string,          // Status modification date
    stat_orig: number,       // Initial order status
    stat_prev: number,       // Previous order status
    stop: number,            // Stop price
    stop_activated: 0 | 1,   // Stop order activation indicator
    stop_init_price: number, // Price to activate stop order
    trailing_price: number,  // Trailing order variance percentage
    type: 1 | 2 | 3 | 4 | 5 | 6, // Order type (see Order Types)
    user_order_id: number,   // User-assigned order ID
    trade: OrderTradeInfo[]  // Trade list for this order
}

/**
 * Order Trade Information
 */
type OrderTradeInfo = {
    acd: number,    // Accumulated coupon interest
    date: string,   // Trade date
    fv: number,     // Coefficient for relative currencies
    go_sum: number, // Initial margin per trade
    id: number,     // Tradernet unique trade ID
    p: number,      // Trade price
    profit: number, // Trade profit
    q: number,      // Number of securities in trade
    v: number       // Trade amount (value)
}

/**
 * Orders Response
 */
type SocketOrdersResponse = OrderDataRow[];
```

#### Order Field Reference

**Order Identification:**
| Field | Type | Description |
|-------|------|-------------|
| `order_id` | number | Tradernet unique order ID |
| `user_order_id` | number | User-assigned order ID (from placement) |
| `instr` | string | Ticker symbol |

**Order Details:**
| Field | Type | Description |
|-------|------|-------------|
| `oper` | number | Operation: `1` = Buy, `2` = Buy on Margin, `3` = Sell, `4` = Sell Short |
| `type` | number | Order type: `1` = Market, `2` = Limit, `3` = Stop, `4` = Stop Limit, `5` = StopLoss, `6` = TakeProfit |
| `p` | number | Order price |
| `q` | number | Order quantity (total) |
| `leaves_qty` | number | Remaining unfilled quantity |
| `cur` | string | Order currency |

**Order Status:**
| Field | Type | Description |
|-------|------|-------------|
| `stat` | number | Current order status (see Order Status Codes reference) |
| `stat_orig` | number | Initial order status (same as `stat` when working with API) |
| `stat_prev` | number | Previous order status |
| `stat_d` | string | Status modification date/time |

**Order Timing:**
| Field | Type | Description |
|-------|------|-------------|
| `date` | string | Order creation date/time (ISO 8601) |
| `exp` | number | Expiration: `1` = Day (end-of-day), `2` = Day+Ext, `3` = GTC (good-til-cancel) |

**Stop Orders:**
| Field | Type | Description |
|-------|------|-------------|
| `stop` | number | Stop price |
| `stop_init_price` | number | Price to activate stop order |
| `stop_activated` | 0\|1 | Stop order activation indicator: `0` = not activated, `1` = activated |
| `trailing_price` | number | Trailing order variance percentage |

**Order Settings:**
| Field | Type | Description |
|-------|------|-------------|
| `aon` | 0\|1 | All or Nothing: `0` = can be partially filled, `1` = must be filled completely |

**User Information:**
| Field | Type | Description |
|-------|------|-------------|
| `auth_login` | string | Login of client who authenticated |
| `creator_login` | string | Login of order creator |
| `owner_login` | string | Login of user for whom order was created |

**Trade Information:**
| Field | Type | Description |
|-------|------|-------------|
| `trade` | array | Array of trade executions for this order |
| `trade[].id` | number | Unique trade ID |
| `trade[].p` | number | Execution price |
| `trade[].q` | number | Executed quantity |
| `trade[].v` | number | Trade value (price × quantity) |
| `trade[].date` | string | Trade execution date/time |
| `trade[].profit` | number | Trade profit/loss |

#### Order Type Codes

| Code | Type | Description |
|------|------|-------------|
| `1` | Market Order | Execute at best available price |
| `2` | Limit Order | Execute at specified price or better |
| `3` | Stop Order | Becomes market order when stop price reached |
| `4` | Stop Limit Order | Becomes limit order when stop price reached |
| `5` | StopLoss | Automatic stop loss order |
| `6` | TakeProfit | Automatic take profit order |

#### Operation Codes

| Code | Operation | Description |
|------|-----------|-------------|
| `1` | Buy | Buy without margin |
| `2` | Buy on Margin | Buy with leverage |
| `3` | Sell | Sell without margin |
| `4` | Sell Short | Short sale with margin |

#### Expiration Codes

| Code | Type | Description |
|------|------|-------------|
| `1` | Day | End-of-day (valid until market close) |
| `2` | Day+Ext | Day/Night or Night/Day (includes extended hours) |
| `3` | GTC | Good-til-cancel (valid until manually cancelled, includes night sessions) |

#### Example Implementation

```javascript
const WS_SOCKETURL = 'wss://wss.tradernet.com/?SID=<your-sid>';
const ws = new WebSocket(WS_SOCKETURL);

ws.onopen = function() {
    console.log('Connected, subscribing to orders...');
    // Subscribe to order updates
    ws.send(JSON.stringify(['orders']));
};

ws.onmessage = function ({ data }) {
    const [event, messageData] = JSON.parse(data);

    if (event === 'orders') {
        messageData.forEach(order => {
            console.log(`Order ${order.order_id}: ${order.instr}`);
            console.log(`  Type: ${getOrderType(order.type)} | Status: ${order.stat}`);
            console.log(`  Quantity: ${order.q} | Filled: ${order.q - order.leaves_qty} | Remaining: ${order.leaves_qty}`);
            console.log(`  Price: ${order.p} | Currency: ${order.cur}`);

            // Display trade executions
            if (order.trade && order.trade.length > 0) {
                console.log(`  Executions (${order.trade.length}):`);
                order.trade.forEach(trade => {
                    console.log(`    ${trade.q} @ ${trade.p} = ${trade.v}`);
                });
            }
        });
    }
};

// Helper function to decode order type
function getOrderType(typeCode) {
    const types = {
        1: 'Market',
        2: 'Limit',
        3: 'Stop',
        4: 'Stop Limit',
        5: 'StopLoss',
        6: 'TakeProfit'
    };
    return types[typeCode] || `Unknown (${typeCode})`;
}
```

#### Example Response

```json
[{
    "aon": 0,
    "cur": "USD",
    "curr_q": 0,
    "date": "2015-12-23T17:05:02.133",
    "exp": 1,
    "fv": 0,
    "order_id": 8757875,
    "instr": "FCX.US",
    "leaves_qty": 0,
    "auth_login": "virtual@virtual.com",
    "creator_login": "virtual@virtual.com",
    "owner_login": "virtual@virtual.com",
    "mkt_id": 30000000001,
    "name": "Freeport-McMoran Cp & Gld",
    "name2": "Freeport-McMoran Cp & Gld",
    "oper": 2,
    "p": 6.5611,
    "q": 2625,
    "rep": 0,
    "stat": 21,
    "stat_d": "2015-12-23T17:05:03.283",
    "stat_orig": 21,
    "stat_prev": 10,
    "stop": 0,
    "stop_activated": 1,
    "stop_init_price": 6.36,
    "trailing_price": 0,
    "type": 1,
    "user_order_id": 1450879514204,
    "trade": [{
        "acd": 0,
        "date": "2015-12-23T17:05:03",
        "fv": 100,
        "go_sum": 0,
        "id": 13446624,
        "p": 6.37,
        "profit": 0,
        "q": 2625,
        "v": 16721.25
    }]
}]
```

This shows a fully executed market order (stat: 21 = executed) with one trade execution.

#### Update Frequency

Order updates are sent in real-time when:
- Order is placed (initial status)
- Order status changes (pending → active → filled/cancelled)
- Partial fills occur (`leaves_qty` decreases)
- Order is modified
- Order is cancelled
- Stop orders are triggered (`stop_activated` becomes 1)

#### Order Status Tracking

Use the `stat`, `stat_orig`, and `stat_prev` fields to track order lifecycle:
- `stat_orig` - Initial status when order was placed
- `stat_prev` - Previous status before current update
- `stat` - Current status

See the **Order Status Codes** reference section for complete status list.

#### Partial Fills

Track partial fills using:
- `q` - Total order quantity
- `leaves_qty` - Remaining unfilled quantity
- Filled quantity = `q - leaves_qty`
- `trade` array - Individual execution details

#### Use Cases

- **Order monitoring** - Track order status in real-time
- **Execution alerts** - Notify when orders fill
- **Order management UI** - Display live order status
- **Trading algorithms** - React to order events
- **Audit trail** - Record all order state changes
- **Fill tracking** - Monitor partial and complete fills

#### Best Practices

1. **Authenticate first** - SID required for order subscription
2. **Track order lifecycle** - Use stat_prev and stat to detect transitions
3. **Handle partial fills** - Check leaves_qty for unfilled quantity
4. **Monitor stop activation** - Track stop_activated for stop orders
5. **Match user_order_id** - Use your own IDs to correlate orders
6. **Process trade array** - Extract execution details from trade list

**Important Notes:**
- **SID required** - Orders subscription only works with authenticated connection
- **Personal data** - Contains sensitive trading information
- **Real-time updates** - Instant notification of order changes
- **Trade executions** - `trade` array contains all fills
- **Status transitions** - Track order lifecycle via stat fields
- **Stop orders** - Monitor `stop_activated` and `stop_init_price`

---

### Subscribing to market statuses (WebSocket)

Subscribe to real-time market status updates via WebSocket.

**Description:** Receive live updates when markets open, close, or change status.

**Availability:** WebSocket only (not available via REST API)

**Event Name:** `markets`

**Authorization:** Optional (works in both demo and live modes)

#### Subscription Format

```javascript
// Subscribe to market status updates (no parameters needed)
ws.send(JSON.stringify(["markets"]));
```

#### Server Response Event

The server sends `markets` events with status updates:

```javascript
["markets", {
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
}]
```

#### Market Status Data Structure

```typescript
/**
 * Market Info Row
 */
type MarketInfoRow = {
    n: string,   // Full market name
    n2: string,  // Market abbreviation/code
    s: string,   // Current market status (OPEN, CLOSE, PRE_OPEN, POST_CLOSE)
    o: string,   // Market opening time (MSK timezone) - Format: "HH:MM:SS"
    c: string,   // Market closing time (MSK timezone) - Format: "HH:MM:SS"
    dt: string   // Time difference relative to MSK in minutes (e.g., "-180")
}

/**
 * Markets Status Response
 */
type MarketsStatusResponse = {
    t: string,              // Current server time (timestamp)
    m: MarketInfoRow[]      // Array of market status updates
}
```

#### Response Fields

| Field | Type | Description |
|-------|------|-------------|
| `t` | string | Current server request time (YYYY-MM-DD HH:MM:SS) |
| `m` | array | Array of market information objects |
| `m[].n` | string | Full market name (e.g., "KASE") |
| `m[].n2` | string | Market abbreviation/code |
| `m[].s` | string | Current market status: `"OPEN"`, `"CLOSE"`, `"PRE_OPEN"`, `"POST_CLOSE"` |
| `m[].o` | string | Market opening time in MSK timezone (HH:MM:SS) |
| `m[].c` | string | Market closing time in MSK timezone (HH:MM:SS) |
| `m[].dt` | string | Time difference relative to Moscow time in minutes (negative = behind, positive = ahead) |

#### Market Status Values

| Status | Description |
|--------|-------------|
| `OPEN` | Market is currently open for trading |
| `CLOSE` | Market is currently closed |
| `PRE_OPEN` | Pre-market session (before official open) |
| `POST_CLOSE` | After-hours session (after official close) |

#### Example Implementation

```javascript
const WS_SOCKETURL = 'wss://wss.tradernet.com/';
const ws = new WebSocket(WS_SOCKETURL);

ws.onopen = function() {
    console.log('Connected, subscribing to market statuses...');
    ws.send(JSON.stringify(["markets"]));
};

ws.onmessage = function (m) {
    const [event, data] = JSON.parse(m.data);

    if (event === 'markets') {
        console.log(`Market Status Update at ${data.t}:`);

        data.m.forEach(market => {
            console.log(`${market.n} (${market.n2}): ${market.s}`);
            console.log(`  Open: ${market.o} | Close: ${market.c}`);
            console.log(`  Timezone offset: ${market.dt} minutes from MSK`);

            // Calculate local market time
            const mskOffset = parseInt(market.dt);
            console.log(`  Local offset: ${mskOffset > 0 ? '+' : ''}${mskOffset / 60} hours from MSK`);
        });
    }
};

ws.onerror = function(error) {
    console.error('WebSocket error:', error);
};

ws.onclose = function(e) {
    console.log('WebSocket closed:', e);
};
```

#### Example Response

```json
{
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
```

This shows KASE (Kazakhstan Stock Exchange) is closed, opens at 08:20 MSK, closes at 14:00 MSK, and is 180 minutes (3 hours) behind Moscow time.

#### Time Zone Conversion

The `dt` field provides time zone offset from Moscow Standard Time (MSK):
- **Negative values** - Market is behind MSK (e.g., `-180` = 3 hours behind)
- **Positive values** - Market is ahead of MSK
- **Zero** - Market is in MSK timezone

**Example conversions:**
- `dt: "-180"` = UTC+2 (MSK is UTC+3, so -180 minutes = -3 hours)
- `dt: "60"` = UTC+4 (MSK + 1 hour)
- `dt: "0"` = UTC+3 (same as MSK)

#### Update Frequency

Market status updates are sent when:
- Markets open (CLOSE → OPEN)
- Markets close (OPEN → CLOSE)
- Pre-market begins (CLOSE → PRE_OPEN)
- After-hours begins (OPEN → POST_CLOSE)
- Extended hours end (PRE_OPEN/POST_CLOSE → CLOSE)

#### Use Cases

- **Trading hours** - Determine if orders can be placed
- **Market monitoring** - Track which markets are active
- **Automated trading** - Schedule operations based on market hours
- **Multi-market systems** - Coordinate across different exchanges
- **User notifications** - Alert when markets open/close
- **Order validation** - Prevent orders when markets closed

#### Best Practices

1. **Subscribe on connection** - Get initial market states
2. **Handle all statuses** - OPEN, CLOSE, PRE_OPEN, POST_CLOSE
3. **Convert timezones** - Use `dt` field for local time calculation
4. **Cache market info** - Store opening/closing times
5. **React to changes** - Update UI when status changes
6. **Validate orders** - Check market status before placing orders

**Important Notes:**
- **No SID required** - Market status works in demo and live modes
- **MSK timezone** - All times in Moscow Standard Time (UTC+3)
- **Time difference** - Use `dt` to convert to local timezone
- **Real-time updates** - Instant notification of status changes
- **Multiple markets** - Receive updates for all active markets
- **Extended hours** - PRE_OPEN and POST_CLOSE indicate special sessions

---

## VARIOUS / REFERENCE DATA

### List of existing offices

**Status:** ⏳ AWAITING DOCUMENTATION

---

### Name list of system files

**Status:** ⏳ AWAITING DOCUMENTATION

---

### Trading platforms

**Status:** ⏳ AWAITING DOCUMENTATION

---

### Instruments details - `getSecurityInfo`

**Command:** `getSecurityInfo`
**Method:** HTTPS GET / HTTPS POST (API V2)

#### Description

Method for obtaining detailed information about a security/ticker including name, currency, minimum price increment, and registration date.

#### Request Parameters

```json
{
    "cmd": "getSecurityInfo",
    "params": {
        "ticker": "AAPL.US",  // string - Required
        "sup": true           // bool - Required
    }
}
```

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `ticker` | string | Yes | The ticker symbol to retrieve information for |
| `sup` | bool | Yes | IMS and trading system format (boolean, NOT int!) |

#### Response Structure (Success)

```typescript
type SecurityInfoResponse = {
    id: number,            // Unique ticker ID
    short_name: string,    // Short ticker name
    default_ticker: string, // Ticker name on the Exchange
    nt_ticker: string,     // Ticker name in Tradernet system
    firstDate: string,     // Company registration date (DD.MM.YYYY)
    currency: string,      // Currency of a security
    min_step: number,      // Minimum price increment
    code: number          // Error code (0 = success)
}
```

**Example Response:**

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

#### Response Structure (Error)

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

#### Example Usage (jQuery)

```javascript
var exampleParams = {
    "cmd": "getSecurityInfo",
    "params": {
        "ticker": "FB.US",
        "sup": true
    }
};

function getSecurityInfo(callback) {
    $.getJSON("https://tradernet.com/api/", {q: JSON.stringify(exampleParams)}, callback);
}

// Get the object
getSecurityInfo(function(json){
    console.log(json);
});
```

---

### Instrument Types Reference Data

**Purpose:** Reference data describing all available instrument types and their type combinations in the Tradernet system.

#### Instrument Type Codes

Complete list of instrument types and their identifiers:

| Type Code | Type Name | Description |
|-----------|-----------|-------------|
| 1 | Stocks | Equities and shares |
| 2 | Bonds | Fixed income securities |
| 3 | Futures | Futures contracts |
| 4 | Options | Option contracts |
| 5 | Indices | Market indices |
| 6 | Currency | Foreign exchange |
| 7 | Night trading | After-hours trading instruments |
| 8 | REPO Securities | Repurchase agreement securities |
| 9 | Direct REPO | Direct repurchase agreements |
| 10 | Repo with netting | Repo with netting agreements |
| 11 | Bond Yield | Bond yield instruments |
| 14 | Currency Swap | Currency swap agreements |
| 16 | Option expiration | Option expiration events |
| 17 | Option exercise | Option exercise events |
| 18 | Equity swap | Equity swap agreements |
| 19 | Structured products | Structured financial products |
| 20 | Futures expiration | Futures expiration events |

#### Instrument Type Combinations

Specific combinations of instrument types (instr_type_c) and instrument kinds (instr_kind_c):

**Stocks (Type 1):**

| instr_type_c | instr_kind_c | instr_type | instr_kind | Description |
|--------------|--------------|------------|------------|-------------|
| 1 | 1 | Share | ordinary stock | Common stock |
| 1 | 2 | Share | preferred share | Preferred stock |
| 1 | 7 | Share | fund/ETF | Exchange-traded funds |
| 1 | 10 | Share | depositary receipt | ADR/GDR |
| 1 | 14 | Share | crypto shares | Cryptocurrency-related equities |

**Bonds (Type 2):**

| instr_type_c | instr_kind_c | instr_type | instr_kind | Description |
|--------------|--------------|------------|------------|-------------|
| 2 | 1 | Bonds | bond | Corporate/government bonds |
| 2 | 3 | Bonds | bond | Alternative bond type |
| 2 | 9 | Bonds | eurobonds | Eurobonds |
| 2 | 15 | Bonds | bond | Additional bond category |

**Futures (Type 3):**

| instr_type_c | instr_kind_c | instr_type | instr_kind | Description |
|--------------|--------------|------------|------------|-------------|
| 3 | 1 | Futures | futures | Standard futures contracts |
| 3 | 5 | Futures | Deliverable Futures contract | Physical delivery futures |
| 3 | 6 | Futures | Futures settlement | Cash-settled futures |

**Options (Type 4):**

| instr_type_c | instr_kind_c | instr_type | instr_kind | Description |
|--------------|--------------|------------|------------|-------------|
| 4 | 1 | Options | option | Standard options |
| 4 | 5 | Options | settlement option | Cash-settled options |
| 4 | 7 | Options | delivery option | Physical delivery options |

**Indices (Type 5):**

| instr_type_c | instr_kind_c | instr_type | instr_kind | Description |
|--------------|--------------|------------|------------|-------------|
| 5 | 1 | Indices | index | Market indices |
| 5 | 6 | Indices | index contract | Index-based contracts |

**Currency (Type 6):**

| instr_type_c | instr_kind_c | instr_type | instr_kind | Description |
|--------------|--------------|------------|------------|-------------|
| 6 | 1 | Cash | currency | Fiat currencies |
| 6 | 8 | Cash | cryptocurrency | Digital currencies |

**Foreign Exchange Contract (Type 7):**

| instr_type_c | instr_kind_c | instr_type | instr_kind | Description |
|--------------|--------------|------------|------------|-------------|
| 7 | 1 | Foreign exchange contract | currency contract | FX contracts |
| 7 | 2 | Foreign exchange contract | currency swap | Currency swap agreements |

**REPO Securities (Type 8):**

| instr_type_c | instr_kind_c | instr_type | instr_kind | Description |
|--------------|--------------|------------|------------|-------------|
| 8 | 1 | REPO Securities | Autorepo | Automatic repurchase agreements |

**Direct Repo (Type 9):**

| instr_type_c | instr_kind_c | instr_type | instr_kind | Description |
|--------------|--------------|------------|------------|-------------|
| 9 | 1 | Direct repo virtual position | Direct repo | Direct repurchase positions |

**Repo with Netting (Type 10):**

| instr_type_c | instr_kind_c | instr_type | instr_kind | Description |
|--------------|--------------|------------|------------|-------------|
| 10 | 1 | Repo with netting | Repo Netting | Netted repo positions |
| 10 | 9 | Repo with netting | unknown | Unspecified repo netting |

**Bond Yield (Type 11):**

| instr_type_c | instr_kind_c | instr_type | instr_kind | Description |
|--------------|--------------|------------|------------|-------------|
| 11 | 13 | Bond Yield | yield of eurobonds | Eurobond yields |

**Usage Notes:**
- Use `instr_type_c` to filter by major instrument category
- Use `instr_kind_c` to filter by specific instrument subtype
- Combination of both provides precise instrument classification
- Reference these codes when parsing trade history or position data

---

### List of request types

**Purpose:** Reference data for CPS (Client Payment System) document types. Used in `getClientCpsHistory` responses and other CPS-related endpoints.

#### CPS Document Type Codes (cpsDocId)

Complete list of request types and their identifiers:

| cpsDocId | Request Type Name | Category |
|----------|------------------|----------|
| -330 | Issue an investment card (archived) | Legacy |
| 1 | Between accounts | Transfers |
| 3 | Create a bill | Banking |
| 4 | Issue promissory note | Banking |
| 5 | Transfers: bill of exchange redemption | Transfers |
| 7 | Assignment of promissory note | Banking |
| 8 | Transfers: buy a bill by assignment | Transfers |
| 9 | Transfers: via SWIFT | Transfers |
| 12 | Money request | Transfers |
| 13 | P2P | Transfers |
| 14 | Open a deposit card | Banking |
| 15 | Open a deposit card | Banking |
| 16 | Investor card opening | Banking |
| 17 | Request for withdrawal from the deposit | Banking |
| 18 | From another bank's card to client's card | Transfers |
| 19 | Grant access to the product | Access |
| 20 | CVV/CVC display | Banking |
| 21 | Opening a card for kids | Banking |
| 22 | To another bank by phone number | Transfers |
| 23 | Product document display | Documents |
| 24 | Map settings | Settings |
| 25 | Card Alerts | Settings |
| 26 | Transfer to Tsifra Bank | Transfers |
| 27 | Open an Invest Prestige card | Banking |
| 28 | 3D Secure confirmation | Security |
| 29 | View trusted card transactions | Banking |
| 31 | Refusal of trusted space | Access |
| 32 | P2P loan | Loans |
| 33 | Product update | Settings |
| 34 | Open an Armenian broker card | Banking |
| 35 | EDS issue | Security |
| 36 | P2P loan repayment | Loans |
| 37 | Transfer through Mastercard | Transfers |
| 38 | Open an RSA broker card | Banking |
| 39 | Open an account with Global | Banking |
| 40 | Change phone number | Profile |
| 42 | Creating foreign exchange contracts | Trading |
| 44 | Open a deposit card | Banking |
| 45 | Money request | Transfers |
| 46 | Execute an order in TRADERNET | Trading |
| 48 | Close card | Banking |
| 49 | Card top-up through a mobile payment system | Deposits |
| 50 | SWIFT transfer | Transfers |
| 51 | Long-term deposit with an account opened with Global | Banking |
| 52 | Within the Bank | Transfers |
| 53 | Between My Accounts | Transfers |
| 54 | Within the Bank | Transfers |
| 55 | Between My Accounts | Transfers |
| 56 | Within the Bank | Transfers |
| 57 | By phone number in Cifra Bank | Transfers |
| 58 | Opening an invest card | Banking |
| 59 | Setting the primary card for intra-bank transfers | Settings |
| 60 | Removing the primary card for intra-bank transfers | Settings |
| 61 | Service Payment | Payments |
| 62 | FPS transfer from a Cifra Bank product | Transfers |
| 63 | Transfer from a Cifra Bank product to an account with Global | Transfers |
| 64 | Transfer from a Cifra Bank product to an account with TFOS | Transfers |
| 65 | Transfer from a Cifra Bank product to an account with TN Armenia | Transfers |
| 66 | Close Account | Banking |
| 67 | Transfers between your Cifra Bank accounts | Transfers |
| 71 | Card delivery | Banking |
| 73 | Opening a new card | Banking |
| 180 | Cash | Cash Operations |
| 181 | Withdraw | Withdrawals |
| 182 | Personal data | Profile |
| 185 | Account settings | Settings |
| 187 | Transfer of uncovered positions (TUP) | Trading |
| 188 | Security | Security |
| 189 | Voice password | Security |
| 190 | Add token | Security |
| 193 | Authentication tools | Security |
| 194 | Support | Support |
| 195 | Chat with Support | Support |
| 196 | Top up your account via bank transfer | Deposits |
| 197 | Order of documents | Documents |
| 198 | Change service plan | Settings |
| 199 | QUIK key registration | Trading Tools |
| 200 | Transfer between your accounts | Transfers |
| 206 | Deposit funds via a cash office | Deposits |
| 207 | Top up your account with a bank card | Deposits |
| 209 | Activate token | Security |
| 210 | Change Phone Number | Profile |
| 211 | Change email | Profile |
| 213 | Close Account | Banking |
| 214 | Currency conversion | Trading |
| 217 | Change details | Profile |
| 218 | Pay with points | Payments |
| 242 | Securities transfer | Trading |
| 243 | Add the Russian market | Trading |
| 244 | Adding the US market | Trading |
| 266 | What to invest in right now? | Advisory |
| 267 | Postal address | Profile |
| 268 | Recurring deposits | Banking |
| 269 | App authorization (iOS / Android) | Security |
| 270 | Replicator | Trading Tools |
| 271 | Add EDS | Security |
| 273 | Buy stocks at IPO prices | Trading |
| 274 | Register of trades | Documents |
| 275 | Trade Orders | Trading |
| 276 | Vote | Corporate Actions |
| 277 | Cancel request | General |
| 278 | Economic profile | Profile |
| 279 | Custodial account | Banking |
| 281 | Repo prolongation | Trading |
| 283 | Corporate actions | Corporate Actions |
| 287 | Freedom Finance bonds | Trading |
| 288 | Trade order | Trading |
| 289 | Connect QUIK | Trading Tools |
| 290 | Form W-8BEN | Tax/Compliance |
| 292 | DGLB | Trading |
| 295 | API key registration | Security |
| 296 | FIX access connection | Trading Tools |
| 297 | Purchase of securities | Trading |
| 298 | Freedom Finance funds | Trading |
| 318 | Additional account | Banking |
| 319 | Dividend Taxes | Tax/Compliance |
| 320 | DAS platform settings | Trading Tools |
| 326 | Conversion of securities | Trading |
| 327 | Application for BPI participation | Trading |
| 328 | Offer agreement for currency purchase or sale | Trading |
| 329 | Request for qualification of a security | Compliance |
| 330 | Freedom Banker Card | Banking |
| 333 | Tax residency proof | Tax/Compliance |
| 334 | Testing for access to instruments | Compliance |
| 335 | Trade order cancellation | Trading |
| 336 | Source of funds confirmation | Compliance |
| 337 | Cash withdrawal to a bank card | Withdrawals |
| 338 | Provide access to your account | Access |
| 340 | Registration of the buy price of the transferred securities | Trading |
| 341 | Transferring assets from Alfa-Bank | Transfers |
| 343 | Package transfer of assets | Transfers |
| 344 | Affiliate Program Unlimited | Programs |
| 345 | Eurasia card opening | Banking |
| 346 | Currency residence certification | Tax/Compliance |
| 348 | Access to friendly country instruments | Access |
| 349 | Long-term funds placement | Banking |
| 351 | Disabling FIX access | Trading Tools |
| 352 | Changing FIX connection settings | Trading Tools |
| 353 | Consolidated order for a structured product | Trading |
| 355 | Device verification | Security |
| 356 | Appropriateness test | Compliance |
| 357 | Updating the passport photo | Profile |
| 362 | Consent to displaying Cifra Bank accounts | Access |
| 363 | Consent to displaying Skybank accounts | Access |
| 364 | Display of bank accounts in TN | Access |
| 365 | Restoring authentication tools | Security |
| 366 | Opening an account at a WalletSolutions office | Banking |
| 367 | Cash withdrawal to a cryptowallet | Withdrawals |
| 368 | Top up your account with digital assets | Deposits |
| 372 | Faster Payments System | Transfers |
| 373 | Cash withdrawal through the Faster Payments System | Withdrawals |
| 374 | Crypto Card Cifra Markets | Banking |
| 400 | Client reservation | General |
| 1000 | Chat with Support | Support |
| 1006 | Limits | Settings |
| 1007 | Blocking | Security |
| 1008 | Card reissue | Banking |
| 1009 | Set PIN | Banking |
| 1013 | Open a deposit | Banking |
| 1014 | Deposits: settings | Settings |
| 1015 | Within the Bank | Transfers |
| 1016 | Between accounts | Transfers |
| 1018 | To Another Bank | Transfers |
| 1021 | Transfers: payments to service entities | Payments |
| 1027 | Between accounts | Transfers |
| 1028 | Automatically executed payment | Payments |
| 1030 | Request for topping up the deposit | Deposits |
| 1033 | Unlock | Security |
| 1034 | Geo-restrictions | Security |
| 1035 | Add a card to Visa Alias | Banking |
| 1036 | Removing a card from Visa Alias | Banking |
| 1037 | To another bank by card number | Transfers |
| 1038 | Account Opening | Banking |
| 10139 | Pay with VISA or MasterCard | Payments |
| 10159 | Pledged payment | Payments |
| 10160 | Telegram connection management | Settings |
| 10163 | Open a repo | Trading |
| 10167 | Professional investor | Compliance |
| 10168 | Change in the risk level | Settings |
| 10173 | Certificate activation | Security |
| 10174 | Suitability test | Compliance |
| 10175 | Investment advice | Advisory |
| 10177 | Market data subscription | Trading Tools |
| 10180 | Switching to IIS 3 | Banking |
| 10181 | Authorization of clients of agent | Security |
| 11035 | WhatsApp connection management | Settings |
| 11037 | Application for accession | General |

**Usage Notes:**
- Use this reference to decode `type_doc_id` or `cpsDocId` values from `getClientCpsHistory` responses
- Categories help group related request types
- Some IDs represent legacy/archived features (e.g., -330)
- High IDs (10000+) typically represent newer features

---

### List of users profile fields

**Status:** ⏳ AWAITING DOCUMENTATION

---

### Types of documents

**Status:** ⏳ AWAITING DOCUMENTATION

---

### Orders statuses

**Purpose:** Reference data for order status codes. Used in `getNotifyOrderJson`, `getOrdersHistory`, and `putStopLoss` responses.

#### Order Status Codes

Complete list of order statuses:

| Status ID | Status Value | Description |
|-----------|-------------|-------------|
| 0 | was ignored | Order was ignored by the system |
| 1 | received | Order received by broker |
| 2 | processing the cancellation | Order cancellation is being processed |
| 10 | active | Order is active and working |
| 11 | sent | Order sent to exchange |
| 12 | partially completed | Order partially filled |
| 20 | partially performed | Order partially executed (synonym for 12) |
| 21 | executed | Order fully executed |
| 30 | partially canceled | Order partially filled then cancelled |
| 31 | canceled | Order fully cancelled |
| 70 | is rejected | Order rejected by exchange/broker |
| 71 | expired | Order expired (time limit reached) |
| 72 | partially executed and expired | Order partially filled then expired |
| 74 | error of sending | Error occurred while sending order |
| 75 | error of cancellation | Error occurred while cancelling order |

**Status Categories:**

**Pending (Active) Statuses:**
- `1` - received
- `10` - active
- `11` - sent
- `12` - partially completed

**Completed (Final) Statuses:**
- `21` - executed (successful)
- `31` - canceled (cancelled)
- `70` - rejected (failed)
- `71` - expired (timed out)
- `72` - partially executed and expired
- `30` - partially canceled

**Processing Statuses:**
- `2` - processing the cancellation

**Error Statuses:**
- `0` - was ignored
- `74` - error of sending
- `75` - error of cancellation

**Usage Notes:**
- Use `stat` field from order responses to get current status
- Use `stat_orig` field to get the initial status
- Use `stat_prev` field to get the previous status
- Active orders have statuses: 1, 10, 11, 12
- Final statuses (21, 31, 70, 71, 72, 30) indicate order is no longer working
- Monitor status transitions to track order lifecycle

**Common Status Transitions:**
1. New order: `1` (received) → `10` (active) → `11` (sent) → `21` (executed)
2. Partial fill: `1` → `10` → `12` (partially completed) → `21` (executed)
3. Cancelled: `10` → `2` (processing cancellation) → `31` (canceled)
4. Rejected: `1` → `11` → `70` (rejected)
5. Expired: `10` → `71` (expired)

---

### Types of signatures

**Status:** ⏳ AWAITING DOCUMENTATION

---

### Types of valid codes

**Status:** ⏳ AWAITING DOCUMENTATION

---

---

# DOCUMENTATION COLLECTION PROGRESS

**Completed:** 17/~70 REST endpoints + 6 WebSocket subscriptions + 3 reference data sections

**API Endpoints:**
- [x] putTradeOrder - Sending order for execution
- [x] getPositionJson - Getting portfolio information
- [x] getClientCpsHistory - Receiving clients' requests history
- [x] getUserCashFlows - Money funds movement
- [x] delTradeOrder - Cancel the order
- [x] putStopLoss - Stop Loss and Take Profit
- [x] getNotifyOrderJson - Receive orders in current period
- [x] getOrdersHistory - Get orders list for period
- [x] getTradesHistory - Retrieving trades history
- [ ] getStockQuotesJson - Get stock ticker data
- [x] getHloc - Get quote historical data (candlesticks)
- [x] tickerFinder - Stock ticker search
- [x] getMarketStatus - Get market status updates
- [ ] getNews - News on securities
- [ ] getTopSecurities - Getting most traded securities
- [ ] getOptionsByMktNameAndBaseAsset - Options
- [ ] getAlertsList - Get price alerts
- [ ] addPriceAlert - Add price alert
- [ ] Delete price alert
- [ ] getBrokerReport - Receiving broker report
- [ ] getCpsFiles - Receiving order files
- [x] getCrossRatesForDate - Exchange rate by date
- [ ] List of currencies
- [ ] getOPQ - Initial user data
- [x] getSecurityInfo - Instruments details
- [ ] getReadyList - Directory of securities
- [ ] Authentication methods

**WebSocket Subscriptions:**
- [x] WebSocket Connection - Connection setup and authentication
- [x] Subscribe to quotes - Real-time price updates
- [x] Subscribe to market depth - Order book updates
- [x] Subscribe to portfolio - Portfolio and position updates
- [x] Subscribe to orders - Order status updates
- [x] Subscribe to market statuses - Market open/close events
- [ ] Subscribe to security sessions - Security session changes

**Reference Data:**
- [x] Instrument Types Reference Data - Complete type and kind combinations
- [x] List of Request Types - CPS document type codes (cpsDocId)
- [x] Orders Statuses - Order status codes and transitions
- [ ] Types of signatures
- [ ] Types of valid codes
- [ ] List of existing offices
- [ ] Trading platforms
- [ ] Types of documents

---

# NOTES

This document serves as a complete reference for the Tradernet API. As documentation is collected:
- It will be added under the appropriate section
- Cross-references will be maintained
- All examples will be preserved
- TypeScript type definitions will be included where available
- Both success and error responses will be documented

This is a living document that will grow as we collect more API documentation.
