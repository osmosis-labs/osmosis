package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v7/x/superfluid/types"
)

type lockingStatus int64

const (
	unlockingStatus lockingStatus = iota
	bondedStatus
)

func (k Keeper) createSyntheticLockup(ctx sdk.Context,
	underlyingLockId uint64, intermediateAcc types.SuperfluidIntermediaryAccount, lockingStat lockingStatus) error {
	unbondingDuration := k.sk.GetParams(ctx).UnbondingTime
	if lockingStat == unlockingStatus {
		isUnlocking := true
		synthdenom := unstakingSyntheticDenom(intermediateAcc.Denom, intermediateAcc.ValAddr)
		return k.lk.CreateSyntheticLockup(ctx, underlyingLockId, synthdenom, unbondingDuration, isUnlocking)
	} else {
		notUnlocking := false
		synthdenom := stakingSyntheticDenom(intermediateAcc.Denom, intermediateAcc.ValAddr)
		return k.lk.CreateSyntheticLockup(ctx, underlyingLockId, synthdenom, unbondingDuration, notUnlocking)
	}
}
