# Math Operations Approximations

## Context

This is a set of scripts to approximate a mathematical operation using polynomial
and rational approximations.

The `main` script is in its respective function of `main.py`. It does the following:

1. Computes polynomial and rational approximations of a given function (e^x by default), 
returning the coefficients.

1. Computes (x,y) coordinates for every kind of approximation given the same x coordinates.
Plots the results for rough comparison.

1. Plots the results for rough comparison.

2. Computes the max error for every approximation given the same x coordinates.

3. Computes and plots max errors for every approximation with a varying number of parameters.

In other words, this script runs various approximation methods, plots their results and deltas
from actual function values. It can be configured to print the maximum error.
The exact behavior is controlled by the global variables at the top of `main.py`.

The following are the resources used to create this:
- <https://xn--2-umb.com/22/approximation>
- <https://sites.tufts.edu/atasissa/files/2019/09/remez.pdf>

In `main.py`, there is also an `exponent_approximation_choice` script.

This is a shorter and simpler version of `main` that isolates the 13-parameter
Chebyshev Rational approximation of e^x. We are planning to use it in production.
Therefore, we need to perform coefficient truncations to 36 decimal points
(the max `osmomath` supported precision). This truncation is applied
to `exponent_approximation_choice` but not `main`.

## Configuration

Several parameters can be changed on the needs basis at the
top of `main.py`.

Some of the parameters include the function we are approximating, the [x_start, x_end] range of
the approximation, and the number of terms to be used. For the full parameter list, see `main.py`.

## Usage

Install [uv](https://docs.astral.sh/uv/getting-started/installation/). From the
repository root:

```bash
# Install the locked environment (uv downloads Python 3.11 if needed).
uv sync --project scripts/approximations --locked

# Run the script without activating a virtual environment.
uv run --project scripts/approximations --locked python scripts/approximations/main.py

# Run tests, including headless plotting checks.
uv run --project scripts/approximations --locked python -m unittest discover -s scripts/approximations -p '*_test.py'
```

Python is restricted to 3.11 because the pinned NumPy and Matplotlib versions
predate Python 3.12, while Pillow requires 3.11 or newer. The unused PyQt5 pins
have been removed; Matplotlib uses its available default backend. Interactive
plots require a working GUI backend (for example, Tk installed with Python).
For headless execution, set `MPLBACKEND=Agg`.

`pyproject.toml` replaces `requirements.txt`; commit `uv.lock` when changing
dependencies. To update a pin, run `uv add --project scripts/approximations
'package==version'`, rerun the tests, and commit both files. The `--locked` flag
rejects stale lockfiles.
