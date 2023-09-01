package stableswap

import (
<<<<<<< HEAD
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v18/x/gamm/types"
=======
	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v19/x/gamm/types"
>>>>>>> ca75f4c3 (refactor(deps): switch to cosmossdk.io/math from fork math (#6238))
)

func (params PoolParams) Validate() error {
	if params.ExitFee.IsNegative() {
		return types.ErrNegativeExitFee
	}

	if params.ExitFee.GTE(osmomath.OneDec()) {
		return types.ErrTooMuchExitFee
	}

	if params.SwapFee.IsNegative() {
		return types.ErrNegativeSpreadFactor
	}

	if params.SwapFee.GTE(osmomath.OneDec()) {
		return types.ErrTooMuchSpreadFactor
	}
	return nil
}
