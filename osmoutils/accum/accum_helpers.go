package accum

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmoutils"
)

// Creates a new position at accumulator's current value with a specific number of shares and unclaimed rewards
func createNewPosition(accum AccumulatorObject, index string, numShareUnits sdk.Dec, unclaimedRewards sdk.DecCoins, options *Options) {
	position := Record{
		NumShares:        numShareUnits,
		InitAccumValue:   accum.value,
		UnclaimedRewards: unclaimedRewards,
		Options:          options,
	}
	osmoutils.MustSet(accum.store, formatPositionPrefixKey(accum.name, index), &position)
}

// Gets addr's current position from store
func getPosition(accum AccumulatorObject, name string) (Record, error) {
	position := Record{}
	found, err := osmoutils.Get(accum.store, formatPositionPrefixKey(accum.name, name), &position)
	if err != nil {
		return Record{}, err
	}
	if !found {
		return Record{}, NoPositionError{name}
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
