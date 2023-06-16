package osmomath

import (
	"errors"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var smallestDec = sdk.SmallestDec()

// Returns square root of d
// returns an error iff one of the following conditions is met:
// - d is negative
// - d is too small to have a representable square root.
// This function guarantees:
// the returned root r, will be such that r^2 >= d
// This function is monotonic, i.e. if d1 >= d2, then sqrt(d1) >= sqrt(d2)
func MonotonicSqrt(d sdk.Dec) (sdk.Dec, error) {
	if d.IsNegative() {
		return d, errors.New("cannot take square root of negative number")
	}
	if d.IsZero() {
		return sdk.ZeroDec(), nil
	}

	// we run newton's approximation on f(x) = x^2 - d
	// which has derivative f'(x) = 2 * x
	// x_{n+1} = x_n - f(x_n)/f'(x_n)
	// f(x_n)/f'(x_n) = (x_n^2 - d) / (2 * x_n) = x_n/2 - d/(2 * x_n)
	// newton's method will be a monotonic sequence converging to the root.
	// source: https://math.mit.edu/~stevenj/18.335/newton-sqrt.pdf
	// So to ensure were over-estimating, we just need to get an initial over-estimate.
	guess := getInitialSquareRootGuess(d)
	delta := sdk.OneDec()
	prev := guess
	var iter = 0

	for ; delta.AbsMut().GT(smallestDec) && iter < maxApproxRootIterations; iter++ {
		// prev = guess
		prev = guess
		if prev.IsZero() {
			prev = smallestDec
		}
		// delta = ((d/prev) - guess)/root
		delta.Set(d).QuoMut(prev)
		delta.SubMut(guess)
		delta.QuoInt64Mut(int64(2))

		guess.AddMut(delta)
	}

	if iter == maxApproxRootIterations {
		return guess, errors.New("failed to converge")
	}
	// Now we have a sqrt answer, up to some accuracy bound.
	// We want to get monotonicity across various inputs.
	// e.g. if d1 <= d2, then sqrt(d1) <= sqrt(d2)
	//
	// We know that the update we were about to do is delta <= smallestDec.
	// furthermore we know that every successive iteration, if we had higher precision, would be much smaller.
	// We also know that once were close to the solution (d_n << 1),
	// the relative error d_n = |(x_n - x) / x|, obeys:
	// d_n+1 = (d_n^2) / 2 + O(d_n^3)
	// Note that we are always "close" to the solution in this case.
	// The smallest number we can take a square root of is 10^-18, whose square root is ~10^-9.
	// so a difference of 10^-18 is quite small (were already mostly correct for 7+ digits of the square root).
	//
	// Therefore our d_n is then on the order of magnitude 10^-18, and if we had infinite precision, d_{n+1} would go to
	// 10^-36.
	//
	// To argue monotonicity off this alone requires reasoning about the following points:
	// - no edge effects around this binary representation, especially with the quotient behavior
	// -
	// TODO: Seems like we could just reason about the quo behavior, round the 10^-18 term (really final 3 bits),
	// and be certain.
	return guess, nil
}

var oneDec = sdk.OneDec()
var oneDecBigInt = sdk.OneDec().BigInt()
var oneDecBigIntMinusOne = big.NewInt(0).Sub(oneDecBigInt, big.NewInt(1))

// returns a guess for square root(d), that is greater than d.
func getInitialSquareRootGuess(d sdk.Dec) sdk.Dec {
	// the underlying bigint value of d is 10^18 * v. (v can be less than 1)
	// we ignore the case of d < 1, we just over-estimate by returning 1.
	// TODO: Consider optimizing.
	if d.LTE(oneDec) {
		return sdk.OneDec()
	}
	// The strategy for the case where d > 1, is to first get ceil(v) = (d+10^18 -1)/10^18.
	//
	// we then get an initial square root guess of ceil(v).
	// we compute this based on a windowed method.
	// we think of v = a * 2^{2n}
	// for an 8 bit integer a.
	// Then we compute ceil(sqrt(v)) = ceil(sqrt(a)) * 2^n
	bigIntD := d.BigInt()
	v := bigIntD.Add(oneDecBigIntMinusOne, bigIntD)
	v = v.Quo(v, oneDecBigInt)
	bitlen := v.BitLen()

	// we find a by computing (v >> (2n - 8)) + 1.
	// n = (v.bitlen() + 1)/2.
	// if v.bitlen() is odd, then we only take a leading 7 bits.
	// if n < 8, we set a = v.
	n := (bitlen + 1 - 8) / 2
	var a int64
	if n < 0 {
		a = v.Int64()
	} else {
		a = big.NewInt(0).Rsh(v, uint(n*2)).Int64() + 1
	}
	if a > 256 {
		panic("code error")
	}
	var sqrtA int64
	if a <= 3 {
		sqrtA = a
	} else if a <= 16 {
		sqrtA = 4
	} else if a <= 64 {
		sqrtA = 8
	} else {
		sqrtA = 16
	}
	res := big.NewInt(sqrtA)
	if n > 0 {
		res = res.Lsh(res, uint(n))
	}
	// handles scaling by 10^18.
	return sdk.NewDecFromBigInt(res)
}
