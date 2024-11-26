package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	lockuptypes "github.com/osmosis-labs/osmosis/v27/x/lockup/types"
	"github.com/osmosis-labs/osmosis/v27/x/valset-pref/types"
)

type (
	ValSet = valSet
)

func (k Keeper) ValidateLockForForceUnlock(ctx sdk.Context, lockID uint64, delegatorAddr string) (*lockuptypes.PeriodLock, osmomath.Int, error) {
	return k.validateLockForForceUnlock(ctx, lockID, delegatorAddr)
}

func (k Keeper) GetValsetRatios(ctx sdk.Context, delegator sdk.AccAddress,
	prefs []types.ValidatorPreference, undelegateAmt osmomath.Int) ([]ValRatio, map[string]stakingtypes.Validator, osmomath.Dec, error) {
	return k.getValsetRatios(ctx, delegator, prefs, undelegateAmt)
}

func (k Keeper) FormatToValPrefArr(ctx sdk.Context, delegations []stakingtypes.Delegation) ([]types.ValidatorPreference, error) {
	return k.formatToValPrefArr(ctx, delegations)
}
