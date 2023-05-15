package math

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
)

var (
	sdkOneInt      = sdk.OneInt()
	sdkOneDec      = sdk.NewDec(1)
	sdkNineDec     = sdk.NewDec(9)
	sdkTenDec      = sdk.NewDec(10)
	powersOfTen    []sdk.Dec
	negPowersOfTen []sdk.Dec

	osmomathBigOneDec = osmomath.NewBigDec(1)
	osmomathBigTenDec = osmomath.NewBigDec(10)
	bigPowersOfTen    []osmomath.BigDec
	bigNegPowersOfTen []osmomath.BigDec
)

// Set precision multipliers
func init() {
	negPowersOfTen = make([]sdk.Dec, sdk.Precision+1)
	for i := 0; i <= sdk.Precision; i++ {
		negPowersOfTen[i] = sdkOneDec.Quo(sdkTenDec.Power(uint64(i)))
	}
	// 10^77 < sdk.MaxInt < 10^78
	powersOfTen = make([]sdk.Dec, 78)
	for i := 0; i <= 77; i++ {
		powersOfTen[i] = sdkTenDec.Power(uint64(i))
	}

	bigNegPowersOfTen = make([]osmomath.BigDec, osmomath.Precision+1)
	for i := 0; i <= osmomath.Precision; i++ {
		bigNegPowersOfTen[i] = osmomathBigOneDec.Quo(osmomathBigTenDec.PowerInteger(uint64(i)))
	}
	// 10^308 < osmomath.MaxInt < 10^309
	bigPowersOfTen = make([]osmomath.BigDec, 309)
	for i := 0; i <= 308; i++ {
		bigPowersOfTen[i] = osmomathBigTenDec.PowerInteger(uint64(i))
	}
}
