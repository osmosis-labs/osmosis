import matplotlib.pyplot as plt
import sympy as sp


approximated_fn = lambda x: sp.Pow(2, x)

# fixed point precision used in Osmosis `osmomath` package.
osmomath_precision = 100

def main():

    x = sp.Float("127.84864288", osmomath_precision + 10)

    print(x)

    # x_f = sp.Float(x, osmomath_precision +20)

    res = approximated_fn(x)

    print(res)

    # diff = sp.Pow(2, 128) * (sp.Float(1, osmomath_precision) - sp.Float("0.999999999999999999999999999999999999"))
    # print(f"diff 128: {diff}, prec 36")

    # diff64 = sp.Pow(2, 64) * (sp.Float(1, osmomath_precision) - sp.Float("0.99999999999999999999"))
    # print(f"diff 64: {diff64}")
    


if __name__ == "__main__":
    
    main()
