# Tradernet SDK - Missing Methods Analysis

## Summary

After comparing the Go implementation against the Python SDK 2.0.0, here are the methods that exist in the Python SDK but are **NOT currently implemented** in our Go version.

---

## ✅ Implemented Methods (13)

1. ✅ `user_info()` → `UserInfo()`
2. ✅ `account_summary()` → `AccountSummary()`
3. ✅ `buy()` → `Buy()`
4. ✅ `sell()` → `Sell()`
5. ✅ `trade()` → `Trade()` (internal)
6. ✅ `get_placed()` → `GetPlaced()`
7. ✅ `get_trades_history()` → `GetTradesHistory()`
8. ✅ `get_quotes()` → `GetQuotes()`
9. ✅ `get_candles()` → `GetCandles()`
10. ✅ `find_symbol()` → `FindSymbol()`
11. ✅ `security_info()` → `SecurityInfo()`
12. ✅ `get_requests_history()` → `GetClientCpsHistory()`
13. ✅ `cancel()` → `Cancel()`

---

## ❌ Missing Methods (24)

### User Management & Registration
1. **`new_user()`** - Create new user account
   - Uses: `plain_request('registerNewUser', ...)`
   - Purpose: User registration
   - **Used by microservice?**: ❌ NO

2. **`check_missing_fields()`** - Check missing profile fields
   - Uses: `authorized_request('checkStep', ...)`
   - Purpose: Validate profile completion
   - **Used by microservice?**: ❌ NO

3. **`get_profile_fields()`** - Get profile fields for offices
   - Uses: `authorized_request('getAnketaFields', ...)`
   - Purpose: Get profile form fields
   - **Used by microservice?**: ❌ NO

4. **`get_user_data()`** - Get initial user data
   - Uses: `authorized_request('getOPQ', ...)`
   - Purpose: Get orders, portfolio, markets, sessions
   - **Used by microservice?**: ❌ NO

### Market Data & Securities
5. **`get_market_status()`** - Get market status
   - Uses: `authorized_request('getMarketStatus', ...)`
   - Purpose: Market operation status
   - **Used by microservice?**: ❌ NO

6. **`get_options()`** - Get options by underlying
   - Uses: `authorized_request('getOptionsByMktNameAndBaseAsset', ...)`
   - Purpose: List options for underlying asset
   - **Used by microservice?**: ❌ NO

7. **`get_most_traded()`** - Get most traded securities
   - Uses: `plain_request('getTopSecurities', ...)`
   - Purpose: Top gainers or most traded
   - **Used by microservice?**: ❌ NO

8. **`export_securities()`** - Export securities data
   - Uses: Direct HTTP GET to `/securities/export`
   - Purpose: Bulk export of security data
   - **Used by microservice?**: ❌ NO

9. **`get_news()`** - Get news on securities
   - Uses: `authorized_request('getNews', ...)`
   - Purpose: News feed for symbols
   - **Used by microservice?**: ❌ NO

10. **`get_all()`** - Get all securities with filters
    - Uses: Internal refbook parsing
    - Purpose: Filter securities by criteria
    - **Used by microservice?**: ❌ NO

11. **`symbol()`** - Get stock data (different from security_info)
    - Uses: `authorized_request('getStockData', ...)`
    - Purpose: Shop/display data for symbol
    - **Used by microservice?**: ❌ NO

12. **`symbols()`** - Get ready list of securities
    - Uses: `authorized_request('getReadyList', ...)`
    - Purpose: Complete list of securities by exchange
    - **Used by microservice?**: ❌ NO

13. **`corporate_actions()`** - Get corporate actions
    - Uses: `authorized_request('getPlannedCorpActions', ...)`
    - Purpose: Planned corporate actions
    - **Used by microservice?**: ❌ NO

### Price Alerts
14. **`get_price_alerts()`** - Get price alerts
    - Uses: `authorized_request('getAlertsList', ...)`
    - Purpose: List user's price alerts
    - **Used by microservice?**: ❌ NO

15. **`add_price_alert()`** - Add price alert
    - Uses: `authorized_request('addPriceAlert', ...)`
    - Purpose: Create price alert
    - **Used by microservice?**: ❌ NO

16. **`delete_price_alert()`** - Delete price alert
    - Uses: `authorized_request('addPriceAlert', {'id': ..., 'del': True})`
    - Purpose: Remove price alert
    - **Used by microservice?**: ❌ NO

### Advanced Trading
17. **`stop()`** - Place stop loss order
    - Uses: `authorized_request('putStopLoss', {'instr_name': ..., 'stop_loss': ...})`
    - Purpose: Stop loss on position
    - **Used by microservice?**: ❌ NO

18. **`trailing_stop()`** - Place trailing stop
    - Uses: `authorized_request('putStopLoss', {'instr_name': ..., 'stop_loss_percent': ..., 'stoploss_trailing_percent': ...})`
    - Purpose: Trailing stop loss
    - **Used by microservice?**: ❌ NO

19. **`take_profit()`** - Place take profit order
    - Uses: `authorized_request('putStopLoss', {'instr_name': ..., 'take_profit': ...})`
    - Purpose: Take profit on position
    - **Used by microservice?**: ❌ NO

20. **`cancel_all()`** - Cancel all orders
    - Uses: `get_placed()` + `cancel()` loop
    - Purpose: Cancel all active orders
    - **Used by microservice?**: ❌ NO

### Orders & Reports
21. **`get_historical()`** - Get orders history
    - Uses: `authorized_request('getOrdersHistory', ...)`
    - Purpose: Historical orders (different from trades)
    - **Used by microservice?**: ❌ NO

22. **`get_order_files()`** - Get order files
    - Uses: `authorized_request('getCpsFiles', ...)`
    - Purpose: Download order documents
    - **Used by microservice?**: ❌ NO

23. **`get_broker_report()`** - Get broker report
    - Uses: `authorized_request('getBrokerReport', ...)`
    - Purpose: Broker's report
    - **Used by microservice?**: ❌ NO

### Other
24. **`get_tariffs_list()`** - Get tariffs list
    - Uses: `authorized_request('GetListTariffs', ...)`
    - Purpose: Available tariff plans
    - **Used by microservice?**: ❌ NO

---

## 📊 Analysis

### Methods Used by Your Microservice
Based on `tradernet_service.py`, your microservice uses:
- ✅ `user_info()` - Used
- ✅ `account_summary()` - Used
- ✅ `buy()` / `sell()` - Used
- ✅ `get_placed()` - Used
- ✅ `get_trades_history()` - Used
- ✅ `get_quotes()` - Used
- ✅ `get_candles()` - Used
- ✅ `find_symbol()` - Used
- ✅ `security_info()` - Used
- ✅ `authorized_request('getClientCpsHistory', ...)` - Used

**Result**: All methods used by your microservice are **already implemented** ✅

### Methods NOT Used by Your Microservice
All 24 missing methods are **NOT used** by your current microservice implementation.

---

## 🎯 Recommendation

### Option 1: Implement Only What's Needed (Recommended)
**Status**: ✅ **COMPLETE** - All methods used by your microservice are implemented.

The missing methods are for:
- User registration/management (not needed for trading bot)
- Price alerts (not used)
- Advanced order types (stop loss, trailing stop, take profit - not used)
- Market data exploration (not used)
- News feeds (not used)
- Corporate actions (not used)
- Reports and files (not used)

### Option 2: Implement All Methods (Future-Proofing)
If you want complete feature parity, you could implement the remaining 24 methods. However, this is **not necessary** for your current use case.

**Priority for future implementation** (if needed):
1. **High Priority** (if you add features):
   - `stop()`, `trailing_stop()`, `take_profit()` - Advanced order types
   - `cancel_all()` - Convenience method
   - `get_historical()` - Order history (different from trades)

2. **Medium Priority** (nice to have):
   - `get_market_status()` - Market status checks
   - `get_news()` - News integration
   - `symbols()` - Security listing

3. **Low Priority** (unlikely to need):
   - User registration methods
   - Price alerts
   - Corporate actions
   - Reports and files

---

## ✅ Conclusion

**Your Go implementation is COMPLETE for your current use case.**

All methods used by your Python microservice are implemented and working. The 24 missing methods are not used by your application and are not needed for your trading bot functionality.

**Recommendation**: Keep the current implementation. Add missing methods only if you need them for new features.
