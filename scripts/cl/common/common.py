from decimal import *
import math

from common.sdk_dec import *

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
  def __init__(self, sqrt_price_current: int, sqrt_price_next: int, liquidity: Decimal):
    self.sqrt_price_start = new(new(sqrt_price_current).sqrt())
    if sqrt_price_next is not None:
        self.sqrt_price_next = new(new(sqrt_price_next).sqrt())
    else:
       self.sqrt_price_next =  None
    self.liquidity = liquidity

def validate_confirmed_results(actual_token_amount: Decimal, spread_rewards_growth_per_share_total: Decimal, expected_token_amount: Decimal, expected_spread_rewards_growth_per_share_total: Decimal):
    """Validates the results of a calc concentrated liquidity test case estimates.

    This validation helper exists to make sure that subsequent changes to the script do not break test cases.
    """

    if math.floor(actual_token_amount) != math.floor(expected_token_amount):
        raise Exception(F"actual_token_amount {actual_token_amount} does not match expected_token_amount {expected_token_amount}")
    
    if abs(spread_rewards_growth_per_share_total - expected_spread_rewards_growth_per_share_total) > (1 / math.pow(10, 18)):
        raise Exception(F"spread_rewards_growth_per_share_total {spread_rewards_growth_per_share_total} does not match expected_spread_rewards_growth_per_share_total {expected_spread_rewards_growth_per_share_total}")
