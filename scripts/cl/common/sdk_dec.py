from decimal import *

precision = 18
quantize_precision = (Decimal("10") ** -18)

# default_ctx = Context(rounding=ROUND_05UP, prec=precision, clamp=1, Emin=-(precision+1), Emax=(precision+1))
# setcontext(default_ctx)

def mul(x: Decimal, y: Decimal) -> Decimal:
    if x.is_nan():
        raise Exception("mul x is NaN")
    if y.is_nan():
        raise Exception("mul y is NaN")
    print(x)
    print(y)
    print(quantize_precision)
    return (x * y).quantize(exp=quantize_precision)

def quo_custom_round(x: Decimal, y: Decimal, custom_round: int):
    if x.is_nan():
        raise Exception("mul x is NaN")
    if y.is_nan():
        raise Exception("mul y is NaN")
    if y.is_zero():
        raise DivisionByZero("quo y is zero")
    return (x / y).quantize(exp=quantize_precision, context=Context(rounding=custom_round, prec=precision, clamp=1))

def quo(x: Decimal, y: Decimal):
    return quo_custom_round(x, y, ROUND_05UP)

def quo_up(x: Decimal, y: Decimal):
    return quo_custom_round(x, y, ROUND_UP)

def quo_trunc(x: Decimal, y: Decimal):
    return quo_custom_round(x, y, ROUND_DOWN)

def new(value: str) -> Decimal:
    """ Return an equivalent of a Cosmos SDK Decimal with fixed precision. 
    """
    return Decimal(value=value)


zero = new("0")

one = new("1")
