package accum

import (
	"errors"
	"fmt"

	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmoutils"
)

// We keep this object as a way to interface with the methods, even though
// only the Accumulator inside is stored in state
type AccumulatorObject struct {
	// Store where accumulator is stored
	store store.KVStore

	// Accumulator's name (pulled from AccumulatorContent)
	name string

	// Accumulator's current value (pulled from AccumulatorContent)
	value sdk.DecCoins
}

type PositionOptions struct{}

// Makes a new accumulator at store/accum/{accumName}
// returns an error if already exists / theres some overlapping keys
func MakeAccumulator(accumStore store.KVStore, accumName string) error {
	if accumStore.Has(formatAccumPrefixKey(accumName)) {
		return errors.New("Accumulator with given name already exists in store")
	}

	// New accumulator values start out at zero
	// TODO: consider whether this should be a parameter instead of always zero
	initAccumValue := sdk.NewDecCoins()

	newAccum := AccumulatorObject{accumStore, accumName, initAccumValue}

	// Stores accumulator in state
	setAccumulator(newAccum, initAccumValue)

	return nil
}

// Gets the current value of the accumulator corresponding to accumName in accumStore
func GetAccumulator(accumStore store.KVStore, accumName string) (AccumulatorObject, error) {
	accumContent := AccumulatorContent{}
	found, err := osmoutils.Get(accumStore, formatAccumPrefixKey(accumName), &accumContent)
	if err != nil {
		return AccumulatorObject{}, err
	}
	if !found {
		return AccumulatorObject{}, errors.New(fmt.Sprintf("Accumulator name %s does not exist in store", accumName))
	}

	accum := AccumulatorObject{accumStore, accumName, accumContent.AccumValue}

	return accum, nil
}

func setAccumulator(accum AccumulatorObject, amt sdk.DecCoins) error {
	// TODO: consider removing name as as a field from AccumulatorContent (doesn't need to be stored in state)
	newAccum := AccumulatorContent{amt}

	osmoutils.MustSet(accum.store, formatAccumPrefixKey(accum.name), &newAccum)

	return nil
}

// TODO: consider making this increment the accumulator's value instead of overwriting it
func (accum AccumulatorObject) UpdateAccumulator(amt sdk.DecCoins) {
	setAccumulator(accum, amt)
}

// NewPosition creates a new position for the given address, with the given number of units
// It takes a snapshot of the current accumulator value, and sets the position's initial value to that
// The position is initialized with empty unclaimed rewards
func (accum AccumulatorObject) NewPosition(addr sdk.AccAddress, numShareUnits sdk.Dec, options PositionOptions) {
	position := Record{
		NumShares:        numShareUnits,
		InitAccumValue:   accum.value,
		UnclaimedRewards: sdk.NewDecCoins(),
	}
	osmoutils.MustSet(accum.store, formatPositionPrefixKey(addr.String()), &position)
}

// func (accum AccumulatorObject) AddToPosition(addr, num_units) error

// func (accum AccumulatorObject) RemovePosition(addr, num_units) error

// func (accum AccumulatorObject) GetPositionSize(addr) (num_units, error)

// ClaimRewards claims the rewards for the given address, and returns the amount of rewards claimed.
// Upon claiming the rewards, the position at the current address is reset to have no
// unclaimed rewards and the accumulator updates.
func (accum AccumulatorObject) ClaimRewards(addr sdk.AccAddress) (sdk.DecCoins, error) {
	position := Record{}
	found, err := osmoutils.Get(accum.store, formatPositionPrefixKey(addr.String()), &position)
	if err != nil {
		return sdk.DecCoins{}, err
	}
	if !found {
		return sdk.DecCoins{}, fmt.Errorf("no position found for address (%s)", addr)
	}

	totalRewards := position.UnclaimedRewards

	accumulatorRewads := accum.value.Sub(position.InitAccumValue).MulDec(position.NumShares)
	totalRewards = totalRewards.Add(accumulatorRewads...)

	// Create a completely new position, with no rewards
	// TODO: decide how to propagate the knowledge of position options.
	accum.NewPosition(addr, position.NumShares, PositionOptions{})

	return totalRewards, nil
}
