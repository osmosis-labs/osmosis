import sympy as sp
from common import *

def get_next_sqrt_price(liquidity: sp.Float, sqrt_price_current: sp.Float, token_in: sp.Float, swap_fee: sp.Float) -> sp.Float:
    """ Return the next sqrt price when swapping token one for zero. 
    """
    return sqrt_price_current + (token_in * (1 - swap_fee) / liquidity)

def get_token_out(liquidity: sp.Float, sqrt_price_current: sp.Float, sqrt_price_next: sp.Float) -> sp.Float:
    """ Returns the token out when swapping token one for zero. 
    """
    return liquidity * (sqrt_price_next - sqrt_price_current) / (sqrt_price_next * sqrt_price_current)

def calc_test_case(liquidity: sp.Float, sqrt_price_current: sp.Float, token_in: sp.Float, swap_fee: sp.Float):
    """ Computes and prints all one for zero test case parameters. 
    """
    sqrt_price_next = get_next_sqrt_price(liquidity, sqrt_price_current, token_in, swap_fee)
    price_next = sp.Pow(sqrt_price_next, 2)
    token_out = get_token_out(liquidity, sqrt_price_current, sqrt_price_next)
    fee_amount_per_share = get_fee_amount_per_share(token_in, swap_fee, liquidity)

    print(F"sqrt_price_next: {sqrt_price_next}")
    print(F"price_next: {price_next}")
    print(F"token_out: {token_out}")
    print(F"fee_amount_per_share: {fee_amount_per_share}")
