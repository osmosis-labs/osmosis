import sympy as sp

precision = 30

class SqrtPriceRange:
  def __init__(self, sqrt_price_current: int, sqrt_price_next: int, liquidity: sp.Float):
    self.sqrt_price_start = sp.sqrt(sdk_dec(sqrt_price_current))
    if sqrt_price_next is not None:
        self.sqrt_price_next = sp.sqrt(sdk_dec(sqrt_price_next))
    self.liquidity = liquidity

def sdk_dec(string: str) -> sp.Float:
    """ Return an equivalent of a Python Decimal. 
    """
    return sp.Float(string, precision)

def get_fee_amount_per_share(token_in: sp.Float, swap_fee: sp.Float, liquidity: sp.Float) -> sp.Float:
    """ Returns the fee amount per share.
    """
    fee_charge_total = token_in * swap_fee
    print(F"token_in: {token_in}")
    print(F"swap_fee: {swap_fee}")
    print(F"liquidity: {liquidity}")
    print(F"fee_charge: {fee_charge_total}")
    return fee_charge_total / liquidity

zero = sdk_dec("0")
