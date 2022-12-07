
import math
import sympy
from typing import Tuple

def get_nodes(fn, x_start: sympy.Float, x_end: sympy.Float, num_terms: int ) -> Tuple[list, list]:
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
        # TODO: evals below can be further reduced to improve precision.
        x_i = ((x_start + x_end) / 2 + (x_end - x_start) / 2 * sympy.cos((2*sympy.Float(i,100) + 1).evalf(chop=1e-100) * sympy.pi.evalf(chop=1e-100) / (2 * sympy.Float(num_terms, 100))))
        y_i = fn(x_i)

        x_estimated.append(x_i)
        y_estimated.append(y_i)

    return x_estimated, y_estimated
