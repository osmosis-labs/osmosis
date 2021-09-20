package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	MinPoolAssets = 2
	MaxPoolAssets = 8

	OneShareExponent                = 18
	GuaranteedWeightPrecision int64 = 1 << 30
)

var (
	_ PoolI = (*Pool)(nil)

	// OneShare represents the amount of subshares in a single pool share
	OneShare = sdk.NewIntWithDecimal(1, OneShareExponent)

	// InitPoolSharesSupply is the amount of new shares to initialize a pool with
	InitPoolSharesSupply = OneShare.MulRaw(100)

	MaxUserSpecifiedWeight sdk.Int = sdk.NewIntFromUint64(1 << 20)
)
