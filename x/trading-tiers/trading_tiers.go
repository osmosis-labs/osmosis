package tradingtiers

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/gogoproto/proto"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v28/x/trading-tiers/types"
)

func (k Keeper) TrackTradingTierVolume(ctx sdk.Context, addr sdk.AccAddress, volumeGeneratedInt osmomath.Int, initialDenom string) error {
	// Check if the account has opted in to trading tier tracking.
	if !k.IsAccountTradeTierOptIn(ctx, addr.String()) {
		return nil
	}

	// Check if token is a whitelisted fee token.
	found, err := k.txFeesKeeper.IsFeeToken(ctx, initialDenom)
	if err != nil {
		return err
	}
	if !found {
		return nil
	}

	err = k.IncreaseAccountDailyOsmoVolume(ctx, addr.String(), volumeGeneratedInt)
	if err != nil {
		return err
	}

	return nil
}

func (k Keeper) IsAccountTradeTierOptIn(ctx sdk.Context, acc string) bool {
	store := ctx.KVStore(k.storeKey)
	key := types.FormatAccountTierOptInKey(acc)
	return store.Has(key)
}

func (k Keeper) IncreaseAccountDailyOsmoVolume(ctx sdk.Context, acc string, incBy osmomath.Int) error {
	store := ctx.KVStore(k.storeKey)
	key := types.FormatAccountDailyOsmoVolumeKey(0, acc)
	currentVolume := sdk.IntProto{}
	_, err := osmoutils.Get(store, key, &currentVolume)
	if err != nil {
		return err
	}
	newVolume := currentVolume.Int.Add(incBy)
	newVolumeProto := sdk.IntProto{Int: newVolume}
	bz, err := proto.Marshal(&newVolumeProto)
	if err != nil {
		return err
	}
	store.Set(key, bz)
	return nil
}

func (k Keeper) GetOsmoUsdValueForEpoch(ctx sdk.Context, epochNum int64) (osmomath.Int, error) {
	store := ctx.KVStore(k.storeKey)
	key := types.FormatOsmoUSDValueKey(epochNum)
	value := sdk.IntProto{}
	_, err := osmoutils.Get(store, key, &value)
	if err != nil {
		return osmomath.Int{}, err
	}
	return value.Int, nil
}

func (k Keeper) SetOsmoUsdValueForEpoch(ctx sdk.Context, epochNum int64, value osmomath.Int) error {
	store := ctx.KVStore(k.storeKey)
	key := types.FormatOsmoUSDValueKey(epochNum)
	valueProto := sdk.IntProto{Int: value}
	bz, err := proto.Marshal(&valueProto)
	if err != nil {
		return err
	}
	store.Set(key, bz)
	return nil
}

// OptInToTradingTier allows an account to opt in to trading tier tracking.
func (k Keeper) OptInToTradingTier(ctx sdk.Context, addr string) error {
	// Check if the account has already opted in.
	store := ctx.KVStore(k.storeKey)
	key := types.FormatAccountTierOptInKey(addr)
	if store.Has(key) {
		return types.ErrAccountAlreadyOptedIn
	}

	// Check that the account has enough stake to opt in.
	minStake := sdk.IntProto{}
	k.paramSpace.Get(ctx, types.KeyStakeMinRequirement, &minStake)
	bondedAmt, err := k.stakingKeeper.GetDelegatorBonded(ctx, sdk.AccAddress(addr))
	if err != nil {
		return err
	}
	if bondedAmt.LT(minStake.Int) {
		return types.InsufficientStakeError{MinStake: minStake.Int, BondedAmt: bondedAmt}
	}

	// Opt in the account.
	store.Set(key, []byte{0x01})
	return nil
}
