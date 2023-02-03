import math
from typing import Tuple
from decimal import *
from common import *
import common.sdk_dec as sdk_dec


def get_next_sqrt_price(liquidity: Decimal, sqrt_price_current: Decimal, token_in: Decimal) -> Decimal:
    """ Return the next sqrt price when swapping token zero for one.
    """
    return sdk_dec.quo(liquidity, (sdk_dec.quo(liquidity, sqrt_price_current) + token_in))


def get_token_out(liquidity: Decimal, sqrt_price_current: Decimal, sqrt_price_next: Decimal) -> Decimal:
    """ Returns the token out when swapping token zero for one given the token in.
    """
    return sdk_dec.mul(liquidity, (sqrt_price_current - sqrt_price_next))


def get_token_in_swap_in_given_out(liquidity: Decimal, sqrt_price_current: Decimal, sqrt_price_next: Decimal) -> Decimal:
    """ Returns the token in when swapping token zero for one given the token out.
    In this case, the calculation is the same as computing token out when given token in.
    """
    return get_token_out(liquidity, sqrt_price_current, sqrt_price_next)


<< << << < HEAD
def calc_amount_zero_delta(liquidity: Decimal, sqrt_price_current: Decimal, sqrt_price_next: Decimal, should_round_up: bool) -> Decimal:


== == == =
<< << << < HEAD
def calc_amount_zero_delta(liquidity: sp.Float, sqrt_price_current: sp.Float, sqrt_price_next: sp.Float, should_round_up: bool):


== == == =


def get_expected_token_in(liquidity: sp.Float, sqrt_price_current: sp.Float, sqrt_price_next: sp.Float):


>>>>>> > 72d3e2d6c(Added all wolfram)
>>>>>> > 6764bd4a3(Added all wolfram)
""" Returns the expected token in when swapping token zero for one.
    """
mul1 = sdk_dec.mul(liquidity, (sqrt_price_current - sqrt_price_next))
 print(mul1)
  result = sdk_dec.quo(sdk_dec.mul(liquidity, (sqrt_price_current -
                                                sqrt_price_next)), sdk_dec.mul(sqrt_price_current, sqrt_price_next))
   if should_round_up:
        return sdk_dec.new(str(math.ceil(result)))
    return result

<< << << < HEAD


def calc_test_case_out_given_in(liquidity: Decimal, sqrt_price_current: Decimal, token_in_remaining: Decimal, swap_fee: Decimal) -> Tuple[Decimal, Decimal, Decimal]:
    """ Computes and prints all zero for one test case parameters. Next sqrt price is computed from the given parameters.
=======
<<<<<<< HEAD
def calc_test_case_out_given_in(liquidity: sp.Float, sqrt_price_current: sp.Float, token_in_remaining: sp.Float, swap_fee: sp.Float) -> Tuple[sp.Float, sp.Float, sp.Float]:
=======
>>>>>>> 6764bd4a3 (Added all wolfram)

def calc_test_case(liquidity: sp.Float, sqrt_price_current: sp.Float, token_in: sp.Float, swap_fee: sp.Float) -> Tuple[sp.Float, sp.Float, sp.Float]:
>>>>>>> 72d3e2d6c (Added all wolfram)
    """ Computes and prints all zero for one test case parameters. Next sqrt price is computed from the given parameters.
    Returns the next square root price, token out and fee amount per share.
    """
<<<<<<< HEAD
    token_in_remaining_after_fee = sdk_dec.mul(token_in_remaining, (sdk_dec.one - swap_fee))
    print(F"token_in_remaining_after_fee: {token_in_remaining_after_fee}")
=======
<<<<<<< HEAD
    token_in_after_fee = token_in_remaining * (1 - swap_fee)
    token_in_after_fee_rounded_up = sp.ceiling(token_in_after_fee)
>>>>>>> 6764bd4a3 (Added all wolfram)

    sqrt_price_next = get_next_sqrt_price(liquidity, sqrt_price_current, token_in_remaining_after_fee)
   
    token_in_after_fee_rounded_up = calc_amount_zero_delta(liquidity, sqrt_price_current, sqrt_price_next, True)

    print(F"token_in_after_fee_rounded_up: {token_in_after_fee_rounded_up}")

    token_out = get_token_out(liquidity, sqrt_price_current, sqrt_price_next)
    
<<<<<<< HEAD
    fee_charge_total = sdk_dec.zero
    if swap_fee > sdk_dec.zero:
        fee_charge_total = token_in_remaining - token_in_after_fee_rounded_up
    fee_amount_per_share = sdk_dec.quo(fee_charge_total, liquidity)
=======
    fee_charge_total = token_in_remaining - token_in_after_fee_rounded_up
    fee_amount_per_share = fee_charge_total / liquidity
=======
    sqrt_price_next = get_next_sqrt_price(
        liquidity, sqrt_price_current, token_in * (1 - swap_fee))
    token_out = get_token_out(liquidity, sqrt_price_current, sqrt_price_next)
    fee_amount_per_share = get_fee_amount_per_share(
        token_in, swap_fee, liquidity)
>>>>>>> 72d3e2d6c (Added all wolfram)
>>>>>>> 6764bd4a3 (Added all wolfram)

    print(F"current sqrt price: {sqrt_price_current}")
    print(F"sqrt_price_next: {sqrt_price_next}")
    print(F"liquidity: {liquidity}")
    print(F"token_out: {token_out}")
    print(F"fee_charge_total: {fee_charge_total}")
    print(F"fee_amount_per_share: {fee_amount_per_share}")

    return sqrt_price_next, token_out, fee_amount_per_share

<<<<<<< HEAD
def calc_test_case_in_given_out(liquidity: Decimal, sqrt_price_current: Decimal, token_out_remaining: Decimal, swap_fee: Decimal) -> Tuple[Decimal, Decimal, Decimal]:
    """ Computes and prints all zero for one test case parameters. Next sqrt price is computed from the given parameters.


== == == =
<< << << < HEAD
def calc_test_case_in_given_out(liquidity: sp.Float, sqrt_price_current: sp.Float, token_out_remaining: sp.Float, swap_fee: sp.Float) -> Tuple[sp.Float, sp.Float, sp.Float]:


== == == =
>>>>>> > 6764bd4a3(Added all wolfram)


def calc_test_case_in_given_out(liquidity: sp.Float, sqrt_price_current: sp.Float, token_in: sp.Float, swap_fee: sp.Float) -> Tuple[sp.Float, sp.Float, sp.Float]:


>>>>>> > 72d3e2d6c(Added all wolfram)
""" Computes and prints all zero for one test case parameters. Next sqrt price is computed from the given parameters.
    Returns the next square root price, token out and fee amount per share.
    """
<< << << < HEAD
sqrt_price_next = get_next_sqrt_price(
    liquidity, sqrt_price_current, token_out_remaining)
token_in = get_token_in_swap_in_given_out(
     liquidity, sqrt_price_current, sqrt_price_next)
<< << << < HEAD
fee_amount_per_share = sdk_dec.quo(sdk_dec.mul(token_in, swap_fee), liquidity)
== == == =
fee_amount_per_share = get_fee_amount_per_share(token_in, swap_fee, liquidity)
== == == =
sqrt_price_next = get_next_sqrt_price(
    liquidity, sqrt_price_current, token_in)
token_in = get_token_in_swap_in_given_out(
     liquidity, sqrt_price_current, sqrt_price_next)
 fee_amount_per_share = get_fee_amount_per_share(
      token_in, swap_fee, liquidity)
>>>>>> > 72d3e2d6c(Added all wolfram)
>>>>>> > 6764bd4a3(Added all wolfram)
token_in_after_fee = token_in * (1 + swap_fee)

print(F"current sqrt price: {sqrt_price_current}")
 print(F"sqrt_price_next: {sqrt_price_next}")
  print(F"liquidity: {liquidity}")
   print(F"token_in_after_fee: {token_in_after_fee}")
    print(F"fee_amount_per_share: {fee_amount_per_share}")

    return sqrt_price_next, token_in_after_fee, fee_amount_per_share

<< << << < HEAD
def calc_test_case_with_next_sqrt_price_out_given_in(liquidity: Decimal, sqrt_price_current: Decimal, sqrt_price_next: Decimal, swap_fee: Decimal) -> Tuple[Decimal, Decimal, Decimal]:


== == == =
<< << << < HEAD
def calc_test_case_with_next_sqrt_price_out_given_in(liquidity: sp.Float, sqrt_price_current: sp.Float, sqrt_price_next: sp.Float, swap_fee: sp.Float) -> Tuple[sp.Float, sp.Float, sp.Float]:


== == == =


def calc_test_case_with_next_sqrt_price(liquidity: sp.Float, sqrt_price_current: sp.Float, sqrt_price_next: sp.Float, swap_fee: sp.Float) -> Tuple[sp.Float, sp.Float, sp.Float]:


>>>>>> > 72d3e2d6c(Added all wolfram)
>>>>>> > 6764bd4a3(Added all wolfram)
""" Computes and prints all zero for one test case parameters when next square root price is known.

    Returns the expected token in, token out and fee amount per share. 
    """
<< << << < HEAD
expected_token_in_before_fee = calc_amount_zero_delta(
    liquidity, sqrt_price_current, sqrt_price_next, True)
<< << << < HEAD
expected_fee = sdk_dec.quo(sdk_dec.mul(
    expected_token_in_before_fee, swap_fee), (sdk_dec.one - swap_fee))
expected_token_in = expected_token_in_before_fee + expected_fee

 token_out = get_token_out(liquidity, sqrt_price_current, sqrt_price_next)
  fee_amount_per_share = sdk_dec.quo(expected_fee, liquidity)
== == == =
== == == =
expected_token_in_before_fee = get_expected_token_in(
    liquidity, sqrt_price_current, sqrt_price_next)
>>>>>> > 72d3e2d6c(Added all wolfram)
expected_token_in = expected_token_in_before_fee * (1 + swap_fee)

token_out = get_token_out(liquidity, sqrt_price_current, sqrt_price_next)
 fee_amount_per_share = get_fee_amount_per_share(
      expected_token_in_before_fee, swap_fee, liquidity)
>>>>>> > 6764bd4a3(Added all wolfram)

print(F"current sqrt price: {sqrt_price_current}")
print(F"given sqrt_price_next: {sqrt_price_next}")
 print(F"liquidity: {liquidity}")
  print(F"expected_token_in_before_fee: {expected_token_in_before_fee}")
   print(F"expected_token_in: {expected_token_in}")
    print(F"token_out: {token_out}")
    print(F"fee_charge_total: {expected_fee}")
    print(F"fee_amount_per_share: {fee_amount_per_share}")

    return expected_token_in, token_out, fee_amount_per_share

<< << << < HEAD
def calc_test_case_with_next_sqrt_price_in_given_out(liquidity: Decimal, sqrt_price_current: Decimal, sqrt_price_next: Decimal, swap_fee: Decimal) -> Tuple[Decimal, Decimal, Decimal]:


== == == =


def calc_test_case_with_next_sqrt_price_in_given_out(liquidity: sp.Float, sqrt_price_current: sp.Float, sqrt_price_next: sp.Float, swap_fee: sp.Float) -> Tuple[sp.Float, sp.Float, sp.Float]:


>>>>>> > 6764bd4a3(Added all wolfram)
""" Computes and prints all zero for one test case parameters when next square root price is known.
    Assems swapping token for token in given out.

    Returns the expected token out, token in and fee amount per share. 
    """
<< << << < HEAD
expected_token_out = calc_amount_zero_delta(
    liquidity, sqrt_price_current, sqrt_price_next, True)
== == == =
expected_token_out = get_expected_token_in(
    liquidity, sqrt_price_current, sqrt_price_next)
>>>>>> > 72d3e2d6c(Added all wolfram)

<< << << < HEAD
token_in = get_token_in_swap_in_given_out(
    liquidity, sqrt_price_current, sqrt_price_next)
fee_amount_per_share = sdk_dec.quo(
     sdk_dec.mul(token_in, swap_fee), liquidity)
== == == =
token_in = get_token_in_swap_in_given_out(
    liquidity, sqrt_price_current, sqrt_price_next)
fee_amount_per_share = get_fee_amount_per_share(
     token_in, swap_fee, liquidity)
>>>>>> > 6764bd4a3(Added all wolfram)
token_in_after_fee = token_in * (1 + swap_fee)

print(F"current sqrt price: {sqrt_price_current}")
 print(F"given sqrt_price_next: {sqrt_price_next}")
  print(F"liquidity: {liquidity}")
   print(F"expected_token_out: {expected_token_out}")
    print(F"token_in_after_fee: {token_in_after_fee}")
    print(F"fee_amount_per_share: {fee_amount_per_share}")

    return expected_token_out, token_in_after_fee, fee_amount_per_share
