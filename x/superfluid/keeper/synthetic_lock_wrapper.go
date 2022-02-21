package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v7/x/superfluid/types"
)

type bondingStatus int64

const (
	Unbonding bondingStatus = iota
	Bonded
)

func (k Keeper) createSyntheticLockup(ctx sdk.Context,
	underlyingLockId uint64, intemediateAccount types.SuperfluidIntermediaryAccount, bond bondingStatus) {

}
