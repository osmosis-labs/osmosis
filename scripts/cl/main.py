from typing import Tuple
from common import *
import zero_for_one as zfo
import one_for_zero as ofz

def estimate_test_case(tick_ranges: list[SqrtPriceRange], token_in_initial: sp.Float, swap_fee: sp.Float, is_zero_for_one: bool) -> Tuple[sp.Float, sp.Float]:
    """ Estimates a calc concentrated liquidity test case.
    
    Given
      - sqrt price range with the start sqrt price, next sqrt price and liquidity
      - initial token in
      - swap fee
      - zero for one boolean flag
    Estimates the final token out and the fee growth per share and prints it to stdout.
    Also, estimates these and other values at each range and prints them to stdout.

    Returns the total token out and the total fee growth per share.
    """

    token_in_consumed_total, token_out_total, fee_growth_per_share_total = zero, zero, zero

    for i in range(len(tick_ranges)):
        tick_range = tick_ranges[i]

        is_with_next_sqrt_price = i != len(tick_ranges) - 1

        if is_with_next_sqrt_price:
            token_in_consumed, token_out, fee_growth_per_share = zero, zero, zero
            if is_zero_for_one:
                token_in_consumed, token_out, fee_growth_per_share = zfo.calc_test_case_with_next_sqrt_price(tick_range.liquidity, tick_range.sqrt_price_start, tick_range.sqrt_price_next, swap_fee)
            else:
                token_in_consumed, token_out, fee_growth_per_share = ofz.calc_test_case_with_next_sqrt_price(tick_range.liquidity, tick_range.sqrt_price_start, tick_range.sqrt_price_next, swap_fee)
            
            token_in_consumed_total += token_in_consumed
            token_out_total += token_out
            fee_growth_per_share_total += fee_growth_per_share

        else:
            token_in_remaining = token_in_initial - token_in_consumed_total

            if token_in_remaining < zero:
                raise Exception(F"token_in_remaining {token_in_remaining} is negative with token_in_initial {token_in_initial} and token_in_consumed_total {token_in_consumed_total}")

            token_out, fee_growth_per_share = zero, zero
            if is_zero_for_one:
                _, token_out, fee_growth_per_share = zfo.calc_test_case(tick_range.liquidity, tick_range.sqrt_price_start, token_in_remaining, swap_fee)


            else:
                _, token_out, fee_growth_per_share = ofz.calc_test_case(tick_range.liquidity, tick_range.sqrt_price_start, token_in_remaining, swap_fee)

            token_out_total += token_out
            fee_growth_per_share_total += fee_growth_per_share
        print("\n")
        print(F"After processing range {i}")
        print(F"current token_out_total: {token_out_total}")
        print(F"current current fee_growth_per_share_total: {fee_growth_per_share_total}")
        print("\n\n\n")

    print("\n\n")
    print("Final results:")
    print("token_out_total: ", token_out_total)
    print("fee_growth_per_share_total: ", fee_growth_per_share_total)

    return token_out_total, fee_growth_per_share_total

def estimate_single_position_within_one_tick_ofz():
    """Estimates and prints the results of a calc concentrated liquidity test case with a single position within one tick.

     go test -timeout 30s -v -run TestKeeperTestSuite/TestCalcAndSwapOutAmtGivenIn/fee_1 github.com/osmosis-labs/osmosis/v14/x/concentrated-liquidity
    """

    is_zero_for_one = False
    swap_fee = sdk_dec("0.01")
    token_in_initial = sdk_dec("42000000")

    tick_ranges = [
        SqrtPriceRange(5000, None, sdk_dec("1517882343.751510418088349649")), # last one must be computed based on remaining token in, therefore it is None
    ]

    token_out_total, fee_growth_per_share_total = estimate_test_case(tick_ranges, token_in_initial, swap_fee, is_zero_for_one)

    expected_token_out_total = sdk_dec("8312.77961614650590788243077782")
    expected_fee_growth_per_share_total = sdk_dec("0.000276701288297452775064000000017")

    # This validation exists to make sure that subsequent changes to the code do not break the test
    if sp.N(token_out_total, 18) != sp.N(expected_token_out_total, 18):
        raise Exception(F"token_out_total {token_out_total} does not match expected_token_out_total {expected_token_out_total}")
    
    if sp.N(fee_growth_per_share_total, 18) != sp.N(expected_fee_growth_per_share_total, 18):
        raise Exception(F"fee_growth_per_share_total {fee_growth_per_share_total} does not match expected_fee_growth_per_share_total {expected_fee_growth_per_share_total}")

def estimate_overlapping_price_range_ofz_test():
    """Estimates and prints the results of a calc concentrated liquidity test case with overlapping price ranges
    when swapping token zero for one.

     go test -timeout 30s -v -run TestKeeperTestSuite/TestCalcAndSwapOutAmtGivenIn/fee_4 github.com/osmosis-labs/osmosis/v14/x/concentrated-liquidity
    """

    is_zero_for_one = False
    swap_fee = sdk_dec("0.1")
    token_in_initial = sdk_dec("10000000000")

    tick_ranges = [
        SqrtPriceRange(5000, 5001, sdk_dec("1517882343.751510418088349649")),
        SqrtPriceRange(5001, 5500, sdk_dec("2188298432.35717914512760058700")),
        SqrtPriceRange(5500, None, sdk_dec("670416088.605668727039250938")), # last one must be computed based on remaining token in, therefore it is None
    ]

    token_out_total, fee_growth_per_share_total = estimate_test_case(tick_ranges, token_in_initial, swap_fee, is_zero_for_one)

    expected_token_out_total = sdk_dec("1708743.47809184831586199935191")
    expected_fee_growth_per_share_total = sdk_dec("0.598328101473707318285291820984")

    # This validation exists to make sure that subsequent changes to the code do not break the test
    if sp.N(token_out_total, 18) != sp.N(expected_token_out_total, 18):
        raise Exception(F"token_out_total {token_out_total} does not match expected_token_out_total {expected_token_out_total}")
    
    if sp.N(fee_growth_per_share_total, 18) != sp.N(expected_fee_growth_per_share_total, 18):
        raise Exception(F"fee_growth_per_share_total {fee_growth_per_share_total} does not match expected_fee_growth_per_share_total {expected_fee_growth_per_share_total}")

def main():
    # fee 1
    estimate_single_position_within_one_tick_ofz()

    # fee 4
    #estimate_overlapping_price_range_ofz_test()

if __name__ == "__main__":
    main()
