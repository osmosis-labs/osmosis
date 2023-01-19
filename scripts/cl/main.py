from common import *
import zero_for_one as zfo
import one_for_zero as ofz

def main():
    swap_fee = sdk_dec("0.05")
    liquidity = sdk_dec("1517882343.751510418088349649")
    sqrt_price_current = sdk_dec("70.710678118654752440")
    sqrt_price_next = sdk_dec("67.416615162732695594")
    token_in = sdk_dec("898695.642826782932516526784010")
    is_zero_for_one = True

    if is_zero_for_one:
        zfo.calc_test_case_with_next_sqrt_price(liquidity, sqrt_price_current, sqrt_price_next, swap_fee)
        # zfo.calc_test_case(liquidity, sqrt_price_current, token_in, swap_fee)
    else:
        ofz.calc_test_case(liquidity, sqrt_price_current, token_in, swap_fee)

if __name__ == "__main__":
    main()
