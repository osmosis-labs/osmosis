import numpy as np

def construct_vandermonde_matrix(x_list: list, y_list: list) -> list[list]:
    """ Constructs a Vandermonde matrix for a rational approximation.
        from the list of x and y values given.
        len(x_list) * len(x_list)
        x_list is the list of all x values to construct the matrix from.
        V = [1 x_0 x_0^2 ... x_0^{n-1} - y_0*x_0 - y_0*x_0^2 ... - y_0*x_0^{n-1}]
            [1 x_1 x_2^1 ... x_1^{n-1} - y_1*x_1 - y_1*x_1^2 ... - y_1*x_1^{n-1}]
            ...
            [1 x_n x_n^2 ... x_n^{n-1} - y_n*x_n - y_n*x_n^2 ... - y_n*x_n^{n-1}]
    """
    num_terms = (len(x_list) + 1) // 2

    matrix = []

    for i in range(num_terms * 2 - 1):
        row = []
        for j in range(num_terms):
            row.append(x_list[i]**j)

        for j in range(num_terms):
            # denominator terms
            if j > 0:
                row.append(-1 * x_list[i]**j * y_list[i])

        matrix.append(row)

    return matrix

def solve(x: np.ndarray, coefficients_numerator: list, coefficients_denominator: list) -> np.ndarray:
    """ Solves the rational function. Assume rational h(x) = p(x) / q(x)
    Given a list of x coordinates, a list of coefficients of p(x) - coefficients_numerator, and a list of
    coefficients of q(x) - coefficients_denominator, returns a list of y coordinates, one for each x coordinate.
    """
    num_numerator_coeffs = len(coefficients_numerator)
    num_denominator_coeffs = len(coefficients_denominator)

    p = 0
    for i in range(num_numerator_coeffs):
        p += coefficients_numerator[i]*x**i
    
    q = 0
    for i in range(num_denominator_coeffs):
        q += coefficients_denominator[i]*x**i

    return p / q
