"""Stock scoring engine.

Scoring weights:
- Technical: 50% (trend, momentum, volatility)
- Analyst: 30% (recommendations, price targets)
- Fundamental: 20% (PE, growth, margins)
"""

import logging
from datetime import datetime
from typing import Optional
from dataclasses import dataclass

import numpy as np

from app.services import yahoo

logger = logging.getLogger(__name__)


@dataclass
class TechnicalScore:
    """Technical analysis score components."""
    trend_score: float  # 0-1, price vs moving averages
    momentum_score: float  # 0-1, rate of change
    volatility_score: float  # 0-1, lower volatility = higher score
    total: float  # Weighted combination


@dataclass
class AnalystScore:
    """Analyst recommendation score components."""
    recommendation_score: float  # 0-1, based on buy/hold/sell
    target_score: float  # 0-1, based on upside potential
    total: float  # Weighted combination


@dataclass
class FundamentalScore:
    """Fundamental analysis score components."""
    valuation_score: float  # 0-1, PE ratio
    growth_score: float  # 0-1, revenue/earnings growth
    profitability_score: float  # 0-1, margins
    total: float  # Weighted combination


@dataclass
class StockScore:
    """Complete stock score with all components."""
    symbol: str
    technical: TechnicalScore
    analyst: AnalystScore
    fundamental: FundamentalScore
    total_score: float  # Final weighted score
    calculated_at: datetime


def calculate_technical_score(symbol: str, yahoo_symbol: str = None) -> Optional[TechnicalScore]:
    """
    Calculate technical score from price data.

    Components:
    - Trend (40%): Price vs 50-day and 200-day MA
    - Momentum (35%): 14-day and 30-day rate of change
    - Volatility (25%): Standard deviation (inverse - lower is better)

    Args:
        symbol: Tradernet symbol
        yahoo_symbol: Optional explicit Yahoo symbol override
    """
    try:
        prices = yahoo.get_historical_prices(symbol, yahoo_symbol=yahoo_symbol, period="1y")

        if len(prices) < 50:
            logger.warning(f"Insufficient price data for {symbol}")
            return None

        closes = np.array([p.close for p in prices])

        # Trend score: price vs moving averages
        current_price = closes[-1]
        ma_50 = np.mean(closes[-50:]) if len(closes) >= 50 else np.mean(closes)
        ma_200 = np.mean(closes[-200:]) if len(closes) >= 200 else np.mean(closes)

        # Score based on position relative to MAs
        # Above both MAs = bullish, below both = bearish
        trend_score = 0.5  # neutral

        if current_price > ma_50:
            trend_score += 0.25
        else:
            trend_score -= 0.25

        if current_price > ma_200:
            trend_score += 0.25
        else:
            trend_score -= 0.25

        # Bonus for golden cross (50 > 200)
        if ma_50 > ma_200:
            trend_score = min(1.0, trend_score + 0.1)

        trend_score = max(0, min(1, trend_score))

        # Momentum score: rate of change
        roc_14 = (closes[-1] - closes[-14]) / closes[-14] if len(closes) >= 14 else 0
        roc_30 = (closes[-1] - closes[-30]) / closes[-30] if len(closes) >= 30 else 0

        # Convert ROC to 0-1 score
        # Assume good momentum is +5% to +20% over 30 days
        momentum_raw = (roc_14 * 0.4 + roc_30 * 0.6)
        momentum_score = 0.5 + (momentum_raw * 5)  # Scale to 0-1 range
        momentum_score = max(0, min(1, momentum_score))

        # Volatility score: lower is better
        returns = np.diff(closes) / closes[:-1]
        volatility = np.std(returns) * np.sqrt(252)  # Annualized

        # Convert to score: 10% vol = 1.0, 50% vol = 0.0
        volatility_score = 1 - (volatility - 0.1) / 0.4
        volatility_score = max(0, min(1, volatility_score))

        # Combined technical score
        total = (
            trend_score * 0.40 +
            momentum_score * 0.35 +
            volatility_score * 0.25
        )

        return TechnicalScore(
            trend_score=round(trend_score, 3),
            momentum_score=round(momentum_score, 3),
            volatility_score=round(volatility_score, 3),
            total=round(total, 3),
        )

    except Exception as e:
        logger.error(f"Failed to calculate technical score for {symbol}: {e}")
        return None


def calculate_analyst_score(symbol: str, yahoo_symbol: str = None) -> Optional[AnalystScore]:
    """
    Calculate analyst score from recommendations and price targets.

    Components:
    - Recommendation (60%): Buy/Hold/Sell consensus
    - Price Target (40%): Upside potential

    Args:
        symbol: Tradernet symbol
        yahoo_symbol: Optional explicit Yahoo symbol override
    """
    try:
        data = yahoo.get_analyst_data(symbol, yahoo_symbol=yahoo_symbol)

        if not data:
            return None

        # Recommendation score (already 0-1 from yahoo service)
        recommendation_score = data.recommendation_score

        # Target score: based on upside potential
        # 0% upside = 0.5, 20%+ upside = 1.0, -20% = 0.0
        upside = data.upside_pct / 100  # Convert to decimal
        target_score = 0.5 + (upside * 2.5)  # Scale
        target_score = max(0, min(1, target_score))

        # Combined analyst score
        total = (
            recommendation_score * 0.60 +
            target_score * 0.40
        )

        return AnalystScore(
            recommendation_score=round(recommendation_score, 3),
            target_score=round(target_score, 3),
            total=round(total, 3),
        )

    except Exception as e:
        logger.error(f"Failed to calculate analyst score for {symbol}: {e}")
        return None


def calculate_fundamental_score(symbol: str, yahoo_symbol: str = None) -> Optional[FundamentalScore]:
    """
    Calculate fundamental score from financial metrics.

    Components:
    - Valuation (40%): P/E ratio relative to market
    - Growth (35%): Revenue and earnings growth
    - Profitability (25%): Margins

    Args:
        symbol: Tradernet symbol
        yahoo_symbol: Optional explicit Yahoo symbol override
    """
    try:
        data = yahoo.get_fundamental_data(symbol, yahoo_symbol=yahoo_symbol)

        if not data:
            return None

        # Valuation score: based on P/E
        # P/E 10 = 1.0, P/E 25 = 0.5, P/E 40+ = 0.0
        pe = data.pe_ratio or data.forward_pe or 25  # Default to market average
        if pe <= 0:
            pe = 25  # Handle negative earnings

        valuation_score = 1 - (pe - 10) / 30
        valuation_score = max(0, min(1, valuation_score))

        # Growth score: revenue and earnings growth
        rev_growth = data.revenue_growth or 0
        earn_growth = data.earnings_growth or 0

        # 20%+ growth = 1.0, 0% = 0.5, -20% = 0.0
        growth_raw = (rev_growth * 0.5 + earn_growth * 0.5)
        growth_score = 0.5 + (growth_raw * 2.5)
        growth_score = max(0, min(1, growth_score))

        # Profitability score: margins
        profit_margin = data.profit_margin or 0
        operating_margin = data.operating_margin or 0

        # 20%+ margin = 1.0, 0% = 0.5, negative = 0.0
        margin_raw = (profit_margin * 0.6 + operating_margin * 0.4)
        profitability_score = 0.5 + (margin_raw * 2.5)
        profitability_score = max(0, min(1, profitability_score))

        # Combined fundamental score
        total = (
            valuation_score * 0.40 +
            growth_score * 0.35 +
            profitability_score * 0.25
        )

        return FundamentalScore(
            valuation_score=round(valuation_score, 3),
            growth_score=round(growth_score, 3),
            profitability_score=round(profitability_score, 3),
            total=round(total, 3),
        )

    except Exception as e:
        logger.error(f"Failed to calculate fundamental score for {symbol}: {e}")
        return None


def calculate_stock_score(symbol: str, yahoo_symbol: str = None) -> Optional[StockScore]:
    """
    Calculate complete stock score with all components.

    Final weights:
    - Technical: 50%
    - Analyst: 30%
    - Fundamental: 20%

    Args:
        symbol: Tradernet symbol
        yahoo_symbol: Optional explicit Yahoo symbol override
    """
    technical = calculate_technical_score(symbol, yahoo_symbol)
    analyst = calculate_analyst_score(symbol, yahoo_symbol)
    fundamental = calculate_fundamental_score(symbol, yahoo_symbol)

    # Handle missing scores with defaults
    tech_score = technical.total if technical else 0.5
    analyst_score = analyst.total if analyst else 0.5
    fund_score = fundamental.total if fundamental else 0.5

    # Calculate final weighted score
    total_score = (
        tech_score * 0.50 +
        analyst_score * 0.30 +
        fund_score * 0.20
    )

    # Create default scores if missing
    if not technical:
        technical = TechnicalScore(0.5, 0.5, 0.5, 0.5)
    if not analyst:
        analyst = AnalystScore(0.5, 0.5, 0.5)
    if not fundamental:
        fundamental = FundamentalScore(0.5, 0.5, 0.5, 0.5)

    return StockScore(
        symbol=symbol,
        technical=technical,
        analyst=analyst,
        fundamental=fundamental,
        total_score=round(total_score, 3),
        calculated_at=datetime.now(),
    )


async def score_all_stocks(db) -> list[StockScore]:
    """
    Score all active stocks in the universe and update database.

    Args:
        db: Database connection

    Returns:
        List of calculated scores
    """
    from aiosqlite import Connection

    # Get all active stocks with their Yahoo symbol overrides
    cursor = await db.execute(
        "SELECT symbol, yahoo_symbol FROM stocks WHERE active = 1"
    )
    rows = await cursor.fetchall()

    scores = []
    for row in rows:
        symbol = row[0]
        yahoo_symbol = row[1]  # May be None
        logger.info(f"Scoring {symbol}...")
        score = calculate_stock_score(symbol, yahoo_symbol)

        if score:
            scores.append(score)

            # Update database
            await db.execute(
                """
                INSERT OR REPLACE INTO scores
                (symbol, technical_score, analyst_score, fundamental_score,
                 total_score, calculated_at)
                VALUES (?, ?, ?, ?, ?, ?)
                """,
                (
                    symbol,
                    score.technical.total,
                    score.analyst.total,
                    score.fundamental.total,
                    score.total_score,
                    score.calculated_at.isoformat(),
                ),
            )

    await db.commit()
    logger.info(f"Scored {len(scores)} stocks")

    return scores
