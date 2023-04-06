# Concentrated Liquidity Test Case Estimate Scripts

## Context

These scripts exist to estimate the results of the swap
concentrated liqudity test cases.

Each existing test case is estimated by calling the
respective function in `scripts/cl/main.py` main function.

The function defining a swap test case specifies which
CL go test case it estimates. See the spec for details.

## Running the scripts

To run with sage installed:
```bash
sage -python scripts/cl/main.py
```

Note, these scripts can also be run by installing `sympy` without sage.

## Running tests

```bash
python3 -m unittest scripts.cl.test.test_common 
```
