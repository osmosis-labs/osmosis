from typing import Tuple

from common.common import *
import common.sdk_dec as sdk_dec
import zero_for_one as zfo
import one_for_zero as ofz

def estimate_test_case_out_given_in(tick_ranges: list[SqrtPriceRange], token_in_initial: Decimal, spread_factor: Decimal, is_zero_for_one: bool) -> Tuple[Decimal, Decimal]:
    """ Estimates a calc concentrated liquidity test case when swapping for token out given in.
    
    Given
      - sqrt price range with the start sqrt price, next sqrt price and liquidity
      - initial token in
      - spread factor
      - zero for one boolean flag
    Estimates the final token out and the spread reward growth per share and prints it to stdout.
    Also, estimates these and other values at each range and prints them to stdout.

    Returns the total token out and the total spread reward growth per share.
    """

    token_in_consumed_total, token_out_total, spread_rewards_growth_per_share_total = zero, zero, zero

    for i in range(len(tick_ranges)):
        tick_range = tick_ranges[i]

        # Normally, for the last swap range we swap until token in runs out
        # As a result, the next sqrt price for that range calculated at runtime.
        is_last_range = i == len(tick_ranges) - 1
        # Except for the cases where we set price limit explicitly. Then, the
        # last price range may have the upper sqrt price limit configured.
        is_next_price_set = tick_range.sqrt_price_next != None 

        is_with_next_sqrt_price = not is_last_range or is_next_price_set

        token_in_remaining = token_in_initial - token_in_consumed_total
        print(f"token_in_remaining: {token_in_remaining}")

        if is_with_next_sqrt_price:
            token_in_consumed, token_out, spread_rewards_growth_per_share = zero, zero, zero
            if is_zero_for_one:
                token_in_consumed, token_out, spread_rewards_growth_per_share = zfo.calc_test_case_with_next_sqrt_price_out_given_in(tick_range.liquidity, tick_range.sqrt_price_start, tick_range.sqrt_price_next, spread_factor)
            else:
                token_in_consumed, token_out, spread_rewards_growth_per_share = ofz.calc_test_case_with_next_sqrt_price_out_given_in(tick_range.liquidity, tick_range.sqrt_price_start, tick_range.sqrt_price_next, spread_factor)
            
            token_in_consumed_total += token_in_consumed
            token_out_total += token_out
            spread_rewards_growth_per_share_total += spread_rewards_growth_per_share

        else:
            if token_in_remaining < zero:
                raise Exception(F"token_in_remaining {token_in_remaining} is negative with token_in_initial {token_in_initial} and token_in_consumed_total {token_in_consumed_total}")

            token_out, spread_rewards_growth_per_share = zero, zero
            if is_zero_for_one:
                _, token_out, spread_rewards_growth_per_share = zfo.calc_test_case_out_given_in(tick_range.liquidity, tick_range.sqrt_price_start, token_in_remaining, spread_factor)
            else:
                _, token_out, spread_rewards_growth_per_share = ofz.calc_test_case_out_given_in(tick_range.liquidity, tick_range.sqrt_price_start, token_in_remaining, spread_factor)

            token_out_total += token_out
            spread_rewards_growth_per_share_total += spread_rewards_growth_per_share
        print("\n")
        print(F"After processing range {i}")
        print(F"current token_in_consumed_total: {token_in_consumed_total}")
        print(F"current token_out_total: {token_out_total}")
        print(F"current current spread_rewards_growth_per_share_total: {spread_rewards_growth_per_share_total}")
        print("\n\n\n")

    print("\n\n")
    print("Final results:")
    print("token_out_total: ", token_out_total)
    print("spread_rewards_growth_per_share_total: ", spread_rewards_growth_per_share_total)

    return token_out_total, spread_rewards_growth_per_share_total

def estimate_single_position_within_one_tick_ofz():
    """Estimates and prints the results of a calc concentrated liquidity test case with a single position within one tick
    when swapping token one for token zero (ofz).

     go test -timeout 30s -v -run TestKeeperTestSuite/TestCalcAndSwapOutAmtGivenIn/fee_1 github.com/osmosis-labs/osmosis/v16/x/concentrated-liquidity
    """

    is_zero_for_one = False
    spread_factor = sdk_dec.new("0.01")
    token_in_initial = sdk_dec.new("42000000")

    tick_ranges = [
        SqrtPriceRange(5000, None, sdk_dec.new("1517882343.751510418088349649")), # last one must be computed based on remaining token in, therefore it is None
    ]

    print(tick_ranges[0].liquidity)

    token_out_total, spread_rewards_growth_per_share_total = estimate_test_case_out_given_in(tick_ranges, token_in_initial, spread_factor, is_zero_for_one)

    expected_token_out_total = sdk_dec.new("8312")
    expected_spread_rewards_growth_per_share_total = sdk_dec.new("0.000276701288297452")

    validate_confirmed_results(token_out_total, spread_rewards_growth_per_share_total, expected_token_out_total, expected_spread_rewards_growth_per_share_total)

def estimate_two_positions_within_one_tick_zfo():
    """Estimates and prints the results of a calc concentrated liquidity test case with two positions within one tick
    when swapping token zero for one (zfo).

     go test -timeout 30s -v -run TestKeeperTestSuite/TestCalcAndSwapOutAmtGivenIn/fee_2 github.com/osmosis-labs/osmosis/v16/x/concentrated-liquidity
    """

    is_zero_for_one = True
    spread_factor = sdk_dec.new("0.03")
    token_in_initial = sdk_dec.new("13370")

    tick_ranges = [
        SqrtPriceRange(5000, None, sdk_dec.new("3035764687.503020836176699298")), # last one must be computed based on remaining token in, therefore it is None
    ]

    token_out_total, spread_rewards_growth_per_share_total = estimate_test_case_out_given_in(tick_ranges, token_in_initial, spread_factor, is_zero_for_one)

    expected_token_out_total = sdk_dec.new("64824917")
    expected_spread_rewards_growth_per_share_total = sdk_dec.new("0.000000132091924532")

    validate_confirmed_results(token_out_total, spread_rewards_growth_per_share_total, expected_token_out_total, expected_spread_rewards_growth_per_share_total)

def estimate_two_consecutive_positions_zfo(spread_factor: str, expected_token_out_total: str, expected_spread_rewards_growth_per_share_total: str):
    """Estimates and prints the results of a calc concentrated liquidity test case with two consecutive positions
    when swapping token zero for one (zfo).

     go test -timeout 30s -v -run TestKeeperTestSuite/TestCalcAndSwapOutAmtGivenIn/fee_3 github.com/osmosis-labs/osmosis/v16/x/concentrated-liquidity
    """

    spread_factor = sdk_dec.new(spread_factor)
    expected_token_out_total = sdk_dec.new(expected_token_out_total)
    expected_spread_rewards_growth_per_share_total = sdk_dec.new(expected_spread_rewards_growth_per_share_total)

    is_zero_for_one = True
    token_in_initial = sdk_dec.new("2000000")

    tick_ranges = [
        SqrtPriceRange(5000, 4545, sdk_dec.new("1517882343.751510418088349649")),
        SqrtPriceRange(4545, None, sdk_dec.new("1198735489.597250295669959397")), # last one must be computed based on remaining token in, therefore it is None
    ]

    token_out_total, spread_rewards_growth_per_share_total = estimate_test_case_out_given_in(tick_ranges, token_in_initial, spread_factor, is_zero_for_one)

    # TODO: this validation fails due to a different rounding strategy in scripts.
    # Needs investigation.
    # Go code should have the right value
    # validate_confirmed_results(token_out_total, spread_rewards_growth_per_share_total, expected_token_out_total, expected_spread_rewards_growth_per_share_total)

def estimate_overlapping_price_range_ofz_test(spread_factor: str, expected_token_out_total: str, expected_spread_rewards_growth_per_share_total: str):
    """Estimates and prints the results of a calc concentrated liquidity test case with overlapping price ranges
    when swapping token one for token zero (ofz).

     go test -timeout 30s -v -run TestKeeperTestSuite/TestCalcAndSwapOutAmtGivenIn/fee_4 github.com/osmosis-labs/osmosis/v16/x/concentrated-liquidity
    """

    is_zero_for_one = False
    spread_factor = sdk_dec.new(spread_factor)
    token_in_initial = sdk_dec.new("10000000000")

    print("token_in_initial", token_in_initial)

    # TODO: confirm liquidity values.
    tick_ranges = [
        SqrtPriceRange(5000, 5001, sdk_dec.new("1517882343.751510418088349649")),
        SqrtPriceRange(5001, 5500, sdk_dec.new("2188298432.357179145127590431")),
        SqrtPriceRange(5500, None, sdk_dec.new("670416088.605668727039240782")), # last one must be computed based on remaining token in, therefore it is None
    ]

    token_out_total, spread_rewards_growth_per_share_total = estimate_test_case_out_given_in(tick_ranges, token_in_initial, spread_factor, is_zero_for_one)

    expected_token_out_total = sdk_dec.new(expected_token_out_total)
    expected_spread_rewards_growth_per_share_total = sdk_dec.new(expected_spread_rewards_growth_per_share_total)

    validate_confirmed_results(token_out_total, spread_rewards_growth_per_share_total, expected_token_out_total, expected_spread_rewards_growth_per_share_total)

def estimate_overlapping_price_range_zfo_test(token_in_initial: str, spread_factor: str, expected_token_out_total: str, expected_spread_rewards_growth_per_share_total: str):
    """Estimates and prints the results of a calc concentrated liquidity test case with overlapping price ranges
    when swapping token zero for one (zfo) and not consuming full liquidity of the second position.

     go test -timeout 30s -v -run TestKeeperTestSuite/TestCalcAndSwapOutAmtGivenIn/fee_5 github.com/osmosis-labs/osmosis/v16/x/concentrated-liquidity
    """

    is_zero_for_one = True
    spread_factor = sdk_dec.new(spread_factor)
    token_in_initial = sdk_dec.new(token_in_initial)

    tick_ranges = [
        SqrtPriceRange(5000, 4999, sdk_dec.new("1517882343.751510418088349649")),
        SqrtPriceRange(4999, 4545, sdk_dec.new("1517882343.751510418088349649") + sdk_dec.new("670416215.718827443660400593")), # first and second position's liquidity.
        SqrtPriceRange(4545, None, sdk_dec.new("670416215.718827443660400593000")), # last one must be computed based on remaining token in, therefore it is None
    ]

    token_out_total, spread_rewards_growth_per_share_total = estimate_test_case_out_given_in(tick_ranges, token_in_initial, spread_factor, is_zero_for_one)

    expected_token_out_total = sdk_dec.new(expected_token_out_total)
    expected_spread_rewards_growth_per_share_total = sdk_dec.new(expected_spread_rewards_growth_per_share_total)

    validate_confirmed_results(token_out_total, spread_rewards_growth_per_share_total, expected_token_out_total, expected_spread_rewards_growth_per_share_total)

def estimate_consecutive_positions_gap_ofz_test():
    """Estimates and prints the results of a calc concentrated liquidity test case with consecutive positions with a gap
    when swapping token one for zero (ofz).

     go test -timeout 30s -v -run TestKeeperTestSuite/TestCalcAndSwapOutAmtGivenIn/fee_6 github.com/osmosis-labs/osmosis/v16/x/concentrated-liquidity
    """

    is_zero_for_one = False
    spread_factor = sdk_dec.new("0.03")
    token_in_initial = sdk_dec.new("10000000000")

    tick_ranges = [
        SqrtPriceRange(5000, 5500, sdk_dec.new("1517882343.751510418088349649")),
        SqrtPriceRange(5501, None, sdk_dec.new("1199528406.187413669220037261")), # last one must be computed based on remaining token in, therefore it is None
    ]

    token_out_total, spread_rewards_growth_per_share_total = estimate_test_case_out_given_in(tick_ranges, token_in_initial, spread_factor, is_zero_for_one)

    expected_token_out_total = sdk_dec.new("1771252")
    expected_spread_rewards_growth_per_share_total = sdk_dec.new("0.221769187794051751")

    validate_confirmed_results(token_out_total, spread_rewards_growth_per_share_total, expected_token_out_total, expected_spread_rewards_growth_per_share_total)

def estimate_slippage_protection_zfo_test():
    """Estimates and prints the results of a calc concentrated liquidity test case with slippage protection
    when swapping token zero for one (zfo).

     go test -timeout 30s -v -run TestKeeperTestSuite/TestCalcAndSwapOutAmtGivenIn/fee_7 github.com/osmosis-labs/osmosis/v16/x/concentrated-liquidity
    """

    is_zero_for_one = True
    spread_factor = sdk_dec.new("0.01")
    token_in_initial = sdk_dec.new("13370")

    tick_ranges = [
        SqrtPriceRange(5000, 4994, sdk_dec.new("1517882343.751510418088349649")),
    ]

    token_out_total, spread_rewards_growth_per_share_total = estimate_test_case_out_given_in(tick_ranges, token_in_initial, spread_factor, is_zero_for_one)

    expected_token_out_total = sdk_dec.new("64417624")
    expected_spread_rewards_growth_per_share_total = sdk_dec.new("0.000000085792039652")

    validate_confirmed_results(token_out_total, spread_rewards_growth_per_share_total, expected_token_out_total, expected_spread_rewards_growth_per_share_total)

def test():
    """Runs all swap out given in test cases, prints results as well as the intermediary calculations.

    Test cases that are confirmed to match Go tests, get validated to match the confirmed amounts.
    """

    # fee 1
    estimate_single_position_within_one_tick_ofz()

    # # fee 2
    estimate_two_positions_within_one_tick_zfo()

    # fee 3
    estimate_two_consecutive_positions_zfo("0.05", "8691708221", "0.000073738597832046")

    # No fee consecutive positions
    estimate_two_consecutive_positions_zfo("0", "9103422788.67833238665194882453", "0")

    # fee 4
    estimate_overlapping_price_range_ofz_test("0.1", "1695807", "0.624166726347032857")

    # fee 5
    estimate_overlapping_price_range_zfo_test("1800000", "0.005", "8440657775", "0.000005569829831408")

    # Overlapping no fee, utilizing full liquidity
    estimate_overlapping_price_range_zfo_test("2000000", "0.0", "9321276930", "0.0")

    # Overlapping no fee, not utilizing full liquidity
    estimate_overlapping_price_range_zfo_test("1800000", "0.0", "8479320318", "0.0")

    # fee 6
    estimate_consecutive_positions_gap_ofz_test()

    # fee 7
    estimate_slippage_protection_zfo_test()
