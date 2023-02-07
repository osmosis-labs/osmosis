import sympy as sp

precision = 30

# SqrtPriceRange represents a price range between the current and the next tick
# as well as the liquidity in that price range.
# When swapping token zero for one, sqrt_price_current >= sqrt_price_next.
# When swapping token one for zero, sqrt_price_current <= sqrt_price_next.
#
# sqrt_price_next can be None. When it is None, that implies that the next sqrt
# price must be calculated. In such a case, next sqrt price depends on the
# remaining amount of token in to be swapped. This occurs for the last sqrt price
# range in the collection of sqrt price ranges that represent a swap.
#
# For example, I might have a swap of 100 ETH in from 5000 to 5001 with liquidity
# X, from 5001 to 5002 with liquidity Y,
# and from 5002 to UNKNOWN with liquidity Z. In this case, the UNKNOWN
# depends on how much ETH we have remaining after consuming liquidity X and Y.
class SqrtPriceRange:
  def __init__(self, sqrt_price_current: int, sqrt_price_next: int, liquidity: sp.Float):
    self.sqrt_price_start = sp.sqrt(fixed_prec_dec(sqrt_price_current))
    if sqrt_price_next is not None:
        self.sqrt_price_next = sp.sqrt(fixed_prec_dec(sqrt_price_next))
    else:
       self.sqrt_price_next =  None
    self.liquidity = liquidity

def fixed_prec_dec(string: str) -> sp.Float:
    """ Return an equivalent of a Python Decimal with fixed precision. 
    """
    return sp.Float(string, precision)

def get_fee_amount_per_share(token_in: sp.Float, swap_fee: sp.Float, liquidity: sp.Float) -> sp.Float:
    """ Returns the fee amount per share.
    """
    fee_charge_total = token_in * swap_fee
    return fee_charge_total / liquidity

zero = fixed_prec_dec("0")

def validate_confirmed_results(actual_token_amount: sp.Float, fee_growth_per_share_total: sp.Float, expected_token_amount: sp.Float, expected_fee_growth_per_share_total: sp.Float):
    """Validates the results of a calc concentrated liquidity test case estimates.

    This validation helper exists to make sure that subsequent changes to the script do not break test cases.
    """

    if sp.N(actual_token_amount, 18) != sp.N(expected_token_amount, 18):
        raise Exception(F"actual_token_amount {actual_token_amount} does not match expected_token_amount {expected_token_amount}")
    
    if sp.N(fee_growth_per_share_total, 18) != sp.N(expected_fee_growth_per_share_total, 18):
        raise Exception(F"fee_growth_per_share_total {fee_growth_per_share_total} does not match expected_fee_growth_per_share_total {expected_fee_growth_per_share_total}")
