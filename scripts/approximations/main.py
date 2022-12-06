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
    # number of terms in the polynomial approximation / numerator of the rational approximation.
    num_terms_approximation_poly = 6

    num_terms_aprroximation_rational_numerator = num_terms_approximation_poly // 2
    num_terms_aprroximation_rational_denominator = num_terms_approximation_poly // 2
    if num_terms_approximation_poly % 2 == 1:
        # Assume num_terms_approximation_poly is 3.
        # Then, num_terms_aprroximation_rational_numerator = 1
        # and num_terms_aprroximation_rational_denominator = 2
        num_terms_aprroximation_rational_numerator += 1

    # number of (x,y) coordinates used to plot the resulting approximation.
    num_points_plot = 100000
    # number of (x,y) coordinates used to plot the
    # actual function that is evenly spaced on the X-axis.
    num_points_plot_accurate = 50000
    # function to approximate
    approximated_fn = lambda x: math.e**x

    # flag controlling whether to plot each approximation.
    # Plots if true.
    shouldPlotApproximations = False

    # flag controlling whether to compute max error for each approximation
    # given the equally spaced x coordinates.
    # Computes if true.
    shouldComputeErrorDelta = True

    #####################
    # 2. Approximations 

    # 2.1. Equispaced Polynomial Approximation
    coefficients_equispaced_poly = approximations.equispaced_poly_approx(approximated_fn, x_start, x_end, num_terms_approximation_poly)
    
    # 2.2. Chebyshev Polynomial Approximation
    coefficients_chebyshev_poly = approximations.chebyshev_poly_approx(approximated_fn, x_start, x_end, num_terms_approximation_poly)
    
    # 2.3. Chebyshev Rational Approximation
    numerator_coefficients_chebyshev_rational, denominator_coefficients_chebyshev_rational = approximations.chebyshev_rational_approx(approximated_fn, x_start, x_end,num_terms_aprroximation_rational_numerator, num_terms_aprroximation_rational_denominator)

    # 2.4. Actual With Large Number of Coordinates (evenly spaced on the X-axis)
    x_accurate = np.linspace(x_start, x_end, num_points_plot_accurate)

    #######################################
    # 3. Compute (x,y) Coordinates To Plot

    # Equispaced x coordinates to be used for plotting every approximation.
    plot_nodes_x = np.linspace(x_start, x_end, num_points_plot)

    # 3.1 Equispaced Polynomial Approximation
    plot_nodes_y_eqispaced_poly = polynomial.evaluate(plot_nodes_x, coefficients_equispaced_poly)

    # 3.2 Chebyshev Polynomial Approximation
    plot_nodes_y_chebyshev_poly = polynomial.evaluate(plot_nodes_x, coefficients_chebyshev_poly)

    # 3.3 Chebyshev Rational Approximation
    y_chebyshev_rational = rational.evaluate(plot_nodes_x, numerator_coefficients_chebyshev_rational.tolist(), denominator_coefficients_chebyshev_rational.tolist())

    # 3.4 Actual With Large Number of Coordinate (evenly spaced on the X-axis)
    y_accurate = approximated_fn(x_accurate)

    #############################
    # 4. Plot Every Approximation

    if shouldPlotApproximations:
        # 4.1 Equispaced Polynomial Approximation
        plt.plot(plot_nodes_x, plot_nodes_y_eqispaced_poly, label="Equispaced Poly")

        # 4.2 Chebyshev Polynomial Approximation
        plt.plot(plot_nodes_x,plot_nodes_y_chebyshev_poly, label="Chebyshev Poly")

        # 4.3 Chebyshev Rational Approximation
        plt.plot(plot_nodes_x,y_chebyshev_rational, label="Chebyshev Rational")

        # 4.4 Actual With Large Number of Coordinates (evenly spaced on the X-axis)
        plt.plot(x_accurate,y_accurate, label=F"Actual - {num_points_plot_accurate} evenly spaced points")

        plt.legend(loc="upper left")
        plt.grid(True)
        plt.title(f"Appproximation of e^x on [{x_start}, {x_end}] with {num_terms_approximation_poly} terms")
        plt.show()

    #############################
    # 5. Compute Errors

    if shouldComputeErrorDelta:
        print(F"\n\nMax Error on [{x_start}, {x_end}]")
        print(F"{num_points_plot} coordinates equally spaced on the X axis")
        print(F"{num_terms_approximation_poly} polynomial terms and ({num_terms_aprroximation_rational_numerator}, {num_terms_aprroximation_rational_denominator}) rational terms used for approximations.\n\n")

        plot_nodes_y_actual = approximated_fn(plot_nodes_x)

        # 5.1 Equispaced Polynomial Approximation
        delta_eqispaced_poly = np.abs(plot_nodes_y_eqispaced_poly - plot_nodes_y_actual)
        print(F"Equispaced Poly: {np.amax(delta_eqispaced_poly)}")

        # 5.2 Chebyshev Polynomial Approximation
        delta_chebyshev_poly = np.abs(plot_nodes_y_chebyshev_poly - plot_nodes_y_actual)
        print(F"Chebyshev Poly: {np.amax(delta_chebyshev_poly)}")

        # 5.3 Chebyshev Rational Approximation
        delta_chebyshev_rational = np.abs(y_chebyshev_rational - plot_nodes_y_actual)
        print(F"Chebyshev Rational: {np.amax(delta_chebyshev_rational)}")

if __name__ == "__main__":
    main()
