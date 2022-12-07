import math
import numpy as np
import matplotlib.pyplot as plt
from typing import Tuple

import polynomial
import rational
import approximations

def approximate_all_with_num_parameters(approximated_fn, x_start: int, x_end: int, num_parameters: int, num_points_plot: int, num_points_plot_accurate: int) -> tuple[np.array, np.array, np.array, np.array]:
     #####################
    # 2. Approximations 

    # 2.1. Equispaced Polynomial Approximation
    coefficients_equispaced_poly = approximations.equispaced_poly_approx(approximated_fn, x_start, x_end, num_parameters)
    
    # 2.2. Chebyshev Polynomial Approximation
    coefficients_chebyshev_poly = approximations.chebyshev_poly_approx(approximated_fn, x_start, x_end, num_parameters)
    
    # 2.3. Chebyshev Rational Approximation
    numerator_coefficients_chebyshev_rational, denominator_coefficients_chebyshev_rational = approximations.chebyshev_rational_approx(approximated_fn, x_start, x_end, num_parameters)

    #######################################
    # 3. Compute (x,y) Coordinates To Plot

    # Equispaced x coordinates to be used for plotting every approximation.
    x_coordinates = np.linspace(x_start, x_end, num_points_plot)

    # 3.1 Equispaced Polynomial Approximation
    y_eqispaced_poly = polynomial.evaluate(x_coordinates, coefficients_equispaced_poly)

    # 3.2 Chebyshev Polynomial Approximation
    y_chebyshev_poly = polynomial.evaluate(x_coordinates, coefficients_chebyshev_poly)

    # 3.3 Chebyshev Rational Approximation
    y_chebyshev_rational = rational.evaluate(x_coordinates, numerator_coefficients_chebyshev_rational.tolist(), denominator_coefficients_chebyshev_rational.tolist())

    # 3.4 Actual With Large Number of Coordinate (evenly spaced on the X-axis)
    y_actual = approximated_fn(x_coordinates)

    return (y_eqispaced_poly, y_chebyshev_poly, y_chebyshev_rational, y_actual)

    # # 5.1 Equispaced Polynomial Approximation
    # delta_eqispaced_poly = np.abs(plot_nodes_y_eqispaced_poly - plot_nodes_y_actual)

    # # 5.2 Chebyshev Polynomial Approximation
    # delta_chebyshev_poly = np.abs(plot_nodes_y_chebyshev_poly - plot_nodes_y_actual)

    # # 5.3 Chebyshev Rational Approximation
    # delta_chebyshev_rational = np.abs(y_chebyshev_rational - plot_nodes_y_actual)

    # return (np.amax(delta_eqispaced_poly), np.amax(delta_chebyshev_poly), np.amax(delta_chebyshev_rational))

def compute_max_error(y_approximation: np.array, y_actual: np.array) -> float:
    return np.amax(np.abs(y_approximation - y_actual))

# This script does the following:
# - Computes polynomial and rational approximations of a given function (e^x by default).
# - Computes (x,y) coordinates for every approximation given the same x coordinates.
# - Plots the results for rough comparison.
# The following are the resources used to write the script:
# https://xn--2-umb.com/22/approximation/
# https://sites.tufts.edu/atasissa/files/2019/09/remez.pdf
def main():

    ##############################
    # 1. Configuration Parameters

    # start of the interval to calculate the approximation on
    x_start = 0
    # end of the interval to calculate the approximation on
    x_end = 1

    # number of (x,y) coordinates used to plot the resulting approximation.
    num_points_plot = 100000
    # number of (x,y) coordinates used to plot the
    # actual function that is evenly spaced on the X-axis.
    num_points_plot_accurate = 50000
    # function to approximate
    approximated_fn = lambda x: math.e**x

    x_axis = []

    deltas_eqispaced_poly = []
    deltas_chebyshev_poly = []
    deltas_chebyshev_rational = []

    ###################
    # 2. Compute Deltas
    # The deltas are taken from actual function values for different number of parameters
    # This is needed to find the most optimal number of parameters to use.
    for num_parameters in range(1, 21):
        x_axis.append(int(num_parameters))
        y_eqispaced_poly, y_chebyshev_poly, y_chebyshev_rational, y_actual = approximate_all_with_num_parameters(approximated_fn, x_start, x_end, num_parameters, num_points_plot, num_points_plot_accurate)

        deltas_eqispaced_poly.append(compute_max_error(y_eqispaced_poly, y_actual))
        deltas_chebyshev_poly.append(compute_max_error(y_chebyshev_poly, y_actual))
        deltas_chebyshev_rational.append(compute_max_error(y_chebyshev_rational, y_actual))

    #####################
    # 3. Plot the results

    # 3.1 Equispaced Polynomial Approximation
    plt.semilogy(x_axis, deltas_eqispaced_poly, label="Equispaced Poly")

    # 3.2 Chebyshev Polynomial Approximation
    plt.semilogy(x_axis, deltas_chebyshev_poly, label="Chebyshev Poly")

    # 3.3 Chebyshev Rational Approximation
    plt.semilogy(x_axis, deltas_chebyshev_rational, label="Chebyshev Rational")

    plt.legend(loc="upper left")
    plt.grid(True)
    plt.title(f"Approximation Errors on [{x_start}, {x_end}]")
    plt.gca().invert_yaxis()
    plt.xlabel('Number of Parameters')
    plt.ylabel(F"-log_10{{ max | f'(x) - f(x) | }} where x is in [{x_start}, {x_end}]")
    plt.show()

if __name__ == "__main__":
    main()
