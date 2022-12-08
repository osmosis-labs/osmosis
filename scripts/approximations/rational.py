import sympy as sp

import polynomial

def construct_rational_eval_matrix(x_list: list, y_list: list, num_terms_numerator: int, num_terms_denominator) -> sp.Matrix:
    """ Constructs a matrix to use for computing coefficients for a rational approximation
        from the list of x and y values given.
        len(x_list) * len(x_list)
        x_list is the list of all x values to construct the matrix from.
        V = [1 x_0 x_0^2 ... x_0^{n-1} - y_0*x_0 - y_0*x_0^2 ... - y_0*x_0^{n-1}]
            [1 x_1 x_2^1 ... x_1^{n-1} - y_1*x_1 - y_1*x_1^2 ... - y_1*x_1^{n-1}]
            ...
            [1 x_n x_n^2 ... x_n^{n-1} - y_n*x_n - y_n*x_n^2 ... - y_n*x_n^{n-1}]
    """
    matrix = []

    for i in range(num_terms_numerator + num_terms_denominator - 1):
        row = []
        for j in range(num_terms_numerator):
            row.append(sp.Pow(x_list[i], j))

        for j in range(num_terms_denominator):
            # denominator terms
            if j > 0:
                row.append(-1 * sp.Pow(x_list[i], j) * y_list[i])

        matrix.append(row)

    return sp.Matrix(matrix)

def evaluate(x: list, coefficients_numerator: list, coefficients_denominator: list):
    """ Evaluates the rational function. Assume rational h(x) = p(x) / q(x)
    Given a list of x coordinates, a list of coefficients of p(x) - coefficients_numerator, and a list of
    coefficients of q(x) - coefficients_denominator, returns a list of y coordinates, one for each x coordinate.
    """
    p = polynomial.evaluate(x, coefficients_numerator)
    q = polynomial.evaluate(x, coefficients_denominator)

    result = []

    for i in range(len(x)):
        result.append(p[i] / q[i])

    return result
