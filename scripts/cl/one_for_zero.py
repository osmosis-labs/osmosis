from typing import Tuple
import sympy as sp
from common import *

def get_next_sqrt_price(liquidity: sp.Float, sqrt_price_current: sp.Float, token_in: sp.Float) -> sp.Float:
    """ Return the next sqrt price when swapping token one for zero.
    """
    return sqrt_price_current + token_in / liquidity

def get_token_out(liquidity: sp.Float, sqrt_price_current: sp.Float, sqrt_price_next: sp.Float) -> sp.Float:
    """ Returns the token out when swapping token one for zero given the token in.
    """
    return liquidity * (sqrt_price_next - sqrt_price_current) / (sqrt_price_next * sqrt_price_current)

def get_token_in_swap_in_given_out(liquidity: sp.Float, sqrt_price_current: sp.Float, sqrt_price_next: sp.Float) -> sp.Float:
    """ Returns the token in when swapping token one for zero given the token out.

    In this case, the calculation is the same as computing token out when given token in.
    """
    return get_token_out(liquidity, sqrt_price_current, sqrt_price_next)

def get_expected_token_in(liquidity: sp.Float, sqrt_price_current: sp.Float, sqrt_price_next: sp.Float):
    """ Returns the expected token in when swapping token one for zero. 
    """
    return liquidity * sp.Abs((sqrt_price_current - sqrt_price_next))

def calc_test_case_out_given_in(liquidity: sp.Float, sqrt_price_current: sp.Float, token_in: sp.Float, swap_fee: sp.Float) -> Tuple[sp.Float, sp.Float, sp.Float]:
    """ Computes and prints all one for zero test case parameters when swapping for out given in.
    Next sqrt price is computed from the given parameters.

    Returns the next square root price, token out and fee amount per share.
    """
    sqrt_price_next = get_next_sqrt_price(liquidity, sqrt_price_current, token_in * (1 - swap_fee))
    price_next = sp.Pow(sqrt_price_next, 2)
    token_out = get_token_out(liquidity, sqrt_price_current, sqrt_price_next)
    fee_amount_per_share = get_fee_amount_per_share(token_in, swap_fee, liquidity)

    print(F"sqrt_price_next: {sqrt_price_next}")
    print(F"price_next: {price_next}")
    print(F"token_out: {token_out}")
    print(F"fee_amount_per_share: {fee_amount_per_share}")

    return sqrt_price_next, token_out, fee_amount_per_share

def calc_test_case_in_given_out(liquidity: sp.Float, sqrt_price_current: sp.Float, token_in: sp.Float, swap_fee: sp.Float) -> Tuple[sp.Float, sp.Float, sp.Float]:
    """ Computes and prints all one for zero test case parameters when swapping for in given out.
    Next sqrt price is computed from the given parameters.

    Returns the next square root price, token in and fee amount per share.
    """
    sqrt_price_next = get_next_sqrt_price(liquidity, sqrt_price_current, token_in)
    price_next = sp.Pow(sqrt_price_next, 2)
    token_in = get_token_in_swap_in_given_out(liquidity, sqrt_price_current, sqrt_price_next)
    fee_amount_per_share = get_fee_amount_per_share(token_in, swap_fee, liquidity)

    token_in_after_fee = token_in * (1 + swap_fee)

    print(F"sqrt_price_next: {sqrt_price_next}")
    print(F"price_next: {price_next}")
    print(F"token_in_after_fee: {token_in_after_fee}")
    print(F"fee_amount_per_share: {fee_amount_per_share}")

    return sqrt_price_next, token_in_after_fee, fee_amount_per_share

def calc_test_case_with_next_sqrt_price_out_given_in(liquidity: sp.Float, sqrt_price_current: sp.Float, sqrt_price_next: sp.Float, swap_fee: sp.Float) -> Tuple[sp.Float, sp.Float, sp.Float]:
    """ Computes and prints one for zero test case parameters when next square root price is known.
    Assumes swapping for token out given in.

    Returns the expected token in, token out and fee amount per share. 
    """
    price_next = sp.Pow(sqrt_price_next, 2)
    expected_token_in_before_fee = get_expected_token_in(liquidity, sqrt_price_current, sqrt_price_next)
    expected_token_in = expected_token_in_before_fee * (1 + swap_fee)

    token_out = get_token_out(liquidity, sqrt_price_current, sqrt_price_next)
    fee_amount_per_share = get_fee_amount_per_share(expected_token_in_before_fee, swap_fee, liquidity)

    print(F"given sqrt_price_next: {sqrt_price_next}")
    print(F"price_next: {price_next}")
    print(F"expected_token_in: {expected_token_in}")
    print(F"token_out: {token_out}")
    print(F"fee_amount_per_share: {fee_amount_per_share}")

    return expected_token_in, token_out, fee_amount_per_share

def calc_test_case_with_next_sqrt_price_in_given_out(liquidity: sp.Float, sqrt_price_current: sp.Float, sqrt_price_next: sp.Float, swap_fee: sp.Float) -> Tuple[sp.Float, sp.Float, sp.Float]:
    """ Computes and prints one for zero test case parameters when next square root price is known.
    Assumes swapping for token in given out.
    
    Returns expected token out, token in after fee, and fee amount per share. 
    """
    expected_token_out = get_expected_token_in(liquidity, sqrt_price_current, sqrt_price_next)
    token_in = get_token_in_swap_in_given_out(liquidity, sqrt_price_current, sqrt_price_next)
    fee_amount_per_share = get_fee_amount_per_share(token_in, swap_fee, liquidity)
    token_in_after_fee = token_in * (1 + swap_fee)

    print(F"current sqrt price: {sqrt_price_current}")
    print(F"given sqrt_price_next: {sqrt_price_next}")
    print(F"liquidity: {liquidity}")
    print(F"expected_token_out: {expected_token_out}")
    print(F"token_in_after_fee: {token_in_after_fee}")
    print(F"fee_amount_per_share: {fee_amount_per_share}")

    return expected_token_out, token_in_after_fee, fee_amount_per_share
