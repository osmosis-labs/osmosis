import math
import numpy as np
import matplotlib.pyplot as plt

import polynomial
import rational
import approximations

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
    # number of paramters to use for the approximations.
    num_parameters = 12

    # number of (x,y) coordinates used to plot the resulting approximation.
    num_points_plot = 10000

    # function to approximate
    approximated_fn = lambda x: math.e**x

    # flag controlling whether to plot each approximation.
    # Plots if true.
    shouldPlotApproximations = True

    # flag controlling whether to compute max error for each approximation
    # given the equally spaced x coordinates.
    # Computes if true.
    shouldComputeErrorDelta = True

    #####################
    # 2. Approximations 

    y_eqispaced_poly, y_chebyshev_poly, y_chebyshev_rational, y_actual = approximations.approximate_all_with_num_parameters(approximated_fn, x_start, x_end, num_parameters, num_points_plot)

    # Equispaced x coordinates to be used for plotting every approximation.
    plot_nodes_x = np.linspace(x_start, x_end, num_points_plot)


    #############################
    # 4. Compute Errors

    if shouldComputeErrorDelta:
        print(F"\n\nMax Error on [{x_start}, {x_end}]")
        print(F"{num_points_plot} coordinates equally spaced on the X axis")
        print(F"{num_parameters} parameters used\n\n")

        plot_nodes_y_actual = approximated_fn(plot_nodes_x)

        # 4.1 Equispaced Polynomial Approximation
        delta_eqispaced_poly = np.abs(y_eqispaced_poly - plot_nodes_y_actual)
        print(F"Equispaced Poly: {np.amax(delta_eqispaced_poly)}")

        # 4.2 Chebyshev Polynomial Approximation
        delta_chebyshev_poly = np.abs(y_chebyshev_poly - plot_nodes_y_actual)
        print(F"Chebyshev Poly: {np.amax(delta_chebyshev_poly)}")

        # 4.3 Chebyshev Rational Approximation
        delta_chebyshev_rational = np.abs(y_chebyshev_rational - plot_nodes_y_actual)
        print(F"Chebyshev Rational: {np.amax(delta_chebyshev_rational)}")

    #############################
    # 5. Plot Every Approximation

    if shouldPlotApproximations:
        # 5.1 Equispaced Polynomial Approximation
        plt.plot(plot_nodes_x, y_eqispaced_poly, label="Equispaced Poly")

        # 5.2 Chebyshev Polynomial Approximation
        plt.plot(plot_nodes_x, y_chebyshev_poly, label="Chebyshev Poly")

        # 5.3 Chebyshev Rational Approximation
        plt.plot(plot_nodes_x, y_chebyshev_rational, label="Chebyshev Rational")

        # 5.4 Actual With Large Number of Coordinates (evenly spaced on the X-axis)
        plt.plot(plot_nodes_x, y_actual, label=F"Actual")

        plt.legend(loc="upper left")
        plt.grid(True)
        plt.title(f"Appproximation of e^x on [{x_start}, {x_end}] with {num_parameters} parameters")
        plt.show()

if __name__ == "__main__":
    main()
