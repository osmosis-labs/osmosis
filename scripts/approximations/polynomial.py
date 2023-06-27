import sympy as sp

def construct_vandermonde_matrix(x_list: list[sp.Float]) -> sp.Matrix:
    """ Constructs a Vandermonde matrix for a polynomial approximation.
    from the list of x values given.
    
    len(x_list) * len(x_list)
    x_list is the list of all x values to construct the matrix from.
    
    V = [1 x_0 x_0^2 ... x_0^{n-1}]
    [1 x_1 x_2^1 ... x_1^{n-1}]
    ...
    [1 x_n x_n^2 ... x_n^{n-1}]

    Vandermonde matrix is a matrix with the terms of a geometric progression in each row.
    We use Vandermonde matrix to convert the system of linear equations into a linear algebra problem
    that we can solve.
    """
    num_terms = len(x_list)

    matrix = []

    for i in range(num_terms):
        row = []
        for j in range(num_terms):
            row.append(sp.Pow(x_list[i], j))
        matrix.append(row)

    return sp.Matrix(matrix)

def evaluate(x, coeffs):
    """ Evaluates the polynomial. Given a list of x coordinates and a list of coefficients, returns a list of
    y coordinates, one for each x coordinate. The coefficients must be in ascending order.
    """
    y = []
    for x_i in x:
        y_i = coeffs[0]
        x_pow = x_i
        for i in range(1, len(coeffs)):
            y_i += coeffs[i] * x_pow
            x_pow *= x_i
        y.append(y_i)
    return y
