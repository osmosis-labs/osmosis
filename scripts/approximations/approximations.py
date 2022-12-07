from typing import Tuple
from sympy import *
import sympy

import rational
import polynomial
import chebyshev

def linspace(start: Float, end: Float, num_points: int) -> list[Float]:
    if num_points == 1:
        return [(end - start) / 2]

    result: list[Float] = []
    for i in range(num_points):
        cur_coord = Float(start + i * (end - start) / (num_points - 1), 40)

        if cur_coord is nan:
            raise ValueError("cur_coord is nan")

        result.append(cur_coord)

    return result

def approx_and_eval_all(approximated_fn, num_parameters: int, x_coordinates) -> tuple:
    x_start = x_coordinates[0]
    x_end = x_coordinates[-1]
    
    # Equispaced Polynomial Approximation
    coefficients_equispaced_poly = equispaced_poly_approx(approximated_fn, x_start, x_end, num_parameters)
    y_eqispaced_poly = polynomial.evaluate(x_coordinates, coefficients_equispaced_poly)

    if True:
        y_actual = []
        for x in x_coordinates:
            y_actual.append(approximated_fn(x))
        return (y_eqispaced_poly, [], [], y_actual)

    # Chebyshev Polynomial Approximation
    coefficients_chebyshev_poly = chebyshev_poly_approx(approximated_fn, x_start, x_end, num_parameters)
    y_chebyshev_poly = polynomial.evaluate(x_coordinates, coefficients_chebyshev_poly)
    
    # Chebyshev Rational Approximation
    numerator_coefficients_chebyshev_rational, denominator_coefficients_chebyshev_rational = chebyshev_rational_approx(approximated_fn, x_start, x_end, num_parameters)
    y_chebyshev_rational = rational.evaluate(x_coordinates, numerator_coefficients_chebyshev_rational.tolist(), denominator_coefficients_chebyshev_rational.tolist())

    # Actual
    y_actual = approximated_fn(x_coordinates)

    return (y_eqispaced_poly, y_chebyshev_poly, y_chebyshev_rational, y_actual)

def compute_max_error(y_approximation, y_actual) -> Float:
    if len(y_approximation) != len(y_actual):
        raise ValueError(F"y_approximation ({len(y_approximation)}) and y_actual ({len(y_actual)}) must be the same length.")

    print(f"\ny_approximation: ", y_approximation)
    print(f"\ny_actual: ", y_actual)

    max: Float = None
    for i in range(len(y_approximation)):
        cur_abs = sympy.Abs(y_approximation[i] - y_actual[i])
        if cur_abs is nan:
            raise ValueError(F"cur_abs is nan. y_approximation[i] ({y_approximation[i]}) and y_actual[i] ({y_actual[i]})")
        if max is None:
            max = cur_abs
        else:
            max = sympy.Max(max, cur_abs)
    return max

def equispaced_poly_approx(fn, x_start: Float, x_end: Float, num_terms: int):
    """ Returns the coefficients for an equispaced polynomial between x_start and x_end with num_terms terms.

    The return value is a list of num_terms polynomial coefficients needed to get the returned y coordinates from returned x coordinates.
    """
    # Compute equispaced coordinates.
    equispaced_nodes_x = linspace(x_start, x_end, num_terms)
    y_nodes = sympy.Matrix([fn(x) for x in equispaced_nodes_x])

    # Construct a system of linear equations.
    vandermonde_matrix = polynomial.construct_vandermonde_matrix(equispaced_nodes_x)

    coef = vandermonde_matrix.solve(y_nodes)

    return coef

def chebyshev_poly_approx(fn, x_start: int, x_end: int, num_terms: int):
    """ Returns the coefficients for a polynomial constructed from Chebyshev nodes between x_start and x_end with num_terms terms.

    Equation for Chebyshev nodes:
    x_i = (x_start + x_end)/2 + (x_end - x_start)/2 * cos((2*i + 1)/(2*num_terms) * pi)

    The return value is a list of num_terms polynomial coefficients needed to get the returned y coordinates from returned x coordinates.
    """
    # Compute Chebyshev coordinates.
    x_estimated, y_estimated = chebyshev.get_nodes(fn, x_start, x_end, num_terms)

    # Construct a system of linear equations.
    vandermonde_matrix = polynomial.construct_vandermonde_matrix(x_estimated)

    # Solve the matrix to get the coefficients used in the final approximation polynomial.
    # coef = np.linalg.solve(np.array(vandermonde_matrix), y_estimated)

    return nan

def chebyshev_rational_approx(fn, x_start: int, x_end: int, num_parameters: int):
    """ Returns a rational approximation between x_start and x_end with num_terms terms
    using Chebyshev nodes.

    Equation for Chebyshev nodes:
    x_i = (x_start + x_end)/2 + (x_end - x_start)/2 * cos((2*i + 1)/(2*num_terms) * pi)

    We approximate h(x) = p(x) / q(x)

    Assume num_terms_numerator = 3.
    Then, we have

    h(x) = (p_0 + p_1 x + p_2 x^2) / (1 + q_1 x + q_2 x^2) 

    The return value is a list with a 2 items where:
    - item 1: num_terms equispaced x coordinates between x_start and x_end
    - item 2: num_terms y coordinates for the equispaced x coordinates
    """
    if num_parameters % 2 == 0:
        # if num_parameters is 6, we want (3, 4) terms
        # assume h(x) = p(x) / q(x)
        # we always set the first term of q(x) to 1 for ease of calculation.
        # so, we want p(x) to have 3 terms and q(x) to have 4 terms.
        num_terms_numerator = num_parameters // 2
        num_terms_denominator = num_parameters // 2 + 1
    else:
        # if num_parameters is 5, we want (3, 3) terms
        # assume h(x) = p(x) / q(x)
        # we always set the first term of q(x) to 1 for ease of calculation.
        # so, we want p(x) to have 3 terms and q(x) to have 4 terms.
        num_terms_numerator = num_parameters // 2 + 1
        num_terms_denominator = num_parameters // 2 + 1

    # Compute Chebyshev coordinates.
    x_chebyshev, y_chebyshev = chebyshev.get_nodes(fn, x_start, x_end, num_parameters)

    # Construct a system of linear equations.
    matrix = rational.construct_rational_eval_matrix(x_chebyshev, y_chebyshev, num_terms_numerator, num_terms_denominator)

    # Solve the matrix to get the coefficients used in the final approximation polynomial.
    # coef = np.linalg.solve(np.array(matrix), y_chebyshev)

    # first num_terms_numerator values are the numerator coefficients
    # next num_terms_numerator - 1 values are the denominator coefficients
    # coef_numerator = coef[:num_terms_numerator]
    # coef_denominator = coef[num_terms_numerator:]

    # h(x) = (p_0 + p_1 x + p_2 x^2) / (1 + q_1 x + q_2 x^2)
    # Therefore, we insert 1 here.
    # coef_denominator = np.insert(coef_denominator, 0, 1, axis=0)

    # return [coef_numerator, coef_denominator]
    return [nan, nan]
