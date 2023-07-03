from typing import Tuple

from common.common import *
import common.sdk_dec as sdk_dec
from decimal import *
import zero_for_one as zfo
import one_for_zero as ofz

def estimate_test_case_in_given_out(tick_ranges: list[SqrtPriceRange], token_out: Decimal, spread_factor: Decimal, is_zero_for_one: bool) -> Tuple[Decimal, Decimal]:
    """ Estimates a calc concentrated liquidity test case when swapping for token in given out.
    
    Given
      - sqrt price range with the start sqrt price, next sqrt price and liquidity
      - token out
      - spread factor
      - zero for one boolean flag
    Estimates the token in with fee applied and the spread reward growth per share and prints it to stdout.
    Also, estimates these and other values at each range and prints them to stdout.

    Returns the total token in and the total spread reward growth per share.
    """

    token_out_consumed_total, token_in_total, spread_rewards_growth_per_share_total = sdk_dec.zero, sdk_dec.zero, sdk_dec.zero

    for i in range(len(tick_ranges)):
        tick_range = tick_ranges[i]

        # Normally, for the last swap range we swap until token in runs out
        # As a result, the next sqrt price for that range calculated at runtime.
        is_last_range = i == len(tick_ranges) - 1
        # Except for the cases where we set price limit explicitly. Then, the
        # last price range may have the upper sqrt price limit configured.
        is_next_price_set = tick_range.sqrt_price_next != None 

        is_with_next_sqrt_price = not is_last_range or is_next_price_set

        if is_with_next_sqrt_price:
            token_out_consumed, token_in, spread_rewards_growth_per_share = sdk_dec.zero, sdk_dec.zero, sdk_dec.zero
            if is_zero_for_one:
                token_out_consumed, token_in, spread_rewards_growth_per_share = zfo.calc_test_case_with_next_sqrt_price_in_given_out(tick_range.liquidity, tick_range.sqrt_price_start, tick_range.sqrt_price_next, spread_factor)
            else:
                token_out_consumed, token_in, spread_rewards_growth_per_share = ofz.calc_test_case_with_next_sqrt_price_in_given_out(tick_range.liquidity, tick_range.sqrt_price_start, tick_range.sqrt_price_next, spread_factor)
            print(F"token_out_consumed {token_out_consumed}")
            print(F"token_in {token_in}")
            token_out_consumed_total += token_out_consumed
            token_in_total += token_in
            spread_rewards_growth_per_share_total += spread_rewards_growth_per_share

        else:
            token_out_remaining = token_out - token_out_consumed_total

            if token_out_remaining < sdk_dec.zero:
                raise Exception(F"token_in_remaining {token_out_remaining} is negative with token_out_initial {token_out} and token_out_consumed_total {token_out_consumed_total}")

            token_in, spread_rewards_growth_per_share = sdk_dec.zero, sdk_dec.zero
            if is_zero_for_one:
                _, token_in, spread_rewards_growth_per_share = zfo.calc_test_case_in_given_out(tick_range.liquidity, tick_range.sqrt_price_start, token_out_remaining, spread_factor)
            else:
                _, token_in, spread_rewards_growth_per_share = ofz.calc_test_case_in_given_out(tick_range.liquidity, tick_range.sqrt_price_start, token_out_remaining, spread_factor)

            token_in_total += token_in
            spread_rewards_growth_per_share_total += spread_rewards_growth_per_share
        print("\n")
        print(F"After processing range {i}")
        print(F"current token_in_total: {token_in_total}")
        print(F"current current spread_rewards_growth_per_share_total: {spread_rewards_growth_per_share_total}")
        print("\n\n\n")

    print("\n\n")
    print("Final results:")
    print("token_in_total: ", token_in_total)
    print("spread_rewards_growth_per_share_total: ", spread_rewards_growth_per_share_total)

    return token_in_total, spread_rewards_growth_per_share_total

def estimate_single_position_within_one_tick_ofz_in_given_out():
    """Estimates and prints the results of a calc concentrated liquidity test case with a single position within one tick
    when swapping token one for token zero (ofz).

     go test -timeout 30s -v -run TestKeeperTestSuite/TestCalcAndSwapInAmtGivenOut/fee_1 github.com/osmosis-labs/osmosis/v16/x/concentrated-liquidity
    """

    is_zero_for_one = False
    spread_factor = sdk_dec.new("0.01")
    token_out_initial = sdk_dec.new("42000000")

    tick_ranges = [
        SqrtPriceRange(5000, None, sdk_dec.new("1517882343.751510418088349649")), # last one must be computed based on remaining token in, therefore it is None
    ]

    token_in, spread_rewards_growth_per_share_total = estimate_test_case_in_given_out(tick_ranges, token_out_initial, spread_factor, is_zero_for_one)

    expected_token_in = sdk_dec.new("8481")
    expected_spread_rewards_growth_per_share_total = sdk_dec.new("0.000000055877384518")

    validate_confirmed_results(token_in, spread_rewards_growth_per_share_total, expected_token_in, expected_spread_rewards_growth_per_share_total)

def estimate_two_positions_within_one_tick_zfo_in_given_out():
    """Estimates and prints the results of a calc concentrated liquidity test case with two positions within one tick
    when swapping token zero for one (zfo).

     go test -timeout 30s -v -run TestKeeperTestSuite/TestCalcAndSwapInAmtGivenOut/fee_2 github.com/osmosis-labs/osmosis/v16/x/concentrated-liquidity
    """

    is_zero_for_one = True
    spread_factor = sdk_dec.new("0.03")
    token_out = sdk_dec.new("13370")

    tick_ranges = [
        SqrtPriceRange(5000, None, sdk_dec.new("3035764687.503020836176699298")), # last one must be computed based on remaining token in, therefore it is None
    ]

    token_in, spread_rewards_growth_per_share_total = estimate_test_case_in_given_out(tick_ranges, token_out, spread_factor, is_zero_for_one)

    expected_token_in = sdk_dec.new("68896070")
    expected_spread_rewards_growth_per_share_total = sdk_dec.new("0.000680843976677818")

    validate_confirmed_results(token_in, spread_rewards_growth_per_share_total, expected_token_in, expected_spread_rewards_growth_per_share_total)

def estimate_two_consecutive_positions_zfo_in_given_out(spread_factor: str, expected_token_in: str, expected_spread_rewards_growth_per_share_total: str):
    """Estimates and prints the results of a calc concentrated liquidity test case with two consecutive positions
    when swapping token zero for one (zfo).

     go test -timeout 30s -v -run TestKeeperTestSuite/TestCalcAndSwapInAmtGivenOut/fee_3 github.com/osmosis-labs/osmosis/v16/x/concentrated-liquidity
    """

    is_zero_for_one = True
    spread_factor = sdk_dec.new(spread_factor)
    token_out = sdk_dec.new("2000000")

    tick_ranges = [
        SqrtPriceRange(5000, 4545, sdk_dec.new("1517882343.751510418088349649")),
        SqrtPriceRange(4545, None, sdk_dec.new("1198735489.597250295669959398")), # last one must be computed based on remaining token in, therefore it is None
    ]

    token_in, spread_rewards_growth_per_share_total = estimate_test_case_in_given_out(tick_ranges, token_out, spread_factor, is_zero_for_one)

    expected_token_in = sdk_dec.new(expected_token_in)
    expected_spread_rewards_growth_per_share_total = sdk_dec.new(expected_spread_rewards_growth_per_share_total)

    validate_confirmed_results(token_in, spread_rewards_growth_per_share_total, expected_token_in, expected_spread_rewards_growth_per_share_total)

def estimate_overlapping_price_range_ofz_test_in_given_out():
    """Estimates and prints the results of a calc concentrated liquidity test case with overlapping price ranges
    when swapping token one for token zero (ofz).

     go test -timeout 30s -v -run TestKeeperTestSuite/TestCalcAndSwapInAmtGivenOut/fee_4 github.com/osmosis-labs/osmosis/v16/x/concentrated-liquidity
    """

    is_zero_for_one = False
    spread_factor = sdk_dec.new("0.1")
    token_out_initial = sdk_dec.new("10000000000")

    tick_ranges = [
        SqrtPriceRange(5000, 5001, sdk_dec.new("1517882343.751510418088349649")),
        SqrtPriceRange(5001, 5500, sdk_dec.new("2188298432.357179145127590431")),
        SqrtPriceRange(5500, None, sdk_dec.new("670416088.605668727039240782")), # last one must be computed based on remaining token in, therefore it is None
    ]

    token_in, spread_rewards_growth_per_share_total = estimate_test_case_in_given_out(tick_ranges, token_out_initial, spread_factor, is_zero_for_one)

    expected_token_in = sdk_dec.new("2071290")
    expected_spread_rewards_growth_per_share_total = sdk_dec.new("0.000143548203873862")

    validate_confirmed_results(token_in, spread_rewards_growth_per_share_total, expected_token_in, expected_spread_rewards_growth_per_share_total)

def estimate_overlapping_price_range_zfo_test_in_given_out(tokein_in_initial: str, spread_factor: str, expected_token_in: str, expected_spread_rewards_growth_per_share_total: str):
    """Estimates and prints the results of a calc concentrated liquidity test case with overlapping price ranges
    when swapping token zero for one (zfo) and not consuming full liquidity of the second position.

     go test -timeout 30s -v -run TestKeeperTestSuite/TestCalcAndSwapInAmtGivenOut/fee_5 github.com/osmosis-labs/osmosis/v16/x/concentrated-liquidity
    """

    is_zero_for_one = True
    spread_factor = sdk_dec.new(spread_factor)
    token_in_initial = sdk_dec.new(tokein_in_initial)

    tick_ranges = [
        SqrtPriceRange(5000, 4999, sdk_dec.new("1517882343.751510418088349649")),
        SqrtPriceRange(4999, 4545, sdk_dec.new("1517882343.751510418088349649") + sdk_dec.new("670416215.718827443660400594")), # first and second position's liquidity.
        SqrtPriceRange(4545, None, sdk_dec.new("670416215.718827443660400594")), # last one must be computed based on remaining token in, therefore it is None
    ]

    token_in, spread_rewards_growth_per_share_total = estimate_test_case_in_given_out(tick_ranges, token_in_initial, spread_factor, is_zero_for_one)

    expected_token_in = sdk_dec.new(expected_token_in)
    expected_spread_rewards_growth_per_share_total = sdk_dec.new(expected_spread_rewards_growth_per_share_total)

    validate_confirmed_results(token_in, spread_rewards_growth_per_share_total, expected_token_in, expected_spread_rewards_growth_per_share_total)

def estimate_consecutive_positions_gap_ofz_test_in_given_out():
    """Estimates and prints the results of a calc concentrated liquidity test case with consecutive positions with a gap
    when swapping token one for zero (ofz).

     go test -timeout 30s -v -run TestKeeperTestSuite/TestCalcAndSwapInAmtGivenOut/fee_6 github.com/osmosis-labs/osmosis/v16/x/concentrated-liquidity
    """

    is_zero_for_one = False
    spread_factor = sdk_dec.new("0.03")
    token_out_initial = sdk_dec.new("10000000000")

    tick_ranges = [
        SqrtPriceRange(5000, 5500, sdk_dec.new("1517882343.751510418088349649")),
        SqrtPriceRange(5501, None, sdk_dec.new("1199528406.187413669220037261")), # last one must be computed based on remaining token in, therefore it is None
    ]

    token_in, spread_rewards_growth_per_share_total = estimate_test_case_in_given_out(tick_ranges, token_out_initial, spread_factor, is_zero_for_one)

    expected_token_in = sdk_dec.new("1876851")
    expected_spread_rewards_growth_per_share_total = sdk_dec.new("0.000041537584780053")

    validate_confirmed_results(token_in, spread_rewards_growth_per_share_total, expected_token_in, expected_spread_rewards_growth_per_share_total)

def estimate_slippage_protection_zfo_test_in_given_out():
    """Estimates and prints the results of a calc concentrated liquidity test case with slippage protection
    when swapping token zero for one (zfo).

     go test -timeout 30s -v -run TestKeeperTestSuite/TestCalcAndSwapInAmtGivenOut/fee_7 github.com/osmosis-labs/osmosis/v16/x/concentrated-liquidity
    """

    is_zero_for_one = True
    spread_factor = sdk_dec.new("0.01")
    token_in_initial = sdk_dec.new("13370")

    tick_ranges = [
        SqrtPriceRange(5000, 4994, sdk_dec.new("1517882343.751510418088349649")),
    ]

    token_in, spread_rewards_growth_per_share_total = estimate_test_case_in_given_out(tick_ranges, token_in_initial, spread_factor, is_zero_for_one)

    expected_token_in = sdk_dec.new("65068308")
    expected_spread_rewards_growth_per_share_total = sdk_dec.new("0.000428678206421614")

    validate_confirmed_results(token_in, spread_rewards_growth_per_share_total, expected_token_in, expected_spread_rewards_growth_per_share_total)

def test():
    """Runs all swap in given out test cases, prints results as well as the intermediary calculations.

    Test cases that are confirmed to match Go tests, get validated to match the confirmed amounts.
    """

    # fee 1
    estimate_single_position_within_one_tick_ofz_in_given_out()

    # fee 2
    estimate_two_positions_within_one_tick_zfo_in_given_out()

    # fee 3
    estimate_two_consecutive_positions_zfo_in_given_out("0.05", "9582550303", "0.353536268175351249")

    # No fee consecutive positions zfo
    estimate_two_consecutive_positions_zfo_in_given_out("0.0", "9103422788", "0.0")

    # fee 4
    estimate_overlapping_price_range_ofz_test_in_given_out()

    # fee 5
    estimate_overlapping_price_range_zfo_test_in_given_out("1800000", "0.005", "8521929968", "0.026114888608913022")

    # No fee overlapping price range zfo, utilizing full liquidity
    estimate_overlapping_price_range_zfo_test_in_given_out("2000000", "0.0", "9321276930", "0.0")

    # No fee overlapping price range zfo, not utilizing full liquidity
    estimate_overlapping_price_range_zfo_test_in_given_out("1800000", "0.0", "8479320318", "0.0")

    # fee 6
    estimate_consecutive_positions_gap_ofz_test_in_given_out()

    # fee 7
    estimate_slippage_protection_zfo_test_in_given_out()
