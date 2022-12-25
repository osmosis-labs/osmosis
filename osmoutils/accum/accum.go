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

func setAccumulator(accum AccumulatorObject, amt sdk.DecCoins) {
	newAccum := AccumulatorContent{amt}
	osmoutils.MustSet(accum.store, formatAccumPrefixKey(accum.name), &newAccum)
}

// TODO: consider making this increment the accumulator's value instead of overwriting it
func (accum AccumulatorObject) UpdateAccumulator(amt sdk.DecCoins) {
	setAccumulator(accum, amt)
}

// NewPosition creates a new position for the given address, with the given number of share units
// It takes a snapshot of the current accumulator value, and sets the position's initial value to that
// The position is initialized with empty unclaimed rewards
func (accum AccumulatorObject) NewPosition(addr sdk.AccAddress, numShareUnits sdk.Dec, options PositionOptions) {
	createNewPosition(accum, addr, numShareUnits, sdk.NewDecCoins(), options)
}

// AddToPosition adds newShares of shares to addr's position.
// This is functionally equivalent to claiming rewards, closing down the position, and 
// opening a fresh one with the new number of shares. We can represent this behavior by
// claiming rewards and moving up the accumulator start value to its current value.
//
// An alternative approach is to simply generate an additional position every time an
// address adds to its position. We do not pursue this path because we want to ensure
// that withdrawal and claiming functions remain constant time and do not scale with the
// number of times a user has added to their position.
func (accum AccumulatorObject) AddToPosition(addr sdk.AccAddress, newShares sdk.Dec) error {
	if !newShares.IsPositive() {
		return errors.New("Attempted to add a non-zero and non-negative number of shares to a position")
	}

	// Get addr's current position
	position, err := getPosition(accum.store, addr)
	if err != nil {
		return err
	}

	// Save current number of shares and unclaimed rewards
	unclaimedRewards := getTotalRewards(accum, position)
	oldNumShares, err := accum.GetPositionSize(addr)
	if err != nil {
		return err
	}

	// Update user's position with new number of shares while moving its unaccrued rewards 
	// into UnclaimedRewards. Starting accumulator value is moved up to accum'scurrent value
	// TODO: decide how to propagate the knowledge of position options.
	createNewPosition(accum, addr, oldNumShares.Add(newShares), unclaimedRewards, PositionOptions{})

	return nil
}

// func (accum AccumulatorObject) RemovePosition(addr, num_units) error

func (accum AccumulatorObject) GetPositionSize(addr sdk.AccAddress) (sdk.Dec, error) {
	position, err := getPosition(accum.store, addr)
	if err != nil {
		return sdk.Dec{}, err
	}

	return position.NumShares, nil
}

// ClaimRewards claims the rewards for the given address, and returns the amount of rewards claimed.
// Upon claiming the rewards, the position at the current address is reset to have no
// unclaimed rewards. The position's accumulator is also set to the current accumulator value.
// Returns error if no position exists for the given address. Returns error if any
// database errors occur.
func (accum AccumulatorObject) ClaimRewards(addr sdk.AccAddress) (sdk.DecCoins, error) {
	position, err := getPosition(accum.store, addr)
	if err != nil {
		return sdk.DecCoins{}, err
	}

	totalRewards := getTotalRewards(accum, position)

	// Create a completely new position, with no rewards
	// TODO: decide how to propagate the knowledge of position options.
	accum.NewPosition(addr, position.NumShares, PositionOptions{})

	return totalRewards, nil
}