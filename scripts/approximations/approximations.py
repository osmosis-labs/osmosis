import sympy as sp

import rational
import polynomial
import chebyshev

def linspace(start: sp.Float, end: sp.Float, num_points: int) -> list[sp.Float]:
    """ Given [start, end] on the x-axis, creates the desired number of
    equispaced points and returns them in increasing order as a list. 
    """
    if num_points == 1:
        return [(end - start) / 2]

    result: list[sp.Float] = []
    for i in range(num_points):
        cur_coord = sp.Float(start + i * (end - start) / (num_points - 1), 200)

        if cur_coord is sp.nan:
            raise ValueError("cur_coord is nan")

        result.append(cur_coord)

    return result

def approx_and_eval_all(approximated_fn, num_parameters: int, x_coordinates) -> tuple:
    """ Performs all supported approximations of the given function, evaluates
    each wih the given x-coordinates.

    Returns y-coordinates as a tuple in the following order:
    - Evaluated Equispaced Polynomial
    - Evaluated Chebyshev Polynomial
    - Evaluated Chebyshev Rational
    - True Y-coordinates as determined by Sympy's implementation of the function.
    """
    x_start = x_coordinates[0]
    x_end = x_coordinates[-1]
    
    # Equispaced Polynomial Approximation
    coefficients_equispaced_poly = equispaced_poly_approx(approximated_fn, x_start, x_end, num_parameters)
    y_eqispaced_poly = polynomial.evaluate(x_coordinates, coefficients_equispaced_poly)

    # Chebyshev Polynomial Approximation
    coefficients_chebyshev_poly = chebyshev_poly_approx(approximated_fn, x_start, x_end, num_parameters)
    y_chebyshev_poly = polynomial.evaluate(x_coordinates, coefficients_chebyshev_poly)

    # Chebyshev Rational Approximation
    numerator_coefficients_chebyshev_rational, denominator_coefficients_chebyshev_rational = chebyshev_rational_approx(approximated_fn, x_start, x_end, num_parameters)
    y_chebyshev_rational = rational.evaluate(x_coordinates, numerator_coefficients_chebyshev_rational, denominator_coefficients_chebyshev_rational)

    # Actual
    y_actual = get_y_actual(approximated_fn, x_coordinates)

    return (y_eqispaced_poly, y_chebyshev_poly, y_chebyshev_rational, y_actual)

def get_y_actual(approximated_fn, x_coordinates) -> list[sp.Float]:
    """ Given x coordinates and a function, returns a list of y coordinates, one for each x coordinate
    that represent the true (x,y) coordinates of the function, as calculated by sp.
    """
    y_actual = []
    for x in x_coordinates:
        y_actual.append(approximated_fn(x))
    return y_actual

def compute_max_error(y_approximation, y_actual) -> sp.Float:
    """ Given an approximated list of y values and actual y values, computes and returns
    the maximum error delta between them.

    CONTRACT:
    - y_approximation and y_actual must be the same length
    - for every i in range(len(y_approximation)), y_approximation[i] and y_actual[i] must correspond to the
    same x coordinate
    """
    if len(y_approximation) != len(y_actual):
        raise ValueError(F"y_approximation ({len(y_approximation)}) and y_actual ({len(y_actual)}) must be the same length.")

    max: sp.Float = None
    for i in range(len(y_approximation)):
        cur_abs = sp.Abs(y_approximation[i] - y_actual[i])
        if cur_abs is sp.nan:
            raise ValueError(F"cur_abs is nan. y_approximation[i] ({y_approximation[i]}) and y_actual[i] ({y_actual[i]})")
        if max is None:
            max = cur_abs
        else:
            max = sp.Max(max, cur_abs)
    return max

def compute_absolute_error_range(y_approximation: list, y_actual: list) -> list[sp.Float]:
    """ Given an approximated list of y values and actual y values, computes and returns
    absolute error between them, computed as | y_approximation[i] - y_actual[i] |.

    CONTRACT:
    - y_approximation and y_actual must be the same length
    - for every i in range(len(y_approximation)), y_approximation[i] and y_actual[i] must correspond to the
    same x coordinate
    """
    result = []
    for i in range(len(y_approximation)):
        cur_abs = sp.Abs(y_approximation[i] - y_actual[i])
        if cur_abs is sp.nan:
            raise ValueError(F"cur_abs is nan. y_approximation[i] ({y_approximation[i]}) and y_actual[i] ({y_actual[i]})")
        result.append(cur_abs)
    return result

def compute_relative_error_range(y_approximation: list, y_actual: list) -> list[sp.Float]:
    """ Given an approximated list of y values and actual y values, computes and returns
    relative error between them, computed as | y_approximation[i] - y_actual[i] | / y_actual[i].

    For y_actual[i] = 0, relative error is defined as 0.

    CONTRACT:
    - y_approximation and y_actual must be the same length
    - for every i in range(len(y_approximation)), y_approximation[i] and y_actual[i] must correspond to the
    same x coordinate
    """
    result = []
    for i in range(len(y_approximation)):
        if y_actual[i] == 0:
            result.append(0)
            continue

        cur_relative_error = sp.Abs(y_approximation[i] - y_actual[i]) / y_actual[i]
        if cur_relative_error is sp.nan:
            raise ValueError(F"cur_abs is nan. y_approximation[i] ({y_approximation[i]}) and y_actual[i] ({y_actual[i]})")
        result.append(cur_relative_error)
    return result

def equispaced_poly_approx(fn, x_start: sp.Float, x_end: sp.Float, num_terms: int):
    """ Returns the coefficients for an equispaced polynomial between x_start and x_end with num_terms terms.

    The return value is a list of num_terms polynomial coefficients needed to get the returned y coordinates from returned x coordinates.
    """
    # Compute equispaced coordinates.
    equispaced_nodes_x = linspace(x_start, x_end, num_terms)
    y_nodes = sp.Matrix([fn(x) for x in equispaced_nodes_x])

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
    coef = vandermonde_matrix.solve(sp.Matrix(y_estimated))

    return coef

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
    coef = matrix.solve(sp.Matrix(y_chebyshev))

    # first num_terms_numerator values are the numerator coefficients
    # next num_terms_numerator - 1 values are the denominator coefficients
    coef_numerator = coef[:num_terms_numerator]
    coef_denominator = coef[num_terms_numerator:]

    # h(x) = (p_0 + p_1 x + p_2 x^2) / (1 + q_1 x + q_2 x^2)
    # Therefore, we insert 1 here.
    coef_denominator = [1] + coef_denominator

    return [coef_numerator, coef_denominator]
