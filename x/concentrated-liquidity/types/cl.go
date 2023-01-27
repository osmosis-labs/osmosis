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

func GetInitialUptimeAccums() []sdk.Dec {
	initUptimeAccums := make([]sdk.Dec, len(SupportedUptimes))
	for uptimeIndex := range SupportedUptimes {
		initUptimeAccums[uptimeIndex] = sdk.NewDec(0)
	}

	return initUptimeAccums
}
