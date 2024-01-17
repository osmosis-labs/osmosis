package poolmanager

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	gogotypes "github.com/cosmos/gogoproto/types"

	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v21/x/poolmanager/types"
)

// IncreaseTakerFeeTrackerForStakers gets the current value of the taker fee tracker for stakers, adds the given amount to it, and sets the new value.
func (k Keeper) IncreaseTakerFeeTrackerForStakers(ctx sdk.Context, takerFeeForStakers sdk.Coin) {
	currentTakerFeeForStakers := k.GetTakerFeeTrackerForStakers(ctx)
	if !takerFeeForStakers.IsZero() {
		newTakerFeeForStakersCoins := currentTakerFeeForStakers.Add(takerFeeForStakers)
		newTakerFeeForStakers := types.TrackedVolume{
			Amount: newTakerFeeForStakersCoins,
		}
		osmoutils.MustSet(ctx.KVStore(k.storeKey), types.KeyTakerFeeStakersProtoRev, &newTakerFeeForStakers)
	}
}

// IncreaseTakerFeeTrackerForCommunityPool gets the current value of the taker fee tracker for the community pool, adds the given amount to it, and sets the new value.
func (k Keeper) IncreaseTakerFeeTrackerForCommunityPool(ctx sdk.Context, takerFeeForCommunityPool sdk.Coin) {
	currentTakerFeeForCommunityPool := k.GetTakerFeeTrackerForCommunityPool(ctx)
	if !takerFeeForCommunityPool.IsZero() {
		newTakerFeeForCommunityPoolCoins := currentTakerFeeForCommunityPool.Add(takerFeeForCommunityPool)
		newTakerFeeForCommunityPool := types.TrackedVolume{
			Amount: newTakerFeeForCommunityPoolCoins,
		}
		osmoutils.MustSet(ctx.KVStore(k.storeKey), types.KeyTakerFeeCommunityPoolProtoRev, &newTakerFeeForCommunityPool)
	}
}

func (k Keeper) SetTakerFeeTrackerForStakers(ctx sdk.Context, takerFeeForStakers sdk.Coins) {
	newTakerFeeForStakers := types.TrackedVolume{
		Amount: takerFeeForStakers,
	}
	osmoutils.MustSet(ctx.KVStore(k.storeKey), types.KeyTakerFeeStakersProtoRev, &newTakerFeeForStakers)
}

func (k Keeper) SetTakerFeeTrackerForCommunityPool(ctx sdk.Context, takerFeeForCommunityPool sdk.Coins) {
	newTakerFeeForCommunityPool := types.TrackedVolume{
		Amount: takerFeeForCommunityPool,
	}
	osmoutils.MustSet(ctx.KVStore(k.storeKey), types.KeyTakerFeeCommunityPoolProtoRev, &newTakerFeeForCommunityPool)
}

func (k Keeper) GetTakerFeeTrackerForStakers(ctx sdk.Context) (currentTakerFeeForStakers sdk.Coins) {
	var takerFeeForStakers types.TrackedVolume
	takerFeeFound, err := osmoutils.Get(ctx.KVStore(k.storeKey), types.KeyTakerFeeStakersProtoRev, &takerFeeForStakers)
	if err != nil {
		// We can only encounter an error if a database or serialization errors occurs, so we panic here.
		// Normally this would be handled by `osmoutils.MustGet`, but since we want to specifically use `osmoutils.Get`,
		// we also have to manually panic here.
		panic(err)
	}

	// If no volume was found, we treat the existing volume as 0.
	// While we can technically require volume to exist, we would need to store empty coins in state for each pool (past and present),
	// which is a high storage cost to pay for a weak guardrail.
	currentTakerFeeForStakers = sdk.Coins(nil)
	if takerFeeFound {
		currentTakerFeeForStakers = takerFeeForStakers.Amount
	}

	return currentTakerFeeForStakers
}

func (k Keeper) GetTakerFeeTrackerForCommunityPool(ctx sdk.Context) (currentTakerFeeForCommunityPool sdk.Coins) {
	var takerFeeForCommunityPool types.TrackedVolume
	takerFeeFound, err := osmoutils.Get(ctx.KVStore(k.storeKey), types.KeyTakerFeeCommunityPoolProtoRev, &takerFeeForCommunityPool)
	if err != nil {
		// We can only encounter an error if a database or serialization errors occurs, so we panic here.
		// Normally this would be handled by `osmoutils.MustGet`, but since we want to specifically use `osmoutils.Get`,
		// we also have to manually panic here.
		panic(err)
	}

	// If no volume was found, we treat the existing volume as 0.
	// While we can technically require volume to exist, we would need to store empty coins in state for each pool (past and present),
	// which is a high storage cost to pay for a weak guardrail.
	currentTakerFeeForCommunityPool = sdk.Coins(nil)
	if takerFeeFound {
		currentTakerFeeForCommunityPool = takerFeeForCommunityPool.Amount
	}

	return currentTakerFeeForCommunityPool
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
