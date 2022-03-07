package poolmodels

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type StandardCFMM interface {
	SolveCFMMamountOut(ctx sdk.Context, swapFee sdk.Dec, balanceInBefore, balanceInAfter sdk.Dec, denomIn string, balanceOut sdk.Dec, denomOut string, amount sdk.Dec) (sdk.Dec, error)
	// NewLPShares()
	SpotPrice(ctx sdk.Context, asset1, asset2 string) (sdk.Dec, error)
}

type CfmmPoolWrapper struct {
	UnderlyingCFMM StandardCFMM
}
