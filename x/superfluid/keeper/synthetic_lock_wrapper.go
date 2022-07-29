package keeper

import (
	"fmt"
	"strings"

	"github.com/osmosis-labs/osmosis/v10/x/superfluid/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func stakingSyntheticDenom(denom, valAddr string) string {
	return fmt.Sprintf("%s/superbonding/%s", denom, valAddr)
}

func unstakingSyntheticDenom(denom, valAddr string) string {
	return fmt.Sprintf("%s/superunbonding/%s", denom, valAddr)
}

// quick fix for getting the validator addresss from a synthetic denom.
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
