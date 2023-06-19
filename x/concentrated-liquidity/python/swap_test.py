from decimal import Decimal, getcontext
from clmath import *

DefaultPoolLiq0 = 1000000
DefaultPoolLiq1 = 5000000000
DefaultLowerPrice     = Decimal(4545)
DefaultSqrtLowerPrice = DefaultLowerPrice.sqrt()
DefaultLowerTick      = (30545000)
DefaultUpperPrice     = Decimal(5500)
DefaultSqrtUpperPrice = DefaultUpperPrice.sqrt()
DefaultUpperTick      = 31500000
DefaultCurrPrice      = Decimal(5000)
DefaultCurrTick                                 = 31000000
DefaultCurrSqrtPrice                              = Decimal("70.710678118654752440")

DefaultLiquidity = Decimal("1517882343.751510418088349649")
correctDefaultLiquidity = get_liquidity_from_amounts(DefaultCurrSqrtPrice, DefaultSqrtLowerPrice, DefaultSqrtUpperPrice, DefaultPoolLiq0, DefaultPoolLiq1)
print("used default liquidity:\n", DefaultLiquidity, "\ncorrect default liquidity:\n", correctDefaultLiquidity)

class SecondPosition:
    # Define this class based on what fields secondPosition has.
     def __init__(self, denom: str, amount: Decimal):
        self.denom = denom
        self.amount = amount

class SwapTest:
    def init_in_given_out(token_arg: Coin,
                 price_limit: Decimal,
                 new_lower_price: Decimal,
                 new_upper_price: Decimal,
                 pool_liq_amount0: int = DefaultPoolLiq0,
                 pool_liq_amount1: int = DefaultPoolLiq1,
                 second_position_lower_price: Decimal = Decimal(0),
                 second_position_upper_price: Decimal = Decimal(0),
                 spread_factor: Decimal = Decimal(0),
                 expect_err: bool = False):
        if token_arg.denom == "usdc":
            token_other_arg = "eth"
        if token_arg.denom == "eth":
            token_other_arg = "usdc"
        return SwapTest(True, token_arg, token_other_arg, price_limit, spread_factor, second_position_lower_price, second_position_upper_price, new_lower_price, new_upper_price, pool_liq_amount0, pool_liq_amount1, expect_err)

    def __init__(self,
                 in_given_out: bool,
                 token_arg: Coin,
                 token_other_arg: str,
                 price_limit: Decimal,
                 spread_factor: Decimal,
                 second_position_lower_price: Decimal,
                 second_position_upper_price: Decimal,
                 new_lower_price: Decimal,
                 new_upper_price: Decimal,
                 pool_liq_amount0: int,
                 pool_liq_amount1: int,
                 expect_err: bool):
        if in_given_out:
            self.token_in = token_arg
            self.token_out_denom = token_other_arg
        else:
            self.token_out = token_arg
            self.token_in_denom = token_other_arg
        self.price_limit = price_limit
        self.spread_factor = spread_factor
        self.second_position_lower_price = second_position_lower_price
        self.second_position_upper_price = second_position_upper_price
        self.new_lower_price = new_lower_price
        self.new_upper_price = new_upper_price
        self.pool_liq_amount0 = pool_liq_amount0
        self.pool_liq_amount1 = pool_liq_amount1
        self.expect_err = expect_err        

    def is_single_position(self):
        return self.second_position_lower_price == Decimal(0) and self.second_position_upper_price == Decimal(0)

    def derive_expected_fields(self):
        # Here you would calculate and set all the "expected" fields
        # For example:
        # self.expected_token_in = calculate_expected_token_in()
        # self.expected_token_out = calculate_expected_token_out()
        # ...
        pass