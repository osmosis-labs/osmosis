package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	MinPoolAssets = 2
	MaxPoolAssets = 8

	OneShareExponent = 18
	// Raise 10 to the power of SigFigsExponent to determine number of significant figures.
	// i.e. SigFigExponent = 8 is 10^8 which is 100000000. This gives 8 significant figures.
	SigFigsExponent = 8
	// TODO: Current fixed cost gas fee per swap -- turn this into a param in the future.
	GasFeeForSwap = 10000
)

var (
	// OneShare represents the amount of subshares in a single pool share.
	OneShare = sdk.NewIntWithDecimal(1, OneShareExponent)

	// InitPoolSharesSupply is the amount of new shares to initialize a pool with.
	InitPoolSharesSupply = OneShare.MulRaw(100)

	// SigFigs is the amount of significant figures used to calculate SpotPrice
	SigFigs = sdk.NewDec(10).Power(SigFigsExponent).TruncateInt()
)
