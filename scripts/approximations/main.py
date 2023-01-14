import matplotlib.pyplot as plt
import sympy as sp

import approximations
import rational

##########################
# Configuration Parameters

# start of the interval to calculate the approximation on
x_start = 0

# end of the interval to calculate the approximation on
x_end = 1

# number of paramters to use for the approximations.
num_parameters = 13

# number of paramters to use for plotting error deltas.
num_parameters_errors = 30

# number of (x,y) coordinates used to plot the resulting approximation.
num_points_plot = 10000

# function to approximate
approximated_fn = lambda x: sp.Pow(2, x)

# fixed point precision used in Osmosis `osmomath` package.
osmomath_precision = 36

# flag controlling whether to plot each approximation.
# Plots if true.
shouldPlotApproximations = True

# flag controlling whether to compute max error for each approximation
# given the equally spaced x coordinates.
# Computes if true.
shouldComputeErrorDelta = True

# flag controlling whether to plot errors over a range.
# Currently, does so only for Chebyshev Rational Approximation.
# Computes if true.
shouldPlotErrorRange  = True

# flag controlling whether to plot max error for every approximation
# with a varying number of parameters. This is useful to find the most
# optimal number of parameters to use for each kind of approximation.
# Plots if true.
shouldPlotMaxError = True

def plot_error_range(x_coordinates, y_approximation, y_actual, is_absolute: bool):
    """ Given x coordinates that correspond to approximated y coordinates and actual y coordinates,
    computes the error between y approximated and y actual and plot them in log scale on y.

    If is_absolute, plots the absolute error. Otherwise, plots the relative error.
    """
    if is_absolute:
        error_kind_str = "Absolute"
        error_deltas = approximations.compute_absolute_error_range(y_approximation, y_actual)
    else:
        error_kind_str = "Relative"
        error_deltas = approximations.compute_relative_error_range(y_approximation, y_actual)

    plt.semilogy(x_coordinates, error_deltas)

    plt.grid(True)
    plt.title(f"Chebyshev Rational e^x {error_kind_str} Errors on [{x_start}, {x_end}]. {num_parameters} params, {num_points_plot} points")
    plt.show()

# This script does the following:
# - Computes polynomial and rational approximations of a given function (e^x by default).
# - Computes (x,y) coordinates for every approximation given the same x coordinates.
# - Plots the results for rough comparison.
# - Computes the max error for every approximation given the same x coordinates.
# - Computes and plots max errors for every approximation with a varying number of parameters.
# This script runs various approximation methods, plots their results and deltas
# from actual function values. The script can also be configured to print the maximum error.
# The exact behavior is controlled by the global
# variables at the top of the file. 
# The following are the resources used:
# https://xn--2-umb.com/22/approximation/
# https://sites.tufts.edu/atasissa/files/2019/09/remez.pdf
def main():
    # Equispaced x coordinates to be used for plotting every approximation.
    x_coordinates = approximations.linspace(x_start, x_end, num_points_plot)

    if shouldComputeErrorDelta or shouldPlotApproximations or shouldPlotErrorRange:
        ###############################################
        # Approximation With Given Number of Parameters
        y_eqispaced_poly, y_chebyshev_poly, y_chebyshev_rational, y_actual = approximations.approx_and_eval_all(approximated_fn, num_parameters, x_coordinates)

        ################
        # Compute Errors
        if shouldComputeErrorDelta:
            print(F"\n\nMax Error on [{x_start}, {x_end}]")
            print(F"{num_points_plot} coordinates equally spaced on the X axis")
            print(F"{num_parameters} parameters used\n\n")

            # Equispaced Polynomial Approximation
            max_error_equispaced_poly = approximations.compute_max_error(y_eqispaced_poly, y_actual)
            print(F"Equispaced Poly: {max_error_equispaced_poly.evalf(chop=1e-100)}")

            # Chebyshev Polynomial Approximation
            max_error_chebyshev_poly = approximations.compute_max_error(y_chebyshev_poly, y_actual)
            print(F"Chebyshev Poly: {max_error_chebyshev_poly.evalf(chop=1e-100)}")

            # Chebyshev Rational Approximation
            max_error_chebyshev_rational = approximations.compute_max_error(y_chebyshev_rational, y_actual)
            print(F"Chebyshev Rational: {max_error_chebyshev_rational.evalf(chop=1e-100)}")

        if shouldPlotErrorRange:
            plot_error_range(x_coordinates, y_chebyshev_rational, y_actual)

        ###############################
        # Plot Every Approximation Kind
        if shouldPlotApproximations:
            # Equispaced Polynomial Approximation
            plt.plot(x_coordinates, y_eqispaced_poly, label="Equispaced Poly")

            # Chebyshev Polynomial Approximation
            plt.plot(x_coordinates, y_chebyshev_poly, label="Chebyshev Poly")

            # Chebyshev Rational Approximation
            plt.plot(x_coordinates, y_chebyshev_rational, label="Chebyshev Rational")

            # Actual With Large Number of Coordinates (evenly spaced on the X-axis)
            plt.plot(x_coordinates, y_actual, label=F"Actual")

            plt.legend(loc="upper left")
            plt.grid(True)
            plt.title(f"Appproximation of e^x on [{x_start}, {x_end}] with {num_parameters} parameters")
            plt.show()

    #####################################################
    # Calculate Errors Given Varying Number of Parameters
    if shouldPlotMaxError:
        x_axis = []

        deltas_eqispaced_poly = []
        deltas_chebyshev_poly = []
        deltas_chebyshev_rational = []

        ################
        # Compute Deltas
        # The deltas are taken from actual function values for different number of parameters
        # This is needed to find the most optimal number of parameters to use.
        for num_parameters_current in range(1, num_parameters_errors + 1):
            x_axis.append(int(num_parameters_current))
            y_eqispaced_poly, y_chebyshev_poly, y_chebyshev_rational, y_actual = approximations.approx_and_eval_all(approximated_fn, num_parameters_current, x_coordinates)

            deltas_eqispaced_poly.append(approximations.compute_max_error(y_eqispaced_poly, y_actual))
            deltas_chebyshev_poly.append(approximations.compute_max_error(y_chebyshev_poly, y_actual))
            deltas_chebyshev_rational.append(approximations.compute_max_error(y_chebyshev_rational, y_actual))

        ##################
        # Plot the results

        # Equispaced Polynomial Approximation
        plt.semilogy(x_axis, deltas_eqispaced_poly, label="Equispaced Poly")

        # Chebyshev Polynomial Approximation
        plt.semilogy(x_axis, deltas_chebyshev_poly, label="Chebyshev Poly")

        # Chebyshev Rational Approximation
        plt.semilogy(x_axis, deltas_chebyshev_rational, label="Chebyshev Rational")

        plt.legend(loc="upper left")
        plt.grid(True)
        plt.title(f"Approximation Errors on [{x_start}, {x_end}]")
        plt.gca().invert_yaxis()
        plt.xlabel('Number of Parameters')
        plt.ylabel(F"-log_10{{ max | f'(x) - f(x) | }} where x is in [{x_start}, {x_end}]")
        plt.show()

# This script isolates the 13-parameter Chebyshev Rational approximation of e^x
# We are planning to use it in production. Therefore, we need to peform coefficient
# truncations to 36 decimal points (the max osmomath supported precision).
def exponent_approximation_choice():
    # Equispaced x coordinates to be used for plotting every approximation.
    x_coordinates = approximations.linspace(x_start, x_end, num_points_plot)
    x_coordinates = [sp.N(sp.Float(coef, osmomath_precision), n=osmomath_precision) for coef in x_coordinates]

    print(x_coordinates)

    # Chebyshev Rational Approximation to get the coefficients.
    # coef_numerator, coef_denominator = approximations.chebyshev_rational_approx(approximated_fn, x_start, x_end, num_parameters)
    # coef_numerator = [sp.N(coef, osmomath_precision + 2) for coef in coef_numerator]
    # coef_denominator = [sp.N(coef, osmomath_precision + 2) for coef in coef_denominator]

    # Hard code and round up numerator coefficientst that are to be used in production
    # Hard code and round down numerator coefficientst that are to be used in production
    # Both of these are calculated us=ing the above commented out code.

    coef_numerator = [
        sp.N(sp.Float("1.000000000000000000000044212244679434", osmomath_precision), n=osmomath_precision),
        sp.N(sp.Float("0.352032455817400196452603772766844426", osmomath_precision), n=osmomath_precision),
        sp.N(sp.Float("0.056507868883666405413116800969512484", osmomath_precision), n=osmomath_precision),
        sp.N(sp.Float("0.005343900728213034434757419480319916", osmomath_precision), n=osmomath_precision),
        sp.N(sp.Float("0.000317708814342353603087543715930732", osmomath_precision), n=osmomath_precision),
        sp.N(sp.Float("0.000011429747507407623028722262874632", osmomath_precision), n=osmomath_precision),
        sp.N(sp.Float("0.000000198381965651614980168744540366", osmomath_precision), n=osmomath_precision),
    ]

    coef_denominator = [
        sp.N(sp.Float("1.0000000000000000000000000000000000000", osmomath_precision), n=osmomath_precision),
        sp.N(sp.Float("-0.341114724742545112949699755780593311", osmomath_precision), n=osmomath_precision),
        sp.N(sp.Float("0.052724071627342653404436933178482287", osmomath_precision), n=osmomath_precision),
        sp.N(sp.Float("-0.004760950735524957576233524801866342", osmomath_precision), n=osmomath_precision),
        sp.N(sp.Float("0.000267168475410566529819971616894193", osmomath_precision), n=osmomath_precision),
        sp.N(sp.Float("-0.000008923715368802211181557353097439", osmomath_precision), n=osmomath_precision),
        sp.N(sp.Float("0.000000140277233177373698516010555916", osmomath_precision), n=osmomath_precision),
    ]

    print(coef_numerator)
    print(coef_denominator)

    # Evaluate approximation.
    y_chebyshev_rational = rational.evaluate(x_coordinates, coef_numerator, coef_denominator)

    # Compute Actual Values
    y_actual = approximations.get_y_actual(approximated_fn, x_coordinates)

    plot_error_range(x_coordinates, y_chebyshev_rational, y_actual, True)

if __name__ == "__main__":
    # Uncomment to run the main script.
    #main()
    exponent_approximation_choice()
