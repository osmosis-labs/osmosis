# This file is for solving for close enough to exact solutions
# to the CFMM equation, to generate test vectors for amm_test.go

cfmm = lambda x,y: x*y*(x*x + y*y)

x0, y0 = 100, 100
yin = 1000
err_threshold = .001

def approx_eq(a, b, tol):
    return abs(a - b) <= tol

def binary_search(x0, y0, yin):
    k = cfmm(x0, y0)
    yf = y0 + yin
    x_low_est = 0
    x_high_est = x0
    x_est = (x_high_est + x_low_est) / 2.
    cur_k = cfmm(x_est, yf)
    while not approx_eq(cur_k, k, err_threshold):
        # x is too high
        if cur_k > k:
            x_high_est = x_est
        elif cur_k < k:
            x_low_est = x_est
        x_est = (x_high_est + x_low_est) / 2.
        cur_k = cfmm(x_est, yf)
    
    return x0 - x_est

xOut = binary_search(x0, y0, yin)
print(cfmm(x0, y0))
print(cfmm(x0 - xOut, y0 + yin))
print(xOut)