from decimal import Decimal, ROUND_FLOOR, ROUND_CEILING, getcontext

class Coin:
    # Define this class based on what fields sdk.Coin has.
    def __init__(self, denom: str, amount: int):
        self.denom = denom
        self.amount = amount

class DecCoins:
    # Define this class based on what fields sdk.DecCoins has.
    def __init__(self, denom: str, amount: Decimal):
        self.denom = denom
        self.amount = amount

oneULPDec = Decimal(1) / Decimal(10 ** 18)
oneULPBigDec = Decimal(1) / Decimal(10 ** 36)

getcontext().prec = 60

# --- General rounding helper ---

def round_decimal(number: Decimal, places, rounding):
    """Round a Decimal to the given number of decimal places."""
    format_string = f'0.{"0" * places}'  # build a string like '0.00' for 2 places
    return number.quantize(Decimal(format_string), rounding=rounding)

# --- SDK precision based rounding helpers ---

def round_sdk_prec_down(number: Decimal):
    return round_decimal(number, 18, ROUND_FLOOR)

def round_sdk_prec_up(number: Decimal):
    return round_decimal(number, 18, ROUND_CEILING)

def round_osmo_prec_down(number: Decimal):
    return round_decimal(number, 36, ROUND_FLOOR)

def round_osmo_prec_up(number: Decimal):
    return round_decimal(number, 36, ROUND_CEILING)

# --- CL liquidity functions ---

def liquidity0(amount: int, sqrt_price_a: Decimal, sqrt_price_b: Decimal) -> Decimal:
    # Swap if sqrt_price_a is greater than sqrt_price_b
    if sqrt_price_a > sqrt_price_b:
        sqrt_price_a, sqrt_price_b = sqrt_price_b, sqrt_price_a

    amount_big_dec = Decimal(amount)

    product = sqrt_price_a * sqrt_price_b
    diff = sqrt_price_b - sqrt_price_a
    if diff == Decimal(0):
        raise Exception(f"liquidity0 diff is zero: sqrtPriceA {sqrt_price_a} sqrtPriceB {sqrt_price_b}")

    result = (amount_big_dec * product) / diff

    return round_sdk_prec_down(result)

def liquidity1(amount: int, sqrt_price_a: Decimal, sqrt_price_b: Decimal) -> Decimal:
    # Swap if sqrt_price_a is greater than sqrt_price_b
    if sqrt_price_a > sqrt_price_b:
        sqrt_price_a, sqrt_price_b = sqrt_price_b, sqrt_price_a

    amount_big_dec = Decimal(amount)

    diff = sqrt_price_b - sqrt_price_a
    if diff == Decimal(0):
        raise Exception(f"liquidity1 diff is zero: sqrtPriceA {sqrt_price_a} sqrtPriceB {sqrt_price_b}")
    
    result = amount_big_dec / diff

    return round_sdk_prec_down(result)

def get_liquidity_from_amounts(sqrt_price, sqrt_price_a, sqrt_price_b, amount0, amount1):
    # Reorder the prices so that sqrt_price_a is the smaller of the two
    if sqrt_price_a > sqrt_price_b:
        sqrt_price_a, sqrt_price_b = sqrt_price_b, sqrt_price_a

    if sqrt_price <= sqrt_price_a:
        # If the current price is less than or equal to the lower tick, then we use the liquidity0 formula
        liquidity = liquidity0(amount0, sqrt_price_a, sqrt_price_b)
    elif sqrt_price < sqrt_price_b:
        # If the current price is between the lower and upper ticks (exclusive of both the lower and upper ticks,
        # as both would trigger a division by zero), then we use the minimum of the liquidity0 and liquidity1 formulas
        liquidity_0 = liquidity0(amount0, sqrt_price, sqrt_price_b)
        liquidity_1 = liquidity1(amount1, sqrt_price, sqrt_price_a)
        liquidity = min(liquidity_0, liquidity_1)
    else:
        # If the current price is greater than the upper tick, then we use the liquidity1 formula
        liquidity = liquidity1(amount1, sqrt_price_b, sqrt_price_a)

    return liquidity

def get_next_sqrt_price_from_amount0_out_round_up(liquidity, sqrtPriceCurrent, tokenOut):
    product_num = liquidity * sqrtPriceCurrent
    product_num = round_osmo_prec_up(product_num)
    product_den =  tokenOut * sqrtPriceCurrent
    product_den = round_osmo_prec_up(product_den)
    return round_osmo_prec_up(product_num / (liquidity - product_den))

def get_next_sqrt_price_from_amount0_in_round_up(liquidity, sqrtPriceCurrent, tokenIn):
    return round_osmo_prec_up(round_osmo_prec_up(liquidity * sqrtPriceCurrent) / (liquidity + round_osmo_prec_down(tokenIn * sqrtPriceCurrent)))

def get_next_sqrt_price_from_amount1_out_round_down(liquidity, sqrtPriceCurrent, tokenOut):
    return round_osmo_prec_down(sqrtPriceCurrent - round_osmo_prec_up(tokenOut / liquidity))

def get_next_sqrt_price_from_amount1_in_round_down(liquidity, sqrtPriceCurrent, tokenIn):
    return sqrtPriceCurrent + round_osmo_prec_down(tokenIn / liquidity)

def calc_amount_zero_delta(liquidity, sqrtPriceA, sqrtPriceB, roundUp):
    if sqrtPriceB > sqrtPriceA:
        sqrtPriceA, sqrtPriceB = sqrtPriceB, sqrtPriceA
        
    diff = sqrtPriceA - sqrtPriceB

    product_num = liquidity * diff
    product_denom = sqrtPriceA * sqrtPriceB

    if roundUp:
        return round_osmo_prec_up(round_osmo_prec_up(product_num) / round_osmo_prec_down(product_denom))

    return round_osmo_prec_down(round_osmo_prec_down(product_num) / round_osmo_prec_up(product_denom))

def calc_amount_one_delta(liquidity, sqrtPriceA, sqrtPriceB, roundUp):
    if sqrtPriceB > sqrtPriceA:
        sqrtPriceA, sqrtPriceB = sqrtPriceB, sqrtPriceA

    diff = sqrtPriceA - sqrtPriceB

    if roundUp:
        return round_osmo_prec_up(liquidity * diff) 

    return round_osmo_prec_down(liquidity * diff) 
