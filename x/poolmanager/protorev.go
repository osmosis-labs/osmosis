package poolmanager

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	gogotypes "github.com/cosmos/gogoproto/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v21/x/poolmanager/types"
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
