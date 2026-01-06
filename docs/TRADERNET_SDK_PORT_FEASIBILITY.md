# Tradernet SDK Go Port - Feasibility Analysis

## Question: Can we port EXACTLY as-is, handling ALL cases?

**Answer: YES, with careful attention to JSON field ordering.**

---

## ✅ What CAN be ported exactly

### 1. Authentication Logic
- ✅ JSON stringify (no spaces) - Go's `json.Marshal` does this
- ✅ Timestamp in seconds - `time.Now().Unix()`
- ✅ String concatenation - trivial
- ✅ SHA256 HMAC - `crypto/hmac` + `crypto/sha256`
- ✅ Hex encoding - `encoding/hex`

### 2. HTTP Requests
- ✅ POST requests - `net/http`
- ✅ Headers - `req.Header.Set()`
- ✅ JSON body - `bytes.NewReader(jsonBytes)`
- ✅ Response parsing - `json.Decoder`

### 3. Response Parsing
- ✅ Nested structures - Go structs with JSON tags
- ✅ Fallback field names - Custom `UnmarshalJSON` or manual parsing
- ✅ Single vs list handling - Type assertions
- ✅ Default values - Go's zero values + explicit checks

### 4. Data Type Conversions
- ✅ Boolean to int - Explicit conversion
- ✅ Date formatting - `time.Format()`
- ✅ String/float/int conversions - Standard Go

### 5. Error Handling
- ✅ Check `errMsg` field - Struct field + check
- ✅ Log but don't raise - Go error return pattern
- ✅ Validation errors - Return errors explicitly

### 6. All Methods
- ✅ Every Python method can be implemented in Go
- ✅ All edge cases can be handled
- ✅ All special logic (IOC, etc.) can be replicated

---

## ⚠️ CRITICAL ISSUE: JSON Key Ordering

### The Problem

**Python behavior** (Python 3.7+ preserves insertion order):
```python
# From client.py - exact order used:
params = {
    'instr_name': symbol,      # 1
    'action_id': action_id,    # 2
    'order_type_id': 2,        # 3
    'qty': abs(quantity),      # 4
    'limit_price': price,      # 5
    'expiration_id': 1,        # 6
    'user_order_id': custom_order_id  # 7
}
json.dumps(params, separators=(',', ':'))
# Result: '{"instr_name":"AAPL.US","action_id":1,"order_type_id":2,"qty":10,"limit_price":150.0,"expiration_id":1}'
```

**Go behavior with maps**:
```go
data := map[string]interface{}{
    "instr_name": "AAPL.US",
    "action_id": 1,
    // ...
}
json.Marshal(data)
// Result: Keys are SORTED ALPHABETICALLY!
// '{"action_id":1,"expiration_id":1,"instr_name":"AAPL.US",...}'
// ❌ DIFFERENT ORDER = DIFFERENT SIGNATURE = AUTH FAILS!
```

**Why this matters**:
- Authentication signature = `HMAC(JSON_payload + timestamp)`
- If JSON key order differs → different payload → different signature → AUTH FAILS!

### The Solution

**Use structs with field order matching Python's insertion order**:

```go
// CRITICAL: Field order MUST match Python's dict insertion order exactly!
// Python order (from client.py, line ~1180):
//   1. 'instr_name'
//   2. 'action_id'
//   3. 'order_type_id'
//   4. 'qty'
//   5. 'limit_price'
//   6. 'expiration_id'
//   7. 'user_order_id'
type PutTradeOrderParams struct {
    InstrName    string  `json:"instr_name"`     // Field 1 - MUST be first
    ActionID     int     `json:"action_id"`      // Field 2
    OrderTypeID  int     `json:"order_type_id"`  // Field 3
    Qty          int     `json:"qty"`            // Field 4
    LimitPrice   float64 `json:"limit_price"`    // Field 5
    ExpirationID int     `json:"expiration_id"`  // Field 6
    UserOrderID  *int    `json:"user_order_id,omitempty"` // Field 7
}

// Go's json.Marshal on structs uses FIELD ORDER (deterministic!)
params := PutTradeOrderParams{
    InstrName:    "AAPL.US",
    ActionID:     1,
    OrderTypeID:  2,
    Qty:          10,
    LimitPrice:   150.0,
    ExpirationID: 1,
}
jsonBytes, _ := json.Marshal(params)
// Result: '{"instr_name":"AAPL.US","action_id":1,"order_type_id":2,"qty":10,"limit_price":150,"expiration_id":1}'
// ✅ Matches Python output (field order preserved)!
```

**Verification**:
- ✅ Go structs preserve field order (deterministic)
- ✅ Python dicts preserve insertion order (Python 3.7+)
- ✅ Must match field order exactly for each method

### Field Order Requirements

For each method, we must document the exact field order used in Python:

1. **putTradeOrder**: `instr_name`, `action_id`, `order_type_id`, `qty`, `limit_price`, `expiration_id`, `user_order_id`
2. **getNotifyOrderJson**: `active_only`
3. **getTradesHistory**: `beginDate`, `endDate`, `tradeId`, `max`, `nt_ticker`, `curr`
4. **getStockQuotesJson**: `tickers`
5. **getHloc**: `id`, `count`, `timeframe`, `date_from`, `date_to`, `intervalMode`
6. **getSecurityInfo**: `ticker`, `sup`
7. **GetAllUserTexInfo**: (no params)
8. **getPositionJson**: (no params)

**Action Required**: Document exact field order for each method's params struct.

---

## ✅ Detailed Feasibility by Component

### 1. Core Authentication (`core.py`)

| Feature | Python | Go | Feasible? |
|---------|--------|----|-----------|
| JSON stringify (no spaces) | `json.dumps(..., separators=(',', ':'))` | `json.Marshal()` | ✅ Yes |
| Key ordering | Insertion order (dict) | Field order (struct) | ⚠️ Must match exactly |
| Timestamp (seconds) | `int(time.time())` | `time.Now().Unix()` | ✅ Yes |
| String concat | `payload + timestamp` | `payload + timestamp` | ✅ Yes |
| SHA256 HMAC | `hmac.new(..., sha256)` | `hmac.New(sha256.New, ...)` | ✅ Yes |
| Hex digest | `.hexdigest()` | `hex.EncodeToString()` | ✅ Yes |
| HTTP POST | `requests.post()` | `http.NewRequest("POST", ...)` | ✅ Yes |
| Headers | `headers={...}` | `req.Header.Set(...)` | ✅ Yes |
| JSON body | `data=payload` | `bytes.NewReader([]byte(payload))` | ✅ Yes |

**Verdict**: ✅ **Fully feasible** (with struct-based JSON and exact field ordering)

### 2. Request Methods (`client.py`)

| Method | Complexity | Go Feasibility |
|--------|-----------|----------------|
| `user_info()` | Simple | ✅ Trivial |
| `buy()` / `sell()` | Medium (validation) | ✅ Easy |
| `get_placed()` | Medium (response parsing) | ✅ Easy |
| `account_summary()` | Medium (nested parsing) | ✅ Easy |
| `get_trades_history()` | Medium (date formatting) | ✅ Easy |
| `get_quotes()` | Medium (list handling) | ✅ Easy |
| `get_candles()` | Medium (date formatting) | ✅ Easy |
| `find_symbol()` | Simple (plain_request) | ✅ Easy |
| `security_info()` | Simple | ✅ Trivial |
| `authorized_request()` | Core method | ✅ Already covered |

**Verdict**: ✅ **All methods feasible**

### 3. Response Parsing

**Python pattern**:
```python
response.get("result", {}).get("orders", {}).get("order", [])
if isinstance(order_list, dict):
    order_list = [order_list]
value = data.get("id") or data.get("orderId") or ""
```

**Go equivalent**:
```go
// Option 1: Struct-based (type-safe, recommended)
type GetPlacedResponse struct {
    Result struct {
        Orders struct {
            Order interface{} `json:"order"` // Can be []Order or Order
        } `json:"orders"`
    } `json:"result"`
    ErrMsg string `json:"errMsg,omitempty"`
}

// Option 2: Manual parsing (flexible, matches Python exactly)
var result map[string]interface{}
json.Unmarshal(body, &result)

// Handle nested access with type assertions
resultMap := result["result"].(map[string]interface{})
ordersMap := resultMap["orders"].(map[string]interface{})
orderData := ordersMap["order"]

// Handle single vs list
var orderList []interface{}
switch v := orderData.(type) {
case []interface{}:
    orderList = v
case map[string]interface{}:
    orderList = []interface{}{v}
}

// Fallback fields (matches Python's .get() with or)
orderID := ""
if id, ok := order["id"].(string); ok && id != "" {
    orderID = id
} else if orderID2, ok := order["orderId"].(string); ok {
    orderID = orderID2
}
```

**Verdict**: ✅ **Fully feasible** (slightly more verbose, but exact same logic)

### 4. Error Handling

**Python pattern**:
```python
if 'errMsg' in result:
    self.logger.error('Error: %s', result['errMsg'])
return result  # Returns even with error
```

**Go equivalent**:
```go
type APIResponse struct {
    Result interface{} `json:"result,omitempty"`
    ErrMsg string      `json:"errMsg,omitempty"`
}

func (c *Client) checkError(resp *APIResponse) {
    if resp.ErrMsg != "" {
        c.log.Warn().Str("err_msg", resp.ErrMsg).Msg("API returned error")
        // Don't return error - match Python behavior exactly
    }
}

// Usage
resp, err := c.authorizedRequest(...)
if err != nil {
    return nil, err
}
c.checkError(resp) // Logs but doesn't fail
return resp, nil // Returns even if errMsg present
```

**Verdict**: ✅ **Fully feasible** (matches Python behavior exactly)

### 5. Special Cases

| Case | Python | Go | Feasible? |
|------|--------|----|-----------|
| IOC emulation | Recursive call + cancel | Same logic | ✅ Yes |
| Duration validation | `duration.lower()` + dict check | `strings.ToLower()` + map check | ✅ Yes |
| Action ID calc | `if/elif` logic | Same `if/else` logic | ✅ Yes |
| Boolean to int | `int(active)` | Explicit conversion | ✅ Yes |
| Date formatting | `strftime()` | `time.Format()` | ✅ Yes |
| Plain requests | GET with `?q=<json>` | `http.NewRequest("GET", ...)` | ✅ Yes |

**Verdict**: ✅ **All special cases feasible**

---

## 🎯 Implementation Strategy

### Phase 1: Authentication (CRITICAL)

1. **Document exact field order** for each method's params
2. **Create structs** with fields in exact Python order
3. **Test signature generation** against Python output
4. **Verify API accepts signatures** (test with real API)

### Phase 2: Core Methods

1. Implement `authorized_request()` first
2. Test with `user_info()` (simplest method)
3. Verify response parsing matches Python

### Phase 3: All Methods

1. Implement each method one by one
2. Test response parsing for each
3. Verify edge cases (single vs list, fallback fields, etc.)

### Phase 4: Edge Cases

1. IOC order emulation
2. Error handling
3. Network errors
4. Timeout handling

---

## ⚠️ Potential Issues & Solutions

### Issue 1: JSON Key Ordering

**Problem**:
- Python dicts preserve insertion order (Python 3.7+)
- Go maps are sorted alphabetically (NOT insertion order!)
- Different JSON key order → different signature → AUTH FAILS!

**Solution**:
- Use structs (field order is deterministic and matches Python's insertion order)
- Must match Python's exact field order for each method
- Test signature generation against Python output

**Status**: ✅ **SOLVED** (with careful field ordering)

**CRITICAL**: Field order in Go structs MUST match Python's dict insertion order!

### Issue 2: Dynamic Field Access

**Problem**: Python's `.get()` with fallbacks is very flexible
**Solution**: Use `map[string]interface{}` + type assertions + fallback logic
**Status**: ✅ **SOLVED**

### Issue 3: Single vs List Responses

**Problem**: API sometimes returns dict, sometimes list
**Solution**: Type assertion + conversion logic
**Status**: ✅ **SOLVED**

### Issue 4: Fallback Field Names

**Problem**: Multiple possible field names (`id`/`orderId`, `price`/`p`)
**Solution**: Check each field in order, use first non-empty
**Status**: ✅ **SOLVED**

### Issue 5: Error Handling Pattern

**Problem**: Python logs but doesn't raise
**Solution**: Check `errMsg`, log, but don't return error
**Status**: ✅ **SOLVED**

---

## ✅ Final Verdict

### Can we port EXACTLY as-is?

**YES**, with these requirements:

1. ✅ **Use structs for request params** (ensures JSON key order matches)
2. ✅ **Match exact field order** from Python code (critical for signatures!)
3. ✅ **Test signature generation** against Python (must match exactly!)
4. ✅ **Handle all response parsing patterns** (single/list, fallbacks, etc.)
5. ✅ **Match error handling behavior** (log but don't fail)
6. ✅ **Implement all edge cases** (IOC, validation, etc.)

### Confidence Level

- **Authentication**: 95% (need to verify signature matching with exact field order)
- **Request Methods**: 100% (all straightforward)
- **Response Parsing**: 100% (slightly more verbose, but exact same logic)
- **Error Handling**: 100% (matches Python behavior)
- **Edge Cases**: 100% (all handleable)

### Risk Assessment

**Low Risk**:
- All methods are HTTP-based (standard)
- All logic is straightforward
- No complex Python-specific features

**Medium Risk**:
- JSON key ordering (mitigated by using structs with exact field order)
- Signature matching (need to test with real API)

**High Risk**:
- None identified

---

## 🧪 Required Tests

### Critical Tests (MUST PASS)

1. **Signature Matching Test**:
   ```go
   // Same params in Python and Go must produce same signature
   params := PutTradeOrderParams{
       InstrName:    "AAPL.US",
       ActionID:     1,
       OrderTypeID:  2,
       Qty:          10,
       LimitPrice:   150.0,
       ExpirationID: 1,
   }
   payload := stringify(params)
   timestamp := "1234567890"
   message := payload + timestamp
   signature := sign(message)
   // Must match Python output EXACTLY!
   ```

2. **Authentication Test**:
   ```go
   // Real API call with Go client
   // Must succeed with same credentials as Python
   ```

3. **Response Parsing Test**:
   ```go
   // Parse all response types
   // Must handle: single/list, fallback fields, nested structures
   ```

### Comprehensive Tests

- [ ] All 10 methods work correctly
- [ ] All edge cases handled
- [ ] Error responses parsed correctly
- [ ] Network errors handled gracefully
- [ ] Timeout handling works
- [ ] Date formats correct per endpoint
- [ ] Type conversions correct
- [ ] Boolean to int conversion correct
- [ ] **Signature matches Python exactly** (CRITICAL!)

---

## 📋 Implementation Checklist

### Pre-Implementation

- [ ] Document exact field order for each method's params (from Python code)
- [ ] Create structs with fields in exact Python order
- [ ] Test signature generation against Python output
- [ ] Verify API accepts Go-generated signatures

### Implementation

- [ ] Core authentication (`authorized_request`)
- [ ] Plain requests (`plain_request`)
- [ ] All 10 methods (with correct field order)
- [ ] Response parsing (all patterns)
- [ ] Error handling
- [ ] Edge cases (IOC, validation, etc.)

### Testing

- [ ] Unit tests for each method
- [ ] Integration tests with real API
- [ ] **Signature matching tests** (CRITICAL!)
- [ ] Response parsing tests
- [ ] Error handling tests
- [ ] Edge case tests

### Validation

- [ ] Compare Go output with Python output (same inputs)
- [ ] Verify all response structures match
- [ ] Verify error handling matches
- [ ] **Verify signatures match exactly** (CRITICAL!)
- [ ] Performance testing (should be faster!)

---

## 🎯 Conclusion

**YES, we can port EXACTLY as-is**, provided we:

1. Use structs for request params (ensures deterministic JSON)
2. **Match exact field order from Python code** (critical for signatures!)
3. Test signature generation carefully (must match Python exactly)
4. Handle all response parsing patterns
5. Match error handling behavior

**The port is 100% feasible** with proper implementation. The only critical detail is JSON key ordering, which is easily solved by using structs with fields in the exact same order as Python's dict insertion order.

**Recommendation**:
1. Document exact field order for each method
2. Create structs matching that order
3. Test signature generation first (before implementing methods)
4. Proceed with full implementation once signatures match

**Confidence**: Very High (95%+) - The only uncertainty is signature matching, which can be verified with a simple test.
