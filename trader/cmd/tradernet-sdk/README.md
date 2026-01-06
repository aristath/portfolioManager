# Tradernet SDK Go Microservice

Standalone Go microservice implementing the Tradernet SDK for testing and validation.

## Building

```bash
cd trader
go build -o tradernet-sdk ./cmd/tradernet-sdk
```

## Running

```bash
./tradernet-sdk
```

The server will start on port 9001 (or the port specified in `TRADERNET_SDK_PORT` environment variable).

## Testing

### Unit Tests

```bash
cd trader
go test ./internal/clients/tradernet/sdk ./cmd/tradernet-sdk -v
```

### Integration Test

**Option 1: Automated Script**
```bash
cd trader/cmd/tradernet-sdk
./test_integration.sh
```

**Option 2: Manual Testing**

1. Build the binary:
   ```bash
   cd trader
   go build -o tradernet-sdk ./cmd/tradernet-sdk
   ```

2. Start the server:
   ```bash
   ./tradernet-sdk
   ```

3. Test with real API credentials (from `.env.aristath`):
   ```bash
   source .env.aristath
   curl -H "X-Tradernet-API-Key: $TRADERNET_API_KEY" \
        -H "X-Tradernet-API-Secret: $TRADERNET_API_SECRET" \
        http://localhost:9001/user-info
   ```

4. Verify the response matches the Python SDK output.

## Endpoints

Complete endpoint documentation is available in [ENDPOINTS.md](./ENDPOINTS.md).

**Quick Reference:**

### Health & Status
- `GET /health` - Health check (no auth)

### User & Account
- `GET /user-info` - Get user information
- `GET /account-summary` - Get account summary (positions, cash)
- `GET /user-data` - Get initial user data (orders, portfolio, markets)

### Trading
- `POST /buy` - Place buy order
- `POST /sell` - Place sell order
- `GET /pending-orders` - Get pending/active orders
- `POST /cancel-order` - Cancel an order
- `POST /cancel-all` - Cancel all orders
- `POST /stop` - Place stop loss order
- `POST /trailing-stop` - Place trailing stop order
- `POST /take-profit` - Place take profit order
- `GET /orders-history` - Get orders history

### Transactions & Reports
- `GET /trades-history` - Get executed trades history
- `GET /cash-movements` - Get cash movements history
- `GET /order-files` - Get order files
- `GET /broker-report` - Get broker report

### Market Data
- `POST /quotes` - Get quotes for symbols
- `POST /candles` - Get historical OHLC data
- `GET /market-status` - Get market status
- `GET /most-traded` - Get most traded securities (no auth)
- `POST /export-securities` - Export securities data
- `GET /news` - Get news on securities

### Securities
- `GET /find-symbol` - Find security by symbol/ISIN (no auth)
- `GET /security-info` - Get security information
- `GET /symbol` - Get stock data
- `GET /symbols` - Get ready list of securities
- `GET /options` - Get options by underlying
- `GET /corporate-actions` - Get corporate actions
- `POST /get-all` - Get all securities (not yet implemented)

### Price Alerts
- `GET /price-alerts` - Get price alerts
- `POST /add-price-alert` - Add price alert
- `POST /delete-price-alert` - Delete price alert

### User Management
- `POST /new-user` - Create new user (no auth)
- `GET /check-missing-fields` - Check missing profile fields
- `GET /profile-fields` - Get profile fields

### Other
- `GET /tariffs-list` - Get tariffs list

See [ENDPOINTS.md](./ENDPOINTS.md) for complete documentation with request/response examples.

## Credentials

Credentials are passed via HTTP headers:
- `X-Tradernet-API-Key` - Public API key
- `X-Tradernet-API-Secret` - Private API secret

The microservice does NOT read `.env` files - it's stateless and accepts credentials per request.
