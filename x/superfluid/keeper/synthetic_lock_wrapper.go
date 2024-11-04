package keeper

import (
	"fmt"
	"strings"
	"time"

	"github.com/osmosis-labs/osmosis/v27/x/superfluid/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TODO: move this into types and have it exported
func stakingSyntheticDenom(denom, valAddr string) string {
	return fmt.Sprintf("%s/superbonding/%s", denom, valAddr)
}

func unstakingSyntheticDenom(denom, valAddr string) string {
	return fmt.Sprintf("%s/superunbonding/%s", denom, valAddr)
}

// quick fix for getting the validator address from a synthetic denom.
func ValidatorAddressFromSyntheticDenom(syntheticDenom string) (string, error) {
	if strings.Contains(syntheticDenom, "superbonding") {
		splitString := strings.Split(syntheticDenom, "/superbonding/")
		lastComponent := splitString[len(splitString)-1]
		return lastComponent, nil
	}
	if strings.Contains(syntheticDenom, "superunbonding") {
		splitString := strings.Split(syntheticDenom, "/superunbonding/")
		lastComponent := splitString[len(splitString)-1]
		return lastComponent, nil
	}
	return "", fmt.Errorf("%s is not a valid synthetic denom suffix", syntheticDenom)
}

type lockingStatus int64

const (
	unlockingStatus lockingStatus = iota
	bondedStatus
)

func (k Keeper) createSyntheticLockup(ctx sdk.Context,
	underlyingLockId uint64, intermediateAcc types.SuperfluidIntermediaryAccount, lockingStat lockingStatus,
) error {
	stakingParams, err := k.sk.GetParams(ctx)
	if err != nil {
		return err
	}
	unbondingDuration := stakingParams.UnbondingTime
	return k.createSyntheticLockupWithDuration(ctx, underlyingLockId, intermediateAcc, unbondingDuration, lockingStat)
}

func (k Keeper) createSyntheticLockupWithDuration(ctx sdk.Context,
	underlyingLockId uint64, intermediateAcc types.SuperfluidIntermediaryAccount, unlockingDuration time.Duration, lockingStat lockingStatus,
) error {
	if lockingStat == unlockingStatus {
		isUnlocking := true
		synthdenom := unstakingSyntheticDenom(intermediateAcc.Denom, intermediateAcc.ValAddr)
		return k.lk.CreateSyntheticLockup(ctx, underlyingLockId, synthdenom, unlockingDuration, isUnlocking)
	} else {
		notUnlocking := false
		synthdenom := stakingSyntheticDenom(intermediateAcc.Denom, intermediateAcc.ValAddr)
		return k.lk.CreateSyntheticLockup(ctx, underlyingLockId, synthdenom, unlockingDuration, notUnlocking)
	}
}
