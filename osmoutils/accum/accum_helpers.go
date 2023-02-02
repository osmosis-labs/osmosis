package accum

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmoutils"
)

var (
	minusOne = sdk.NewDec(-1)
)

// Creates a new position or override an existing position
// at accumulator's current value with a specific number of shares and unclaimed rewards
func initOrUpdatePosition(accum AccumulatorObject, accumulatorValue sdk.DecCoins, index string, numShareUnits sdk.Dec, unclaimedRewards sdk.DecCoins, options *Options) {
	position := Record{
		NumShares:        numShareUnits,
		InitAccumValue:   accumulatorValue,
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

	// TODO: add a check that accum.value is greater than position.InitAccumValue
	accumulatorRewards := accum.value.Sub(position.InitAccumValue).MulDec(position.NumShares)
	totalRewards = totalRewards.Add(accumulatorRewards...)

	return totalRewards
}

// validateAccumulatorValue validates the provided accumulator.
// All coins in custom accumulator value must be non-negative.
// Custom accumulator value must be a superset of the old accumulator value.
// Fails if any coin is negative. On success, returns nil.
func validateAccumulatorValue(customAccumulatorValue, oldPositionAccumulatorValue sdk.DecCoins) error {
	if customAccumulatorValue.IsAnyNegative() {
		return NegativeCustomAccError{customAccumulatorValue}
	}
	fmt.Printf("customAccumulatorValue: %v \n", customAccumulatorValue)
	fmt.Printf("oldPositionAccumulatorValue: %v \n", oldPositionAccumulatorValue)
	newValue, IsAnyNegative := customAccumulatorValue.SafeSub(oldPositionAccumulatorValue)
	if IsAnyNegative {
		return NegativeAccDifferenceError{newValue.MulDec(minusOne)}
	}
	return nil
}

// func validateAccumulatorValue(customAccumulatorValue, oldPositionAccumulatorValue sdk.DecCoins) error {
// 	if customAccumulatorValue.IsAnyNegative() {
// 		return NegativeCustomAccError{customAccumulatorValue}
// 	}
// 	fmt.Printf("customAccumulatorValue: %v \n", customAccumulatorValue)
// 	fmt.Printf("oldPositionAccumulatorValue: %v \n", oldPositionAccumulatorValue)

// 	// Normalize coins to panic on negative coins but only when the value is less than or equal to -1
// 	// This is important since we have remainder coins that exist in the accumulator until they reach integer values
// 	customAccumulatorValueCoins := sdk.NormalizeCoins(customAccumulatorValue)
// 	oldPositionAccumulatorValueCoins := sdk.NormalizeCoins(oldPositionAccumulatorValue)
// 	newValue, IsAnyNegative := customAccumulatorValueCoins.SafeSub(oldPositionAccumulatorValueCoins)
// 	if IsAnyNegative {
// 		return NegativeAccDifferenceError{sdk.NewDecCoinsFromCoins(newValue...)}
// 	}
// 	return nil
// }
