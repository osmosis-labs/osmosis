import sympy as sp
from typing import Tuple

# N.B. We must evaluate x values early.
# Otherwise, solving for the coefficients will take a long time.
_x_eval_precision = 200
_x_eval_chop = 1e-200

def get_nodes(fn, x_start: sp.Float, x_end: sp.Float, num_terms: int) -> Tuple[list, list]:
    """ Returns Chebyshev nodes between x_start and x_end with num_terms terms
    and the given function fn.

    Equation for Chebyshev nodes:
    x_i = (x_start + x_end)/2 + (x_end - x_start)/2 * cos((2*i + 1)/(2*num_terms) * pi)
    
    The first returned value is a list of x coordinates for the Chebyshev nodes.
    The second returned value is a list of y coordinates for the Chebyshev nodes.
    """
    x_estimated = []
    y_estimated = []

    for i in range (num_terms):
        x_i = ((x_start + x_end) / 2 + (x_end - x_start) / 2 * sp.cos((2*sp.Float(i,_x_eval_precision) + 1).evalf(chop=_x_eval_chop) * sp.pi.evalf(chop=_x_eval_chop) / (2 * sp.Float(num_terms, _x_eval_precision))))
        y_i = fn(x_i)

        x_estimated.append(x_i)
        y_estimated.append(y_i)

    return x_estimated, y_estimated
