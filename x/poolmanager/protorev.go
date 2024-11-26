package poolmanager

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	gogotypes "github.com/cosmos/gogoproto/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v27/x/poolmanager/types"
)

// GetTakerFeeTrackerForStakers returns the taker fee for stakers tracker for all denoms that has been
// collected since the accounting height.
func (k Keeper) GetTakerFeeTrackerForStakers(ctx sdk.Context) []sdk.Coin {
	return osmoutils.GetCoinArrayFromPrefix(ctx, k.storeKey, types.KeyTakerFeeStakersProtoRevArray)
}

// GetTakerFeeTrackerForStakersByDenom returns the taker fee for stakers tracker for the specified denom that has been
// collected since the accounting height. If the denom is not found, a zero coin is returned.
func (k Keeper) GetTakerFeeTrackerForStakersByDenom(ctx sdk.Context, denom string) (sdk.Coin, error) {
	return osmoutils.GetCoinByDenomFromPrefix(ctx, k.storeKey, types.KeyTakerFeeStakersProtoRevArray, denom)
}

// UpdateTakerFeeTrackerForStakersByDenom increases the take fee for stakers tracker for the specified denom by the specified amount.
func (k Keeper) UpdateTakerFeeTrackerForStakersByDenom(ctx sdk.Context, denom string, increasedAmt osmomath.Int) error {
	return osmoutils.IncreaseCoinByDenomFromPrefix(ctx, k.storeKey, types.KeyTakerFeeStakersProtoRevArray, denom, increasedAmt)
}

// GetTakerFeeTrackerForCommunityPool returns the taker fee for community pool tracker for all denoms that has been
// collected since the accounting height.
func (k Keeper) GetTakerFeeTrackerForCommunityPool(ctx sdk.Context) []sdk.Coin {
	return osmoutils.GetCoinArrayFromPrefix(ctx, k.storeKey, types.KeyTakerFeeCommunityPoolProtoRevArray)
}

// GetTakerFeeTrackerForCommunityPoolByDenom returns the taker fee for community pool tracker for the specified denom that has been
// collected since the accounting height. If the denom is not found, a zero coin is returned.
func (k Keeper) GetTakerFeeTrackerForCommunityPoolByDenom(ctx sdk.Context, denom string) (sdk.Coin, error) {
	return osmoutils.GetCoinByDenomFromPrefix(ctx, k.storeKey, types.KeyTakerFeeCommunityPoolProtoRevArray, denom)
}

// UpdateTakerFeeTrackerForCommunityPoolByDenom increases the take fee for community pool tracker for the specified denom by the specified amount.
func (k Keeper) UpdateTakerFeeTrackerForCommunityPoolByDenom(ctx sdk.Context, denom string, increasedAmt osmomath.Int) error {
	return osmoutils.IncreaseCoinByDenomFromPrefix(ctx, k.storeKey, types.KeyTakerFeeCommunityPoolProtoRevArray, denom, increasedAmt)
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

// GetLegacyTakerFeeTrackerForStakers is the legacy getter, to be used in the v22 upgrade handler and removed in v23.
func (k Keeper) GetLegacyTakerFeeTrackerForStakers(ctx sdk.Context) (currentTakerFeeForStakers sdk.Coins) {
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

// GetLegacyTakerFeeTrackerForCommunityPool is the legacy getter, to be used in the v22 upgrade handler and removed in v23.
func (k Keeper) GetLegacyTakerFeeTrackerForCommunityPool(ctx sdk.Context) (currentTakerFeeForCommunityPool sdk.Coins) {
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
