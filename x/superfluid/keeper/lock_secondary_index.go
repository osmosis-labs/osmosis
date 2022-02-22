package keeper

import (
	"fmt"
	"strings"
)

func stakingSecondaryIndex(denom, valAddr string) string {
	return fmt.Sprintf("%s/superbonding/%s", denom, valAddr)
}

func unstakingSecondaryIndex(denom, valAddr string) string {
	return fmt.Sprintf("%s/superunbonding/%s", denom, valAddr)
}

// quick fix for getting the validator addresss from a synthetic denom
func ValidatorAddressFromSyntheticDenom(suffix string) (string, error) {
	if strings.Contains(suffix, "superbonding") {
		return strings.TrimLeft(suffix, "/superbonding/"), nil
	}
	if strings.Contains(suffix, "superunbonding") {
		return strings.TrimLeft(suffix, "/superunbonding/"), nil
	}
	return "", fmt.Errorf("%s is not a valid synthetic denom suffix", suffix)
}

type lockingStatus int64

const (
	unlockedStatus lockingStatus = iota
	unlockingStatus
	bondedStatus
)

// func (k Keeper) SetLockSuperfluidBonded(ctx sdk.Context, lockId uint64, denom, validatorAddr string) error {
// 	k.lk.AddSecondaryIndex()
// }

// func (k Keeper) createSyntheticLockup(ctx sdk.Context,
// 	underlyingLockId uint64, intermediateAcc types.SuperfluidIntermediaryAccount, lockingStat lockingStatus) error {
// 	unbondingDuration := k.sk.GetParams(ctx).UnbondingTime
// 	if lockingStat == unlockingStatus {
// 		isUnlocking := true
// 		synthdenom := unstakingSyntheticDenom(intermediateAcc.Denom, intermediateAcc.ValAddr)
// 		return k.lk.CreateSyntheticLockup(ctx, underlyingLockId, synthdenom, unbondingDuration, isUnlocking)
// 	} else {
// 		notUnlocking := false
// 		synthdenom := stakingSyntheticDenom(intermediateAcc.Denom, intermediateAcc.ValAddr)
// 		return k.lk.CreateSyntheticLockup(ctx, underlyingLockId, synthdenom, unbondingDuration, notUnlocking)
// 	}
// }
