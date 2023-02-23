from typing import Tuple
import math
from decimal import *
from common import *
import common.sdk_dec as sdk_dec


def get_next_sqrt_price(liquidity: Decimal, sqrt_price_current: Decimal, token_in: Decimal) -> Decimal:
    """ Return the next sqrt price when swapping token one for zero.
    """
    return sqrt_price_current + sdk_dec.quo(token_in, liquidity)


def get_token_out(liquidity: Decimal, sqrt_price_current: Decimal, sqrt_price_next: Decimal) -> Decimal:
    """ Returns the token out when swapping token one for zero given the token in.
    """
    return sdk_dec.mul(liquidity, sdk_dec.quo((sqrt_price_next - sqrt_price_current), sdk_dec.mul(sqrt_price_next, sqrt_price_current)))


def get_token_in_swap_in_given_out(liquidity: Decimal, sqrt_price_current: Decimal, sqrt_price_next: Decimal) -> Decimal:
    """ Returns the token in when swapping token one for zero given the token out.
    In this case, the calculation is the same as computing token out when given token in.
    """
    return get_token_out(liquidity, sqrt_price_current, sqrt_price_next)


def calc_amount_one_delta(liquidity: Decimal, sqrt_price_current: Decimal, sqrt_price_next: Decimal, should_round_up: bool) -> Decimal:
    """ Returns the expected token in when swapping token one for zero. 
    """
    result = sdk_dec.mul(liquidity, abs(sqrt_price_current - sqrt_price_next))
    if should_round_up:
        return Decimal(math.ceil(result))
    return result


def calc_test_case_out_given_in(liquidity: Decimal, sqrt_price_current: Decimal, token_in_remaining: Decimal, swap_fee: Decimal) -> Tuple[Decimal, Decimal, Decimal]:
    """ Computes and prints all one for zero test case parameters when swapping for out given in.
    Next sqrt price is computed from the given parameters.
    Returns the next square root price, token out and fee amount per share.
    """

    token_in_remaining_after_fee = sdk_dec.mul(
        token_in_remaining, (sdk_dec.one - swap_fee))

    sqrt_price_next = get_next_sqrt_price(
        liquidity, sqrt_price_current, token_in_remaining_after_fee)

    print(F"token_in_remaining_after_fee: {token_in_remaining_after_fee}")

    token_in_after_fee_rounded_up = calc_amount_one_delta(
        liquidity, sqrt_price_current, sqrt_price_next, True)

    print(F"token_in_after_fee_rounded_up: {token_in_after_fee_rounded_up}")
    token_out = get_token_out(liquidity, sqrt_price_current, sqrt_price_next)

    fee_charge_total = sdk_dec.zero
    if swap_fee > sdk_dec.zero:
        fee_charge_total = token_in_remaining - token_in_after_fee_rounded_up
    fee_amount_per_share = sdk_dec.quo(fee_charge_total, liquidity)

    print(F"liquidity: {liquidity}")
    print(F"sqrt_price_current: {sqrt_price_current}")
    print(F"sqrt_price_next: {sqrt_price_next}")
    print(F"token_out: {token_out}")
    print(F"fee_amount_per_share: {fee_amount_per_share}")

    return sqrt_price_next, token_out, fee_amount_per_share


def calc_test_case_in_given_out(liquidity: Decimal, sqrt_price_current: Decimal, token_in: Decimal, swap_fee: Decimal) -> Tuple[Decimal, Decimal, Decimal]:
    """ Computes and prints all one for zero test case parameters when swapping for in given out.
    Next sqrt price is computed from the given parameters.
    Returns the next square root price, token in and fee amount per share.
    """
    sqrt_price_next = get_next_sqrt_price(
        liquidity, sqrt_price_current, token_in)
    price_next = math.pow(sqrt_price_next, 2)
    
    token_in = get_token_in_swap_in_given_out(
        liquidity, sqrt_price_current, sqrt_price_next)

    total_fee = sdk_dec.quo(sdk_dec.mul(
        token_in, swap_fee), (sdk_dec.one - swap_fee))

    fee_amount_per_share = sdk_dec.quo(total_fee, liquidity)

    token_in_after_fee = token_in + total_fee

    print(F"sqrt_price_next: {sqrt_price_next}")
    print(F"price_next: {price_next}")
    print(F"liquidity: {liquidity}")
    print(F"token_in_after_fee: {token_in_after_fee}")
    print(F"fee_amount_per_share: {fee_amount_per_share}")

    return sqrt_price_next, token_in_after_fee, fee_amount_per_share


def calc_test_case_with_next_sqrt_price_out_given_in(liquidity: Decimal, sqrt_price_current: Decimal, sqrt_price_next: Decimal, swap_fee: Decimal) -> Tuple[Decimal, Decimal, Decimal]:
    """ Computes and prints one for zero test case parameters when next square root price is known.
    Assumes swapping for token out given in.
    Returns the expected token in, token out and fee amount per share. 
    """
    expected_token_in_before_fee = calc_amount_one_delta(
        liquidity, sqrt_price_current, sqrt_price_next, True)
    expected_fee = sdk_dec.zero
    if swap_fee > sdk_dec.zero:
        expected_fee = sdk_dec.quo(sdk_dec.mul(
            expected_token_in_before_fee, swap_fee), sdk_dec.one - swap_fee)

    expected_token_in = expected_token_in_before_fee + expected_fee

    token_out = get_token_out(liquidity, sqrt_price_current, sqrt_price_next)
    fee_amount_per_share = sdk_dec.quo(expected_fee, liquidity)

    print(F"liquidity: {liquidity}")
    print(F"sqrt_price_current: {sqrt_price_current}")
    print(F"given sqrt_price_next: {sqrt_price_next}")
    print(F"expected_token_in_before_fee: {expected_token_in_before_fee}")
    print(F"expected_token_in: {expected_token_in}")
    print(F"token_out: {token_out}")
    print(F"fee_charge_total: {expected_fee}")
    print(F"fee_amount_per_share: {fee_amount_per_share}")

    return expected_token_in, token_out, fee_amount_per_share


def calc_test_case_with_next_sqrt_price_in_given_out(liquidity: Decimal, sqrt_price_current: Decimal, sqrt_price_next: Decimal, swap_fee: Decimal) -> Tuple[Decimal, Decimal, Decimal]:
    """ Computes and prints one for zero test case parameters when next square root price is known.
    Assumes swapping for token in given out.

    Returns expected token out, token in after fee, and fee amount per share. 
    """
    
    expected_token_out = calc_amount_one_delta(
        liquidity, sqrt_price_current, sqrt_price_next, True)
    token_in = get_token_in_swap_in_given_out(
        liquidity, sqrt_price_current, sqrt_price_next)

    total_fee = sdk_dec.quo(sdk_dec.mul(
        token_in, swap_fee), (sdk_dec.one - swap_fee))


    fee_amount_per_share = sdk_dec.quo(total_fee, liquidity)
    token_in_after_fee = token_in + total_fee

    print(F"current sqrt price: {sqrt_price_current}")
    print(F"given sqrt_price_next: {sqrt_price_next}")
    print(F"liquidity: {liquidity}")
    print(F"expected_token_out: {expected_token_out}")
    print(F"token_in_after_fee: {token_in_after_fee}")
    print(F"fee_amount_per_share: {fee_amount_per_share}")

    return expected_token_out, token_in_after_fee, fee_amount_per_share
