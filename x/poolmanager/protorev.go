package poolmanager

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	gogotypes "github.com/cosmos/gogoproto/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v21/x/poolmanager/types"

	"github.com/cosmos/cosmos-sdk/store/prefix"
)

// GetTakerFeeTrackerForStakers returns the taker fee for stakers tracker for all denoms that has been
// collected since the accounting height.
func (k Keeper) GetTakerFeeTrackerForStakers(ctx sdk.Context) []sdk.Coin {
	takerFeesForStakers := make([]sdk.Coin, 0)

	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.KeyTakerFeeStakersProtoRevArray)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		bz := iterator.Value()
		takerFeeForStakers := sdk.Coin{}
		if err := takerFeeForStakers.Unmarshal(bz); err == nil {
			takerFeesForStakers = append(takerFeesForStakers, takerFeeForStakers)
		}
	}

	return takerFeesForStakers
}

// GetTakerFeeTrackerForStakersByDenom returns the taker fee for stakers tracker for the specified denom that has been
// collected since the accounting height.
func (k Keeper) GetTakerFeeTrackerForStakersByDenom(ctx sdk.Context, denom string) (sdk.Coin, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyTakerFeeStakersProtoRevArray)
	key := types.GetKeyPrefixTakerFeeStakersProtoRevByDenom(denom)

	bz := store.Get(key)
	if len(bz) == 0 {
		return sdk.NewCoin(denom, osmomath.ZeroInt()), nil
	}

	takerFeeForStakers := sdk.Coin{}
	if err := takerFeeForStakers.Unmarshal(bz); err != nil {
		return sdk.NewCoin(denom, osmomath.ZeroInt()), err
	}

	return takerFeeForStakers, nil
}

// UpdateTakerFeeTrackerForStakersByDenom increases the take fee for stakers tracker for the specified denom by the specified amount.
func (k Keeper) UpdateTakerFeeTrackerForStakersByDenom(ctx sdk.Context, denom string, increasedAmt osmomath.Int) error {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyTakerFeeStakersProtoRevArray)
	key := types.GetKeyPrefixTakerFeeStakersProtoRevByDenom(denom)

	takerFeeForStakers, _ := k.GetTakerFeeTrackerForStakersByDenom(ctx, denom)
	takerFeeForStakers.Amount = takerFeeForStakers.Amount.Add(increasedAmt)
	bz, err := takerFeeForStakers.Marshal()
	if err != nil {
		return err
	}

	store.Set(key, bz)
	return nil
}

// GetTakerFeeTrackerForCommunityPool returns the taker fee for community pool tracker for all denoms that has been
// collected since the accounting height.
func (k Keeper) GetTakerFeeTrackerForCommunityPool(ctx sdk.Context) []sdk.Coin {
	takerFeesForCommunityPool := make([]sdk.Coin, 0)

	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.KeyTakerFeeCommunityPoolProtoRevArray)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		bz := iterator.Value()
		takerFeeForCommunityPool := sdk.Coin{}
		if err := takerFeeForCommunityPool.Unmarshal(bz); err == nil {
			takerFeesForCommunityPool = append(takerFeesForCommunityPool, takerFeeForCommunityPool)
		}
	}

	return takerFeesForCommunityPool
}

// GetTakerFeeTrackerForCommunityPoolByDenom returns the taker fee for community pool tracker for the specified denom that has been
// collected since the accounting height.
func (k Keeper) GetTakerFeeTrackerForCommunityPoolByDenom(ctx sdk.Context, denom string) (sdk.Coin, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyTakerFeeCommunityPoolProtoRevArray)
	key := types.GetKeyPrefixTakerFeeCommunityPoolProtoRevByDenom(denom)

	bz := store.Get(key)
	if len(bz) == 0 {
		return sdk.NewCoin(denom, osmomath.ZeroInt()), nil
	}

	takerFeeForCommunityPool := sdk.Coin{}
	if err := takerFeeForCommunityPool.Unmarshal(bz); err != nil {
		return sdk.NewCoin(denom, osmomath.ZeroInt()), err
	}

	return takerFeeForCommunityPool, nil
}

// UpdateTakerFeeTrackerForCommunityPoolByDenom increases the take fee for community pool tracker for the specified denom by the specified amount.
func (k Keeper) UpdateTakerFeeTrackerForCommunityPoolByDenom(ctx sdk.Context, denom string, increasedAmt osmomath.Int) error {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyTakerFeeCommunityPoolProtoRevArray)
	key := types.GetKeyPrefixTakerFeeCommunityPoolProtoRevByDenom(denom)

	takerFeeForCommunityPool, _ := k.GetTakerFeeTrackerForCommunityPoolByDenom(ctx, denom)
	takerFeeForCommunityPool.Amount = takerFeeForCommunityPool.Amount.Add(increasedAmt)
	bz, err := takerFeeForCommunityPool.Marshal()
	if err != nil {
		return err
	}

	store.Set(key, bz)
	return nil
}

// GetTakerFeeTrackerStartHeight gets the height from which we started accounting for taker fees.
func (k Keeper) GetTakerFeeTrackerStartHeight(ctx sdk.Context) int64 {
	startHeight := gogotypes.Int64Value{}
	osmoutils.MustGet(ctx.KVStore(k.storeKey), types.KeyTakerFeeProtoRevAccountingHeight, &startHeight)
	return startHeight.Value
}

// SetTakerFeeTrackerStartHeight sets the height from which we started accounting for taker fees.
func (k Keeper) SetTakerFeeTrackerStartHeight(ctx sdk.Context, startHeight int64) {
	osmoutils.MustSet(ctx.KVStore(k.storeKey), types.KeyTakerFeeProtoRevAccountingHeight, &gogotypes.Int64Value{Value: startHeight})
}
