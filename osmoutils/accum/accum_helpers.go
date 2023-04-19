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
	osmoutils.MustSet(accum.store, FormatPositionPrefixKey(accum.name, index), &position)
}

// Gets addr's current position from store
func GetPosition(accum AccumulatorObject, name string) (Record, error) {
	position := Record{}
	found, err := osmoutils.Get(accum.store, FormatPositionPrefixKey(accum.name, name), &position)
	if err != nil {
		return Record{}, err
	}
	if !found {
		return Record{}, NoPositionError{name}
	}

	return position, nil
}

// Gets total unclaimed rewards, including existing and newly accrued unclaimed rewards
func GetTotalRewards(accum AccumulatorObject, position Record) (sdk.DecCoins, error) {
	totalRewards := position.UnclaimedRewards
	fmt.Println("totalRewards: ", totalRewards)

	// TODO: add a check that accum.value is greater than position.InitAccumValue
	for _, coin := range accum.value {
		if position.InitAccumValue.AmountOf(coin.Denom).LT(coin.Amount) {
			return nil, fmt.Errorf("custom accumulator value %s is less than the old accumulator value %s", accum.value, position.InitAccumValue)
		}
	}
	accumulatorRewards := accum.value.Sub(position.InitAccumValue).MulDec(position.NumShares)
	fmt.Println("accumulatorRewards: ", accumulatorRewards)
	totalRewards = totalRewards.Add(accumulatorRewards...)
	fmt.Println("totalRewards: ", totalRewards)
	fmt.Println()

	return totalRewards, nil
}

// validateAccumulatorValue validates the provided accumulator.
// All coins in custom accumulator value must be non-negative.
// Custom accumulator value must be a superset of the old accumulator value.
// Fails if any coin is negative. On success, returns nil.
func validateAccumulatorValue(customAccumulatorValue, oldPositionAccumulatorValue sdk.DecCoins) error {
	if customAccumulatorValue.IsAnyNegative() {
		return NegativeCustomAccError{customAccumulatorValue}
	}
	newValue, IsAnyNegative := customAccumulatorValue.SafeSub(oldPositionAccumulatorValue)
	if IsAnyNegative {
		return NegativeAccDifferenceError{newValue.MulDec(minusOne)}
	}
	return nil
}
