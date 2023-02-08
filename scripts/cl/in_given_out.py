from typing import Tuple

from common import *
import zero_for_one as zfo
import one_for_zero as ofz

def estimate_test_case_in_given_out(tick_ranges: list[SqrtPriceRange], token_out: sp.Float, swap_fee: sp.Float, is_zero_for_one: bool) -> Tuple[sp.Float, sp.Float]:
    """ Estimates a calc concentrated liquidity test case when swapping for token in given out.
    
    Given
      - sqrt price range with the start sqrt price, next sqrt price and liquidity
      - token out
      - swap fee
      - zero for one boolean flag
    Estimates the token in with fee applied and the fee growth per share and prints it to stdout.
    Also, estimates these and other values at each range and prints them to stdout.

    Returns the total token in and the total fee growth per share.
    """

    token_out_consumed_total, token_in_total, fee_growth_per_share_total = zero, zero, zero

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
            token_out_consumed, token_in, fee_growth_per_share = zero, zero, zero
            if is_zero_for_one:
                token_out_consumed, token_in, fee_growth_per_share = zfo.calc_test_case_with_next_sqrt_price_in_given_out(tick_range.liquidity, tick_range.sqrt_price_start, tick_range.sqrt_price_next, swap_fee)
            else:
                token_out_consumed, token_in, fee_growth_per_share = ofz.calc_test_case_with_next_sqrt_price_in_given_out(tick_range.liquidity, tick_range.sqrt_price_start, tick_range.sqrt_price_next, swap_fee)
            print(F"token_out_consumed {token_out_consumed}")
            print(F"token_in {token_in}")
            token_out_consumed_total += token_out_consumed
            token_in_total += token_in
            fee_growth_per_share_total += fee_growth_per_share

        else:
            token_out_remaining = token_out - token_out_consumed_total

            if token_out_remaining < zero:
                raise Exception(F"token_in_remaining {token_out_remaining} is negative with token_out_initial {token_out} and token_out_consumed_total {token_out_consumed_total}")

            token_in, fee_growth_per_share = zero, zero
            if is_zero_for_one:
                _, token_in, fee_growth_per_share = zfo.calc_test_case_in_given_out(tick_range.liquidity, tick_range.sqrt_price_start, token_out_remaining, swap_fee)
            else:
                _, token_in, fee_growth_per_share = ofz.calc_test_case_in_given_out(tick_range.liquidity, tick_range.sqrt_price_start, token_out_remaining, swap_fee)

            token_in_total += token_in
            fee_growth_per_share_total += fee_growth_per_share
        print("\n")
        print(F"After processing range {i}")
        print(F"current token_in_total: {token_in_total}")
        print(F"current current fee_growth_per_share_total: {fee_growth_per_share_total}")
        print("\n\n\n")

    print("\n\n")
    print("Final results:")
    print("token_in_total: ", token_in_total)
    print("fee_growth_per_share_total: ", fee_growth_per_share_total)

    return token_in_total, fee_growth_per_share_total

def estimate_single_position_within_one_tick_ofz_in_given_out():
    """Estimates and prints the results of a calc concentrated liquidity test case with a single position within one tick
    when swapping token one for token zero (ofz).

     go test -timeout 30s -v -run TestKeeperTestSuite/TestCalcAndSwapInAmtGivenOut/fee_1 github.com/osmosis-labs/osmosis/v14/x/concentrated-liquidity
    """

    is_zero_for_one = False
    swap_fee = fixed_prec_dec("0.01")
    token_out_initial = fixed_prec_dec("42000000")

    tick_ranges = [
        SqrtPriceRange(5000, None, fixed_prec_dec("1517882343.751510418088349649")), # last one must be computed based on remaining token in, therefore it is None
    ]

    token_in, fee_growth_per_share_total = estimate_test_case_in_given_out(tick_ranges, token_out_initial, swap_fee, is_zero_for_one)

    expected_token_in = fixed_prec_dec("8480.68138458406954789169099991")
    expected_fee_growth_per_share_total = fixed_prec_dec("0.0000000553186106731409146737705304277")

    validate_confirmed_results(token_in, fee_growth_per_share_total, expected_token_in, expected_fee_growth_per_share_total)

def estimate_two_positions_within_one_tick_zfo_in_given_out():
    """Estimates and prints the results of a calc concentrated liquidity test case with two positions within one tick
    when swapping token zero for one (zfo).

     go test -timeout 30s -v -run TestKeeperTestSuite/TestCalcAndSwapInAmtGivenOut/fee_2 github.com/osmosis-labs/osmosis/v14/x/concentrated-liquidity
    """

    is_zero_for_one = True
    swap_fee = fixed_prec_dec("0.03")
    token_out = fixed_prec_dec("13370")

    tick_ranges = [
        SqrtPriceRange(5000, None, fixed_prec_dec("3035764687.503020836176699298")), # last one must be computed based on remaining token in, therefore it is None
    ]

    token_in, fee_growth_per_share_total = estimate_test_case_in_given_out(tick_ranges, token_out, swap_fee, is_zero_for_one)

    expected_token_in = fixed_prec_dec("68834063.6068587597543212771274")
    expected_fee_growth_per_share_total = fixed_prec_dec("0.000660418657377483623332014151904")

    validate_confirmed_results(token_in, fee_growth_per_share_total, expected_token_in, expected_fee_growth_per_share_total)

def estimate_two_consecutive_positions_zfo_in_given_out(swap_fee: str, expected_token_in: str, expected_fee_growth_per_share_total: str):
    """Estimates and prints the results of a calc concentrated liquidity test case with two consecutive positions
    when swapping token zero for one (zfo).

     go test -timeout 30s -v -run TestKeeperTestSuite/TestCalcAndSwapInAmtGivenOut/fee_3 github.com/osmosis-labs/osmosis/v14/x/concentrated-liquidity
    """

    is_zero_for_one = True
    swap_fee = fixed_prec_dec(swap_fee)
    token_out = fixed_prec_dec("2000000")

    tick_ranges = [
        SqrtPriceRange(5000, 4545, fixed_prec_dec("1517882343.751510418088349649")),
        SqrtPriceRange(4545, None, fixed_prec_dec("1198735489.597250295669959398")), # last one must be computed based on remaining token in, therefore it is None
    ]

    token_in, fee_growth_per_share_total = estimate_test_case_in_given_out(tick_ranges, token_out, swap_fee, is_zero_for_one)

    expected_token_in = fixed_prec_dec(expected_token_in)
    expected_fee_growth_per_share_total = fixed_prec_dec(expected_fee_growth_per_share_total)

    validate_confirmed_results(token_in, fee_growth_per_share_total, expected_token_in, expected_fee_growth_per_share_total)

def estimate_overlapping_price_range_ofz_test_in_given_out():
    """Estimates and prints the results of a calc concentrated liquidity test case with overlapping price ranges
    when swapping token one for token zero (ofz).

     go test -timeout 30s -v -run TestKeeperTestSuite/TestCalcAndSwapInAmtGivenOut/fee_4 github.com/osmosis-labs/osmosis/v14/x/concentrated-liquidity
    """

    is_zero_for_one = False
    swap_fee = fixed_prec_dec("0.1")
    token_out_initial = fixed_prec_dec("10000000000")

    tick_ranges = [
        SqrtPriceRange(5000, 5001, fixed_prec_dec("1517882343.751510418088349649")),
        SqrtPriceRange(5001, 5500, fixed_prec_dec("2188298432.357179145127590431")),
        SqrtPriceRange(5500, None, fixed_prec_dec("670416088.605668727039240782")), # last one must be computed based on remaining token in, therefore it is None
    ]

    token_in, fee_growth_per_share_total = estimate_test_case_in_given_out(tick_ranges, token_out_initial, swap_fee, is_zero_for_one)

    expected_token_in = fixed_prec_dec("2050578.06505516031487626102620")
    expected_fee_growth_per_share_total = fixed_prec_dec("0.000129193383486476217032589405019")

    validate_confirmed_results(token_in, fee_growth_per_share_total, expected_token_in, expected_fee_growth_per_share_total)

def estimate_overlapping_price_range_zfo_test_in_given_out(tokein_in_initial: str, swap_fee: str, expected_token_in: str, expected_fee_growth_per_share_total: str):
    """Estimates and prints the results of a calc concentrated liquidity test case with overlapping price ranges
    when swapping token zero for one (zfo) and not consuming full liquidity of the second position.

     go test -timeout 30s -v -run TestKeeperTestSuite/TestCalcAndSwapInAmtGivenOut/fee_5 github.com/osmosis-labs/osmosis/v14/x/concentrated-liquidity
    """

    is_zero_for_one = True
    swap_fee = fixed_prec_dec(swap_fee)
    token_in_initial = fixed_prec_dec(tokein_in_initial)

    tick_ranges = [
        SqrtPriceRange(5000, 4999, fixed_prec_dec("1517882343.751510418088349649")),
        SqrtPriceRange(4999, 4545, fixed_prec_dec("1517882343.751510418088349649") + fixed_prec_dec("670416215.718827443660400594")), # first and second position's liquidity.
        SqrtPriceRange(4545, None, fixed_prec_dec("670416215.718827443660400594")), # last one must be computed based on remaining token in, therefore it is None
    ]

    token_in, fee_growth_per_share_total = estimate_test_case_in_given_out(tick_ranges, token_in_initial, swap_fee, is_zero_for_one)

    expected_token_in = fixed_prec_dec(expected_token_in)
    expected_fee_growth_per_share_total = fixed_prec_dec(expected_fee_growth_per_share_total)

    validate_confirmed_results(token_in, fee_growth_per_share_total, expected_token_in, expected_fee_growth_per_share_total)

def estimate_consecutive_positions_gap_ofz_test_in_given_out():
    """Estimates and prints the results of a calc concentrated liquidity test case with consecutive positions with a gap
    when swapping token one for zero (ofz).

     go test -timeout 30s -v -run TestKeeperTestSuite/TestCalcAndSwapInAmtGivenOut/fee_6 github.com/osmosis-labs/osmosis/v14/x/concentrated-liquidity
    """

    is_zero_for_one = False
    swap_fee = fixed_prec_dec("0.03")
    token_out_initial = fixed_prec_dec("10000000000")

    tick_ranges = [
        SqrtPriceRange(5000, 5500, fixed_prec_dec("1517882343.751510418088349649")),
        SqrtPriceRange(5501, None, fixed_prec_dec("1199528406.187413669220037261")), # last one must be computed based on remaining token in, therefore it is None
    ]

    token_in, fee_growth_per_share_total = estimate_test_case_in_given_out(tick_ranges, token_out_initial, swap_fee, is_zero_for_one)

    expected_token_in = fixed_prec_dec("1875162.23481286214014647237476")
    expected_fee_growth_per_share_total = fixed_prec_dec("0.0000402914572366514266238726907754")

    validate_confirmed_results(token_in, fee_growth_per_share_total, expected_token_in, expected_fee_growth_per_share_total)

def estimate_slippage_protection_zfo_test_in_given_out():
    """Estimates and prints the results of a calc concentrated liquidity test case with slippage protection
    when swapping token zero for one (zfo).

     go test -timeout 30s -v -run TestKeeperTestSuite/TestCalcAndSwapInAmtGivenOut/fee_7 github.com/osmosis-labs/osmosis/v14/x/concentrated-liquidity
    """

    is_zero_for_one = True
    swap_fee = fixed_prec_dec("0.01")
    token_in_initial = fixed_prec_dec("13370")

    tick_ranges = [
        SqrtPriceRange(5000, 4994, fixed_prec_dec("1517882343.751510418088349649")),
    ]

    token_in, fee_growth_per_share_total = estimate_test_case_in_given_out(tick_ranges, token_in_initial, swap_fee, is_zero_for_one)

    expected_token_in = fixed_prec_dec("65061801.2370366020634290878154")
    expected_fee_growth_per_share_total = fixed_prec_dec("0.000424391424357398265504790467604")

    validate_confirmed_results(token_in, fee_growth_per_share_total, expected_token_in, expected_fee_growth_per_share_total)

def test():
    """Runs all swap in given out test cases, prints results as well as the intermediary calculations.

    Test cases that are confirmed to match Go tests, get validated to match the confirmed amounts.
    """

    # fee 1
    estimate_single_position_within_one_tick_ofz_in_given_out()

    # fee 2
    estimate_two_positions_within_one_tick_zfo_in_given_out()

    # fee 3
    estimate_two_consecutive_positions_zfo_in_given_out("0.05", "9558593928.11224900598454626594", "0.335859454766583687373296697168")

    # No fee consecutive positions zfo
    estimate_two_consecutive_positions_zfo_in_given_out("0.0", "9103422788.67833238665194882471", "0.0")

    # fee 4
    estimate_overlapping_price_range_ofz_test_in_given_out()

    # # fee 5
    estimate_overlapping_price_range_zfo_test_in_given_out("1800000", "0.005", "8521716920.24423119398212788491", "0.0259843141658684574040461818975")

    # No fee overlapping price range zfo, utilizing full liquidity
    estimate_overlapping_price_range_zfo_test_in_given_out("2000000", "0.0", "9321276930.73297863398988127324", "0.0")

    # No fee overlapping price range zfo, not utilizing full liquidity
    estimate_overlapping_price_range_zfo_test_in_given_out("1800000", "0.0", "8479320318.65097631242002774618", "0.0")

    # fee 6
    estimate_consecutive_positions_gap_ofz_test_in_given_out()

    # fee 7
    estimate_slippage_protection_zfo_test_in_given_out()
