package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	gogotypes "github.com/gogo/protobuf/types"

	"github.com/osmosis-labs/osmosis/osmoutils"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v20/x/poolmanager/types"
	"github.com/osmosis-labs/osmosis/v20/x/txfees/types"
)

// IncreaseTxFeesTracker gets the current value of the txfees tracker, adds the given amount to it, and sets the new value.
func (k Keeper) IncreaseTxFeesTracker(ctx sdk.Context, txFees sdk.Coin) {
	currentTxFees := k.GetTxFeesTrackerValue(ctx)
	if !txFees.IsZero() {
		newnewTxFeesCoins := currentTxFees.Add(txFees)
		newTxFees := poolmanagertypes.TrackedVolume{
			Amount: newnewTxFeesCoins,
		}
		osmoutils.MustSet(ctx.KVStore(k.storeKey), types.KeyTxFeeProtorevTracker, &newTxFees)
	}
}

func (k Keeper) SetTxFeesTrackerValue(ctx sdk.Context, txFees sdk.Coins) {
	newtxFees := poolmanagertypes.TrackedVolume{
		Amount: txFees,
	}
	osmoutils.MustSet(ctx.KVStore(k.storeKey), types.KeyTxFeeProtorevTracker, &newtxFees)
}

func (k Keeper) GetTxFeesTrackerValue(ctx sdk.Context) (currentTxFees sdk.Coins) {
	var txFees poolmanagertypes.TrackedVolume
	txFeesFound, err := osmoutils.Get(ctx.KVStore(k.storeKey), types.KeyTxFeeProtorevTracker, &txFees)
	if err != nil {
		// We can only encounter an error if a database or serialization errors occurs, so we panic here.
		// Normally this would be handled by `osmoutils.MustGet`, but since we want to specifically use `osmoutils.Get`,
		// we also have to manually panic here.
		panic(err)
	}

	// If no volume was found, we treat the existing volume as 0.
	// While we can technically require volume to exist, we would need to store empty coins in state for each pool (past and present),
	// which is a high storage cost to pay for a weak guardrail.
	currentTxFees = sdk.NewCoins()
	if txFeesFound {
		currentTxFees = txFees.Amount
	}

	return currentTxFees
}

// GetTxFeesTrackerStartHeight gets the height from which we started accounting for txfees.
func (k Keeper) GetTxFeesTrackerStartHeight(ctx sdk.Context) uint64 {
	startHeight := gogotypes.UInt64Value{}
	osmoutils.MustGet(ctx.KVStore(k.storeKey), types.KeyTxFeeProtorevTrackerStartHeight, &startHeight)
	return startHeight.Value
}

// SetTxFeesTrackerStartHeight sets the height from which we started accounting for txfees.
func (k Keeper) SetTxFeesTrackerStartHeight(ctx sdk.Context, startHeight uint64) {
	osmoutils.MustSet(ctx.KVStore(k.storeKey), types.KeyTxFeeProtorevTrackerStartHeight, &gogotypes.UInt64Value{Value: startHeight})
}
