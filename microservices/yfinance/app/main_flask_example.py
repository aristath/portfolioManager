"""Yahoo Finance microservice - Flask example (for comparison).

This is an example showing how the service would look using Flask instead of FastAPI.
This file is for reference only - not used in production.

Note: Since yfinance calls are synchronous, Flask works perfectly fine.
"""

import json
import logging
from datetime import datetime
from typing import Optional

from app.config import settings
from app.service import get_yahoo_finance_service
from flask import Flask, jsonify, request
from flask_cors import CORS

# Configure logging
logging.basicConfig(
    level=settings.log_level,
    format="%(asctime)s - %(name)s - %(levelname)s - %(message)s",
)
logger = logging.getLogger(__name__)

# Create Flask app
app = Flask(__name__)

# Enable CORS
CORS(app)

# Get service instance
service = get_yahoo_finance_service()


# Helper function for standard responses
def success_response(data: dict):
    """Create a success response."""
    return jsonify(
        {
            "success": True,
            "data": data,
            "timestamp": datetime.utcnow().isoformat(),
        }
    )


def error_response(error: str, status_code: int = 200):
    """Create an error response."""
    return (
        jsonify(
            {
                "success": False,
                "error": error,
                "timestamp": datetime.utcnow().isoformat(),
            }
        ),
        status_code,
    )


# Health check endpoint
@app.route("/health", methods=["GET"])
def health_check():
    """Health check endpoint."""
    return jsonify(
        {
            "status": "healthy",
            "service": settings.service_name,
            "version": settings.version,
            "timestamp": datetime.utcnow().isoformat(),
        }
    )


# Quote endpoints
@app.route("/api/quotes/<symbol>", methods=["GET"])
def get_quote(symbol: str):
    """Get current price for a symbol."""
    try:
        yahoo_symbol = request.args.get("yahoo_symbol")

        price = service.get_current_price(symbol, yahoo_symbol)
        if price is not None:
            return success_response({"symbol": symbol, "price": price})
        return error_response(f"No price available for {symbol}")
    except Exception as e:
        logger.exception(f"Error getting quote for {symbol}")
        return error_response(str(e))


@app.route("/api/quotes/batch", methods=["POST"])
def get_batch_quotes():
    """Get current prices for multiple symbols."""
    try:
        body = request.get_json()
        if not body:
            return error_response("Request body must be JSON", status_code=400)

        symbols = body.get("symbols", [])
        yahoo_overrides = body.get("yahoo_overrides")

        if not symbols:
            return error_response("symbols field is required", status_code=400)

        quotes = service.get_batch_quotes(symbols, yahoo_overrides)
        return success_response({"quotes": quotes})
    except Exception as e:
        logger.exception("Error getting batch quotes")
        return error_response(str(e))


# Historical data endpoints
@app.route("/api/historical", methods=["POST"])
def get_historical_prices_post():
    """Get historical OHLCV data (POST endpoint)."""
    try:
        body = request.get_json()
        if not body:
            return error_response("Request body must be JSON", status_code=400)

        symbol = body.get("symbol")
        yahoo_symbol = body.get("yahoo_symbol")
        period = body.get("period", "1y")
        interval = body.get("interval", "1d")

        if not symbol:
            return error_response("symbol field is required", status_code=400)

        prices = service.get_historical_prices(symbol, yahoo_symbol, period, interval)

        # Convert Pydantic models to dict for JSON serialization
        prices_dict = [price.dict() if hasattr(price, "dict") else price for price in prices]

        return success_response(
            {
                "symbol": symbol,
                "prices": prices_dict,
            }
        )
    except Exception as e:
        logger.exception("Error getting historical prices")
        return error_response(str(e))


@app.route("/api/historical/<symbol>", methods=["GET"])
def get_historical_prices_get(symbol: str):
    """Get historical OHLCV data (GET endpoint)."""
    try:
        yahoo_symbol = request.args.get("yahoo_symbol")
        period = request.args.get("period", "1y")
        interval = request.args.get("interval", "1d")

        prices = service.get_historical_prices(symbol, yahoo_symbol, period, interval)

        # Convert Pydantic models to dict for JSON serialization
        prices_dict = [price.dict() if hasattr(price, "dict") else price for price in prices]

        return success_response(
            {
                "symbol": symbol,
                "prices": prices_dict,
            }
        )
    except Exception as e:
        logger.exception(f"Error getting historical prices for {symbol}")
        return error_response(str(e))


# Fundamental data endpoints
@app.route("/api/fundamentals/<symbol>", methods=["GET"])
def get_fundamentals(symbol: str):
    """Get fundamental analysis data."""
    try:
        yahoo_symbol = request.args.get("yahoo_symbol")

        data = service.get_fundamental_data(symbol, yahoo_symbol)
        if data:
            return success_response(data.dict() if hasattr(data, "dict") else data)
        return error_response(f"No fundamental data available for {symbol}")
    except Exception as e:
        logger.exception(f"Error getting fundamentals for {symbol}")
        return error_response(str(e))


# Analyst data endpoints
@app.route("/api/analyst/<symbol>", methods=["GET"])
def get_analyst_data(symbol: str):
    """Get analyst recommendations and price targets."""
    try:
        yahoo_symbol = request.args.get("yahoo_symbol")

        data = service.get_analyst_data(symbol, yahoo_symbol)
        if data:
            return success_response(data.dict() if hasattr(data, "dict") else data)
        return error_response(f"No analyst data available for {symbol}")
    except Exception as e:
        logger.exception(f"Error getting analyst data for {symbol}")
        return error_response(str(e))


# Security info endpoints
@app.route("/api/security/industry/<symbol>", methods=["GET"])
def get_security_industry(symbol: str):
    """Get security industry/sector."""
    try:
        yahoo_symbol = request.args.get("yahoo_symbol")

        data = service.get_security_industry(symbol, yahoo_symbol)
        if data:
            return success_response(data.dict() if hasattr(data, "dict") else data)
        return error_response(f"No industry data available for {symbol}")
    except Exception as e:
        logger.exception(f"Error getting industry for {symbol}")
        return error_response(str(e))


@app.route("/api/security/country-exchange/<symbol>", methods=["GET"])
def get_security_country_exchange(symbol: str):
    """Get security country and exchange."""
    try:
        yahoo_symbol = request.args.get("yahoo_symbol")

        data = service.get_security_country_exchange(symbol, yahoo_symbol)
        if data:
            return success_response(data.dict() if hasattr(data, "dict") else data)
        return error_response(f"No country/exchange data available for {symbol}")
    except Exception as e:
        logger.exception(f"Error getting country/exchange for {symbol}")
        return error_response(str(e))


@app.route("/api/security/info/<symbol>", methods=["GET"])
def get_security_info(symbol: str):
    """Get comprehensive security information."""
    try:
        yahoo_symbol = request.args.get("yahoo_symbol")

        data = service.get_security_info(symbol, yahoo_symbol)
        if data:
            return success_response(data.dict() if hasattr(data, "dict") else data)
        return error_response(f"No security info available for {symbol}")
    except Exception as e:
        logger.exception(f"Error getting security info for {symbol}")
        return error_response(str(e))


# Flask doesn't have built-in startup/shutdown events like FastAPI
# But you can use @app.before_first_request or @app.before_request if needed

if __name__ == "__main__":
    # For development only
    app.run(host="0.0.0.0", port=settings.port, debug=False)

