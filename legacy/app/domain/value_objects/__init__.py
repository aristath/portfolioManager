"""Domain value objects.

Import from app.shared.domain.value_objects instead.
"""

from app.domain.value_objects.product_type import ProductType
from app.shared.domain.value_objects import Currency, Money, Price

__all__ = ["Currency", "Money", "Price", "ProductType"]
