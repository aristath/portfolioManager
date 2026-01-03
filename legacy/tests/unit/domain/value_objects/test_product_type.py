"""Tests for ProductType value object."""

from app.domain.value_objects.product_type import ProductType


class TestProductTypeFromYahooQuoteType:
    """Test ProductType.from_yahoo_quote_type() heuristics."""

    def test_equity_detection(self):
        """Test EQUITY detection from Yahoo quoteType."""
        assert (
            ProductType.from_yahoo_quote_type("EQUITY", "Apple Inc.")
            == ProductType.EQUITY
        )

    def test_etf_detection(self):
        """Test ETF detection from Yahoo quoteType."""
        assert (
            ProductType.from_yahoo_quote_type("ETF", "Vanguard S&P 500 ETF")
            == ProductType.ETF
        )

    def test_mutual_fund_detection(self):
        """Test MUTUALFUND detection from Yahoo quoteType."""
        assert (
            ProductType.from_yahoo_quote_type("MUTUALFUND", "Vanguard 500 Index Fund")
            == ProductType.MUTUALFUND
        )

    def test_etc_detection_from_gold(self):
        """Test ETC detection from commodity name (Gold)."""
        assert (
            ProductType.from_yahoo_quote_type("MUTUALFUND", "Gold ETC")
            == ProductType.ETC
        )

    def test_etc_detection_from_silver(self):
        """Test ETC detection from commodity name (Silver)."""
        assert (
            ProductType.from_yahoo_quote_type("MUTUALFUND", "iShares Silver Trust")
            == ProductType.ETC
        )

    def test_etc_detection_from_oil(self):
        """Test ETC detection from commodity name (Oil)."""
        assert (
            ProductType.from_yahoo_quote_type("MUTUALFUND", "United States Oil Fund")
            == ProductType.ETC
        )

    def test_etc_detection_from_aluminium(self):
        """Test ETC detection from commodity name (Aluminium)."""
        assert (
            ProductType.from_yahoo_quote_type("MUTUALFUND", "WisdomTree Aluminium ETC")
            == ProductType.ETC
        )

    def test_etc_detection_from_copper(self):
        """Test ETC detection from commodity name (Copper)."""
        assert (
            ProductType.from_yahoo_quote_type(
                "MUTUALFUND", "Global X Copper Miners ETF"
            )
            == ProductType.ETC
        )

    def test_etc_detection_from_platinum(self):
        """Test ETC detection from commodity name (Platinum)."""
        assert (
            ProductType.from_yahoo_quote_type("MUTUALFUND", "Aberdeen Platinum Trust")
            == ProductType.ETC
        )

    def test_etc_detection_from_palladium(self):
        """Test ETC detection from commodity name (Palladium)."""
        assert (
            ProductType.from_yahoo_quote_type("MUTUALFUND", "ETFS Palladium Shares")
            == ProductType.ETC
        )

    def test_etc_detection_from_natural_gas(self):
        """Test ETC detection from commodity name (Natural Gas)."""
        assert (
            ProductType.from_yahoo_quote_type(
                "MUTUALFUND", "United States Natural Gas Fund"
            )
            == ProductType.ETC
        )

    def test_etc_detection_case_insensitive(self):
        """Test ETC detection is case insensitive."""
        assert (
            ProductType.from_yahoo_quote_type("MUTUALFUND", "GOLD etc")
            == ProductType.ETC
        )
        assert (
            ProductType.from_yahoo_quote_type("MUTUALFUND", "Silver TRUST")
            == ProductType.ETC
        )

    def test_etf_overrides_etc_heuristic_when_quote_type_is_etf(self):
        """Test that quoteType=ETF takes precedence over commodity heuristics."""
        # Even if name contains "gold", quoteType=ETF should win
        assert (
            ProductType.from_yahoo_quote_type("ETF", "Gold Miners ETF")
            == ProductType.ETF
        )

    def test_etf_detection_from_name_with_etf(self):
        """Test ETF detection from name containing 'ETF' when quoteType is MUTUALFUND."""
        assert (
            ProductType.from_yahoo_quote_type("MUTUALFUND", "UCITS ETF")
            == ProductType.ETF
        )

    def test_unknown_quote_type(self):
        """Test UNKNOWN for unrecognized quoteType."""
        assert (
            ProductType.from_yahoo_quote_type("CRYPTOCURRENCY", "Bitcoin ETF")
            == ProductType.UNKNOWN
        )

    def test_none_quote_type(self):
        """Test UNKNOWN for None quoteType."""
        assert (
            ProductType.from_yahoo_quote_type(None, "Some Product")
            == ProductType.UNKNOWN
        )

    def test_empty_product_name(self):
        """Test handling of empty product name."""
        assert (
            ProductType.from_yahoo_quote_type("MUTUALFUND", "")
            == ProductType.MUTUALFUND
        )
        assert ProductType.from_yahoo_quote_type("EQUITY", None) == ProductType.EQUITY

    def test_mutual_fund_without_etc_indicators(self):
        """Test that regular mutual funds are not misclassified as ETCs."""
        assert (
            ProductType.from_yahoo_quote_type("MUTUALFUND", "Fidelity 500 Index Fund")
            == ProductType.MUTUALFUND
        )
        assert (
            ProductType.from_yahoo_quote_type(
                "MUTUALFUND", "Vanguard Total Stock Market"
            )
            == ProductType.MUTUALFUND
        )


class TestProductTypeStringConversion:
    """Test ProductType string conversion methods."""

    def test_value_property(self):
        """Test that ProductType.value gives the string value."""
        assert ProductType.EQUITY.value == "EQUITY"
        assert ProductType.ETF.value == "ETF"
        assert ProductType.ETC.value == "ETC"
        assert ProductType.MUTUALFUND.value == "MUTUALFUND"
        assert ProductType.UNKNOWN.value == "UNKNOWN"

    def test_name_property(self):
        """Test that ProductType.name gives the member name."""
        assert ProductType.EQUITY.name == "EQUITY"
        assert ProductType.ETF.name == "ETF"
        assert ProductType.ETC.name == "ETC"

    def test_repr_conversion(self):
        """Test that ProductType has useful repr."""
        assert "EQUITY" in repr(ProductType.EQUITY)
        assert "ETF" in repr(ProductType.ETF)


class TestProductTypeEquality:
    """Test ProductType equality and comparison."""

    def test_equality(self):
        """Test ProductType equality."""
        assert ProductType.EQUITY == ProductType.EQUITY
        assert ProductType.ETF == ProductType.ETF
        assert ProductType.EQUITY != ProductType.ETF

    def test_identity(self):
        """Test ProductType identity (enum singletons)."""
        assert ProductType.EQUITY is ProductType.EQUITY
        assert ProductType.ETF is ProductType.ETF
