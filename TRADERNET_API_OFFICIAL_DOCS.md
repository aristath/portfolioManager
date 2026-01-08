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

**Status:** ⏳ AWAITING DOCUMENTATION

---

### Stock ticker search - `tickerFinder`

**Status:** ⏳ AWAITING DOCUMENTATION

---

### Get updates on market status - `getMarketStatus`

**Status:** ⏳ AWAITING DOCUMENTATION

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

### Adding ticker to list

**Status:** ⏳ AWAITING DOCUMENTATION

---

### Deleting ticker from list

**Status:** ⏳ AWAITING DOCUMENTATION

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

**Status:** ⏳ AWAITING DOCUMENTATION

---

### Subscribe to quotes (WebSocket)

**Status:** ⏳ AWAITING DOCUMENTATION

---

### Subscribe to market depth (WebSocket)

**Status:** ⏳ AWAITING DOCUMENTATION

---

### Subscribing to security sessions (WebSocket)

**Status:** ⏳ AWAITING DOCUMENTATION

---

### Subscribe to portfolio (WebSocket)

**Status:** ⏳ AWAITING DOCUMENTATION

---

### Subscribe to orders (WebSocket)

**Status:** ⏳ AWAITING DOCUMENTATION

---

### Subscribing to market statuses (WebSocket)

**Status:** ⏳ AWAITING DOCUMENTATION

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

**Completed:** 12/~70 endpoints + 3 reference data sections

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
- [ ] getHloc - Get quote historical data
- [ ] tickerFinder - Stock ticker search
- [ ] getMarketStatus - Get market status updates
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
- [ ] WebSocket endpoints

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
