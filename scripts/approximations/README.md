# Math Operations Approximations

## Context

This is a script to approximate a mathematical operation using polynomial
and rational approximations.

This script does the following:

1. Computes polynomial and rational approximations of a given function (e^x by default), 
returning the coefficients.

2. Computes (x,y) coordinates for every kind of approximation given the same x coordinates.
Plots the results for rough comparison.


The following are the resources used to write the script:
https://xn--2-umb.com/22/approximation/
https://sites.tufts.edu/atasissa/files/2019/09/remez.pdf

## Configuration

There are several parameters that can be changed on the needs basis at the
top of `main.py`.

Some of the parameters include the function we are approximating, the [x_start, x_end] range of
the approximation, and the number of terms to be used. For the full parameter list, see `main.py`.

## Usage

Assuming that you are in the root of the repository:

```bash
# Create a virtual environment.
python3 -m venv scripts/approximations/venv

# Start the environment
source scripts/approximations/venv/bin/activate

# Install dependencies in the virtual environment.
pip install -r scripts/approximations/requirements.txt

# Run the script.
python3 scripts/approximations/main.py

# Run tests
python3 scripts/approximations/rational_test.py
```
