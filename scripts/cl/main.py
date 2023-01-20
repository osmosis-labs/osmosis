from common import *
import zero_for_one as zfo
import one_for_zero as ofz

def main():
    swap_fee = sdk_dec("0.1")
    liquidity = sdk_dec("670416088.605668727039250938")
    sqrt_price_current = sdk_dec("74.161984870956629487")
    sqrt_price_next = sdk_dec("76.4422024931482315166509926684")
    token_in = sdk_dec("1697476281.47202717595504088493")
    is_zero_for_one = False
    is_with_next_sqrt_price = False

    if is_zero_for_one:
        if is_with_next_sqrt_price:
            zfo.calc_test_case_with_next_sqrt_price(liquidity, sqrt_price_current, sqrt_price_next, swap_fee)
        else:
            zfo.calc_test_case(liquidity, sqrt_price_current, token_in, swap_fee)
    else:
        if is_with_next_sqrt_price:
            ofz.calc_test_case_with_next_sqrt_price(liquidity, sqrt_price_current, sqrt_price_next, swap_fee)
        else:
            ofz.calc_test_case(liquidity, sqrt_price_current, token_in, swap_fee)

if __name__ == "__main__":
    main()
