import sympy as sp

def calc_liquidity_0(amount_zero: sp.Float, sqrtPriceA: sp.Float, sqrtPriceB: sp.Float) -> sp.Float:
    """Calculates and returns liquidity zero. 
    """
    return amount_zero * (sqrtPriceA * sqrtPriceB) / (sqrtPriceB - sqrtPriceA)

def calc_liquidity_1(amount_one: sp.Float, sqrtPriceA: sp.Float, sqrtPriceB: sp.Float) -> sp.Float:
    """Calculates and returns liquidity one. 
    """
    return amount_one * (sqrtPriceB - sqrtPriceA)
