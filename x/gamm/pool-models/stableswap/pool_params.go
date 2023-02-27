package stableswap

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v15/x/gamm/types"
)

func (params PoolParams) Validate() error {
	if params.SwapFee.IsNegative() {
		return types.ErrNegativeSwapFee
	}

	if params.SwapFee.GTE(sdk.OneDec()) {
		return types.ErrTooMuchSwapFee
	}
	return nil
}
