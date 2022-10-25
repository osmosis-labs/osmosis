package types

import (
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func OrderInitialPoolDenoms(denom0, denom1 string) (string, string, error) {
	if denom0 == denom1 {
		return "", "", fmt.Errorf("cannot have the same asset in a single pool")
	}
	if denom0 > denom1 {
		denom1, denom0 = denom0, denom1
	}

	return denom0, denom1, nil
}

func AddLiquidity(x, y sdk.Dec) (z sdk.Dec) {
	if y.LT(sdk.ZeroDec()) {
		z = x.Sub(y.Neg())
		return z
	} else {
		z = x.Add(y)
		return z
	}
}
