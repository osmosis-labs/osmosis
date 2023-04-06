from decimal import *

def calc_liquidity_0(amount_zero: Decimal, sqrtPriceA: Decimal, sqrtPriceB: Decimal) -> Decimal:
    """Calculates and returns liquidity zero. 
    """
    return amount_zero * (sqrtPriceA * sqrtPriceB) / (sqrtPriceB - sqrtPriceA)

def calc_liquidity_1(amount_one: Decimal, sqrtPriceA: Decimal, sqrtPriceB: Decimal) -> Decimal:
    """Calculates and returns liquidity one. 
    """
    return amount_one * (sqrtPriceB - sqrtPriceA)
