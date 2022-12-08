import unittest
import math
import sympy as sp

import rational

class TestChebyshevRational(unittest.TestCase):

    # assume (3,3) rational 
    # h(x) = p(x) / q(x) = (p_0 + p_1 * x + p_2 * x^2) / (1 + q_1 * x + q_2 * x^2)
    #  p_0 + p_1 * x + p_2 * x^2 - q_1 * y * x - q_2 * y * x^2 = y
    # construct 5 x 5 matrix solving for p and q given x_0 to x_4 and y_0 to y_4
    # assume function is y = e**x
    def test_construct_matrix_3_3(self):
        fn = lambda x: math.e**x
        x = [1, 2, 3, 4, 5]
        y = list(map(fn, x))

        coeffs = rational.construct_rational_eval_matrix(x, y, 3, 3)
        
        # correct matrix size
        self.assertEqual(len(x) * len(x), len(coeffs))

        # first row is correct
        for i in range(len(x)):
            x_i = x[i]
            y_i = y[i]

            expected_row = [sp.Pow(x_i, 0), sp.Pow(x_i, 1), sp.Pow(x_i, 2), -1 * sp.Pow(x_i, 1) * y_i, -1 * sp.Pow(x_i, 2) * y_i]

            actual_row = coeffs.row(i).tolist()[0]

            self.assertEqual(expected_row, actual_row)

if __name__ == '__main__':
    unittest.main()
