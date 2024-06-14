package tradingtiers

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/gogoproto/proto"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v25/x/tradingtiers/types"
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
	key := types.FormatAccountDailyOsmoVolumeKey("", acc)
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

func (k Keeper) GetOsmoUsdValueForEpoch(ctx sdk.Context, epochNum string) (osmomath.Int, error) {
	store := ctx.KVStore(k.storeKey)
	key := types.FormatOsmoUSDValueKey(epochNum)
	value := sdk.IntProto{}
	_, err := osmoutils.Get(store, key, &value)
	if err != nil {
		return osmomath.Int{}, err
	}
	return value.Int, nil
}
