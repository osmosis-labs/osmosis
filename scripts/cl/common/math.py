import math
from decimal import *
import common.sdk_dec as sdk_dec

def calc_liquidity_0(amount_zero: Decimal, sqrtPriceA: Decimal, sqrtPriceB: Decimal) -> Decimal:
    """Calculates and returns liquidity zero. 
    """
    return amount_zero * (sqrtPriceA * sqrtPriceB) / (sqrtPriceB - sqrtPriceA)

def calc_liquidity_1(amount_one: Decimal, sqrtPriceA: Decimal, sqrtPriceB: Decimal) -> Decimal:
    """Calculates and returns liquidity one. 
    """
    return amount_one * (sqrtPriceB - sqrtPriceA)


def calc_amount_zero_delta(liquidity: Decimal, sqrt_price_current: Decimal, sqrt_price_next: Decimal, should_round_up: bool) -> Decimal:
    """ Returns the expected token in when swapping token zero for one. 
    """
    mul1 = sdk_dec.mul(liquidity, (sqrt_price_current - sqrt_price_next))
    print(mul1)
    result = sdk_dec.quo(sdk_dec.mul(liquidity, (sqrt_price_current -
                         sqrt_price_next)), sdk_dec.mul(sqrt_price_current, sqrt_price_next))
    if should_round_up:
        return sdk_dec.new(str(math.ceil(result)))
    return result


def calc_amount_one_delta(liquidity: Decimal, sqrt_price_current: Decimal, sqrt_price_next: Decimal, should_round_up: bool) -> Decimal:
    """ Returns the expected token in when swapping token one for zero. 
    """
    result = sdk_dec.mul(liquidity, abs(sqrt_price_current - sqrt_price_next))
    if should_round_up:
        return Decimal(math.ceil(result))
    return result
