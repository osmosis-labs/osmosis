import unittest
import math

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

        coeffs = rational.construct_vandermonde_matrix(x, y)
        
        # number of rows is correct
        self.assertEqual(len(x), len(coeffs))

        # first row is correct
        for i in range(len(coeffs)):
            # number of columns in each row is correct
            self.assertEqual(len(coeffs), len(coeffs[i]))

            x_i = x[i]
            y_i = y[i]

            expected_row = [1, x_i, x_i**2, -1 * x_i * y_i, -1 * x_i**2 * y_i]

            actual_rot = coeffs[i]

            self.assertEqual(expected_row, actual_rot)

if __name__ == '__main__':
    unittest.main()
