package accum

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmoutils"
)

// Creates a new position at accumulator's current value with a specific number of shares and unclaimed rewards
func createNewPosition(accum AccumulatorObject, addr sdk.AccAddress, numShareUnits sdk.Dec, unclaimedRewards sdk.DecCoins, options PositionOptions) {
	position := Record{
		NumShares:        numShareUnits,
		InitAccumValue:   accum.value,
		UnclaimedRewards: unclaimedRewards,
	}
	osmoutils.MustSet(accum.store, formatPositionPrefixKey(addr.String()), &position)
}

// Gets addr's current position from store
func getPosition(store store.KVStore, addr sdk.AccAddress) (Record, error) {
	position := Record{}
	found, err := osmoutils.Get(store, formatPositionPrefixKey(addr.String()), &position)
	if err != nil {
		return Record{}, err
	}
	if !found {
		return Record{}, fmt.Errorf("no position found for address (%s)", addr)
	}

	return position, nil
}

// Gets total unclaimed rewards, including existing and newly accrued unclaimed rewards
func getTotalRewards(accum AccumulatorObject, position Record) sdk.DecCoins {
	totalRewards := position.UnclaimedRewards

	accumulatorRewards := accum.value.Sub(position.InitAccumValue).MulDec(position.NumShares)
	totalRewards = totalRewards.Add(accumulatorRewards...)

	return totalRewards
}