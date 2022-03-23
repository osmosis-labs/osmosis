package balancer

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	MaxUserSpecifiedWeight    sdk.Int = sdk.NewIntFromUint64(1 << 20)
	GuaranteedWeightPrecision int64   = 1 << 30
)
