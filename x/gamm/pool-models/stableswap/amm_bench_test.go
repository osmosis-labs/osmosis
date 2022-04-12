package stableswap

import (
	"math/rand"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func BenchmarkCFMM(b *testing.B) {
	// Uses solveCfmm
	for i := 0; i < b.N; i++ {
		runCalc(solveCfmm)
	}
}

func BenchmarkBinarySearch(b *testing.B) {
	twoDec, _ := sdk.NewDecFromStr("2.0")
	threshold, _ := sdk.NewDecFromStr("0.00001") // 0.001%
	// logic copied from cfmm.py
	approx_eq := func(a, b, tol sdk.Dec) bool {
		return a.Sub(b).Abs().Quo(a).LTE(tol)
	}
	solve := func(xReserve, yReserve, yIn sdk.Dec) sdk.Dec {
		k := cfmmConstant(xReserve, yReserve)
		yf := yReserve.Add(yIn)
		x_low_est := sdk.ZeroDec()
		x_high_est := xReserve
		x_est := (x_high_est.Add(x_low_est)).Quo(twoDec)
		cur_k := cfmmConstant(x_est, yf)
		for !approx_eq(cur_k, k, threshold) { // cap max iteration to 256
			if cur_k.GT(k) {
				x_high_est = x_est
			} else if cur_k.LT(k) {
				x_low_est = x_est
			}
			x_est = (x_high_est.Add(x_low_est)).Quo(twoDec)
			cur_k = cfmmConstant(x_est, yf)
		}
		return xReserve.Sub(x_est)
	}
	for i := 0; i < b.N; i++ {
		runCalc(solve)
	}
}

func BenchmarkApproximation(b *testing.B) {
	twoDec, _ := sdk.NewDecFromStr("2.0")
	threshold, _ := sdk.NewDecFromStr("0.00001") // 0.001%
	deriv := func(x, y sdk.Dec) sdk.Dec {
		// dy(x)/dx = -(3x^2y+y^3)/(x^3+3xy^2)
		//           = -8xy/3(x^2+3y^2)-y/3x
		x2 := x.Mul(x)
		y2 := y.Mul(y)
		termneg8xy := x.Mul(y).MulInt64(8).Neg()
		term9x23y2 := y2.MulInt64(3).Add(x2).MulInt64(3)
		termydiv3x := y.Quo(x.MulInt64(3))
		return termneg8xy.Quo(term9x23y2).Sub(termydiv3x)
	}

	approx := func(x, y, dy sdk.Dec) sdk.Dec {
		return x.Add(deriv(x, y).Mul(dy))
	}

	// 
	solve := func(xReserve, yReserve, yIn sdk.Dec) sdk.Dec {
		k := cfmmConstant(xReserve, yReserve)
		yf := yReserve.Add(yIn)
		//x_low_est := sdk.ZeroDec()
		//x_high_est := xReserve
		x_est := approx(xReserve, yReserve, yIn)
		cur_k := cfmmConstant(x_est, yf)
		for !approx_eq(cur_k, k, threshold) { // cap max iteration to 256
			/*
			if cur_k.GT(k) {
				x_high_est = x_est
			} else if cur_k.LT(k) {
				x_low_est = x_est
			}
			*/
			// x_est = (x_high_est.Add(x_low_est)).Quo(twoDec)
			// replace binary search 
			x_est = approx(x_est, )
			cur_k = cfmmConstant(x_est, yf)
		}
		return xReserve.Sub(x_est)
	}	

	for i := 0; i < b.N; i++ {
		runCalc(solve)
	}
}

func runCalc(solve func(sdk.Dec, sdk.Dec, sdk.Dec) sdk.Dec) {
	xReserve := sdk.NewDec(rand.Int63n(100000) + 50000)
	yReserve := sdk.NewDec(rand.Int63n(100000) + 50000)
	yIn := sdk.NewDec(rand.Int63n(100000))
	solve(xReserve, yReserve, yIn)
}
