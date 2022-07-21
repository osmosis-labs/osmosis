package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v10/x/superfluid/types"
)

var (
	StakingSyntheticDenom   = stakingSyntheticDenom
	UnstakingSyntheticDenom = unstakingSyntheticDenom
)

func (k Keeper) DeleteIntermediaryAccountIfNoDelegation(ctx sdk.Context, intermediaryAcc types.SuperfluidIntermediaryAccount) {
	k.deleteIntermediaryAccountIfNoDelegation(ctx, intermediaryAcc)
}
