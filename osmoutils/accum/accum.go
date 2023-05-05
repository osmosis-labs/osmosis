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

	// Accumulator's total shares across all positions
	totalShares sdk.Dec
}

// Makes a new accumulator at store/accum/{accumName}
// Returns error if already exists / theres some overlapping keys
func MakeAccumulator(accumStore store.KVStore, accumName string) error {
	if accumStore.Has(formatAccumPrefixKey(accumName)) {
		return errors.New("Accumulator with given name already exists in store")
	}

	// New accumulator values start out at zero
	// TODO: consider whether this should be a parameter instead of always zero
	initAccumValue := sdk.NewDecCoins()
	initTotalShares := sdk.ZeroDec()

	newAccum := &AccumulatorObject{accumStore, accumName, initAccumValue, initTotalShares}

	// Stores accumulator in state
	setAccumulator(newAccum, initAccumValue, initTotalShares)

	return nil
}

// Makes a new accumulator at store/accum/{accumName}
// Returns error if already exists / theres some overlapping keys
func MakeAccumulatorWithValueAndShare(accumStore store.KVStore, accumName string, accumValue sdk.DecCoins, totalShares sdk.Dec) error {
	if accumStore.Has(formatAccumPrefixKey(accumName)) {
		return errors.New("Accumulator with given name already exists in store")
	}

	newAccum := AccumulatorObject{accumStore, accumName, accumValue, totalShares}

	// Stores accumulator in state
	setAccumulator(&newAccum, accumValue, totalShares)

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
		return AccumulatorObject{}, AccumDoesNotExistError{AccumName: accumName}
	}

	accum := AccumulatorObject{accumStore, accumName, accumContent.AccumValue, accumContent.TotalShares}

	return accum, nil
}

// MustGetPosition returns the position associated with the given address. No errors in position retrieval are allowed.
func (accum AccumulatorObject) MustGetPosition(name string) Record {
	position := Record{}
	osmoutils.MustGet(accum.store, FormatPositionPrefixKey(accum.name, name), &position)
	return position
}

// GetPosition returns the position associated with the given address. If the position does not exist, returns an error.
func (accum AccumulatorObject) GetPosition(name string) (Record, error) {
	position := Record{}
	found, err := osmoutils.Get(accum.store, FormatPositionPrefixKey(accum.name, name), &position)
	if err != nil {
		return Record{}, err
	}

	if !found {
		return Record{}, NoPositionError{Name: name}
	}
	return position, nil
}

func setAccumulator(accum *AccumulatorObject, value sdk.DecCoins, shares sdk.Dec) {
	newAccum := AccumulatorContent{value, shares}
	osmoutils.MustSet(accum.store, formatAccumPrefixKey(accum.name), &newAccum)
}

// AddToAccumulator updates the accumulator's value by amt.
// It does so by increasing the value of the accumulator by
// the given amount. Persists to store. Mutates the receiver.
func (accum *AccumulatorObject) AddToAccumulator(amt sdk.DecCoins) {
	accum.value = accum.value.Add(amt...)
	setAccumulator(accum, accum.value, accum.totalShares)
}

// NewPosition creates a new position for the given name, with the given number of share units.
// The name can be an owner's address, or any other unique identifier for a position.
// It takes a snapshot of the current accumulator value, and sets the position's initial value to that.
// The position is initialized with empty unclaimed rewards
// If there is an existing position for the given address, it is overwritten.
func (accum *AccumulatorObject) NewPosition(name string, numShareUnits sdk.Dec, options *Options) error {
	return accum.NewPositionCustomAcc(name, numShareUnits, accum.value, options)
}

// NewPositionCustomAcc creates a new position for the given name, with the given number of share units.
// The name can be an owner's address, or any other unique identifier for a position.
// It sets the position's accumulator to the given value of customAccumulatorValue.
// All custom accumulator values must be non-negative.
// The position is initialized with empty unclaimed rewards
// If there is an existing position for the given address, it is overwritten.
func (accum *AccumulatorObject) NewPositionCustomAcc(name string, numShareUnits sdk.Dec, customAccumulatorValue sdk.DecCoins, options *Options) error {
	if customAccumulatorValue.IsAnyNegative() {
		return NegativeCustomAccError{customAccumulatorValue}
	}

	if err := options.validate(); err != nil {
		return err
	}

	initOrUpdatePosition(*accum, customAccumulatorValue, name, numShareUnits, sdk.NewDecCoins(), options)

	// Update total shares in accum (re-fetch accum from state to ensure it's up to date)
	updatedAccum, err := GetAccumulator(accum.store, accum.name)
	if err != nil {
		return err
	}
	accum.totalShares = updatedAccum.totalShares.Add(numShareUnits)
	setAccumulator(accum, accum.value, accum.totalShares)

	return nil
}

// AddToPosition adds newShares of shares to an existing position with the given name.
// This is functionally equivalent to claiming rewards, closing down the position, and
// opening a fresh one with the new number of shares. We can represent this behavior by
// claiming rewards and moving up the accumulator start value to its current value.
//
// An alternative approach is to simply generate an additional position every time an
// address adds to its position. We do not pursue this path because we want to ensure
// that withdrawal and claiming functions remain constant time and do not scale with the
// number of times a user has added to their position.
//
// Returns nil on success. Returns error when:
// - newShares are negative or zero.
// - there is no existing position at the given address
// - other internal or database error occurs.
func (accum *AccumulatorObject) AddToPosition(name string, newShares sdk.Dec) error {
	return accum.AddToPositionCustomAcc(name, newShares, accum.value)
}

// AddToPositionCustomAcc adds newShares of shares to an existing position with the given name.
// This is functionally equivalent to claiming rewards, closing down the position, and
// opening a fresh one with the new number of shares. We can represent this behavior by
// claiming rewards. The accumulator of the new position is set to given customAccumulatorValue.
// All custom accumulator values must be non-negative. They must also be a superset of the
// old accumulator value associated with the position.
//
// An alternative approach is to simply generate an additional position every time an
// address adds to its position. We do not pursue this path because we want to ensure
// that withdrawal and claiming functions remain constant time and do not scale with the
// number of times a user has added to their position.
//
// Returns nil on success. Returns error when:
// - newShares are negative or zero.
// - there is no existing position at the given address
// - other internal or database error occurs.
func (accum *AccumulatorObject) AddToPositionCustomAcc(name string, newShares sdk.Dec, customAccumulatorValue sdk.DecCoins) error {
	if !newShares.IsPositive() {
		return errors.New("Attempted to add zero or negative number of shares to a position")
	}

	// Get addr's current position
	position, err := GetPosition(*accum, name)
	if err != nil {
		return err
	}

	// Save current number of shares and unclaimed rewards
	unclaimedRewards := GetTotalRewards(*accum, position)
	oldNumShares, err := accum.GetPositionSize(name)
	if err != nil {
		return err
	}

	// Update user's position with new number of shares while moving its unaccrued rewards
	// into UnclaimedRewards. Starting accumulator value is moved up to accum'scurrent value
	initOrUpdatePosition(*accum, customAccumulatorValue, name, oldNumShares.Add(newShares), unclaimedRewards, position.Options)

	// Update total shares in accum (re-fetch accum from state to ensure it's up to date)
	updatedAccum, err := GetAccumulator(accum.store, accum.name)
	if err != nil {
		return err
	}
	accum.totalShares = updatedAccum.totalShares.Add(newShares)
	setAccumulator(accum, accum.value, accum.totalShares)

	return nil
}

// RemovePosition removes the specified number of shares from a position. Specifically, it claims
// the unclaimed and newly accrued rewards and returns them alongside the redeemed shares. Then, it
// overwrites the position record with the updated number of shares. Since it accrues rewards, it
// also moves up the position's accumulator value to the current accum val.
func (accum *AccumulatorObject) RemoveFromPosition(name string, numSharesToRemove sdk.Dec) error {
	return accum.RemoveFromPositionCustomAcc(name, numSharesToRemove, accum.value)
}

// RemovePositionCustomAcc removes the specified number of shares from a position. Specifically, it claims
// the unclaimed and newly accrued rewards and returns them alongside the redeemed shares. Then, it
// overwrites the position record with the updated number of shares. Since it accrues rewards, it
// also resets the position's accumulator value to the given customAccumulatorValue.
// All custom accumulator values must be non-negative. They must also be a superset of the
// old accumulator value associated with the position.
func (accum *AccumulatorObject) RemoveFromPositionCustomAcc(name string, numSharesToRemove sdk.Dec, customAccumulatorValue sdk.DecCoins) error {
	// Cannot remove zero or negative shares
	if numSharesToRemove.LTE(sdk.ZeroDec()) {
		return fmt.Errorf("Attempted to remove no/negative shares (%s)", numSharesToRemove)
	}

	// Get addr's current position
	position, err := GetPosition(*accum, name)
	if err != nil {
		return err
	}

	// Ensure not removing more shares than exist
	if numSharesToRemove.GT(position.NumShares) {
		return fmt.Errorf("Attempted to remove more shares (%s) than exist in the position (%s)", numSharesToRemove, position.NumShares)
	}

	// Save current number of shares and unclaimed rewards
	unclaimedRewards := GetTotalRewards(*accum, position)
	oldNumShares, err := accum.GetPositionSize(name)
	if err != nil {
		return err
	}

	// Update user's position with new number of shares
	initOrUpdatePosition(*accum, customAccumulatorValue, name, oldNumShares.Sub(numSharesToRemove), unclaimedRewards, position.Options)

	updatedAccum, err := GetAccumulator(accum.store, accum.name)
	if err != nil {
		return err
	}
	accum.totalShares = updatedAccum.totalShares.Sub(numSharesToRemove)
	setAccumulator(accum, accum.value, accum.totalShares)

	return nil
}

// UpdatePosition updates the position with the given name by adding or removing
// the given number of shares. If numShares is positive, it is equivalent to calling
// AddToPosition. If numShares is negative, it is equivalent to calling RemoveFromPosition.
// Also, it moves up the position's accumulator value to the current accum value.
// Fails with error if numShares is zero. Returns nil on success.
func (accum *AccumulatorObject) UpdatePosition(name string, numShares sdk.Dec) error {
	return accum.UpdatePositionCustomAcc(name, numShares, accum.value)
}

// UpdatePositionCustomAcc updates the position with the given name by adding or removing
// the given number of shares. If numShares is positive, it is equivalent to calling
// AddToPositionCustomAcc. If numShares is negative, it is equivalent to calling RemoveFromPositionCustomAcc.
// Fails with error if numShares is zero. Returns nil on success.
// It also resets the position's accumulator value to the given customAccumulatorValue.
// All custom accumulator values must be non-negative. They must also be a superset of the
// old accumulator value associated with the position.
func (accum *AccumulatorObject) UpdatePositionCustomAcc(name string, numShares sdk.Dec, customAccumulatorValue sdk.DecCoins) error {
	if numShares.Equal(sdk.ZeroDec()) {
		return ZeroSharesError
	}

	if numShares.IsNegative() {
		return accum.RemoveFromPositionCustomAcc(name, numShares.Neg(), customAccumulatorValue)
	}

	return accum.AddToPositionCustomAcc(name, numShares, customAccumulatorValue)
}

// SetPositionCustomAcc sets the position's accumulator to the given value.
// Does not update shares or attempt to claim rewards.
// The new accumulator value must be greater than or equal to the old accumulator value.
// Returns nil on success, error otherwise.
func (accum *AccumulatorObject) SetPositionCustomAcc(name string, customAccumulatorValue sdk.DecCoins) error {
	// Get addr's current position
	position, err := GetPosition(*accum, name)
	if err != nil {
		return err
	}

	// Update the user's position with the new accumulator value. The unclaimed rewards, options, and
	// the number of shares stays the same as in the original position.
	initOrUpdatePosition(*accum, customAccumulatorValue, name, position.NumShares, position.UnclaimedRewards, position.Options)

	return nil
}

// DeletePosition claims rewards and deletes the position from the accumulator state.
// Prior to deletion, claims rewards and returns them. Decrements total accumulator share
// counter by the number of shares in the position tracker.
// Returns error if:
// - fails to fetch a position
// - fails to claim rewards
// - fails to retrieve total accumulator shares
func (accum *AccumulatorObject) DeletePosition(positionName string) (sdk.DecCoins, error) {
	position, err := accum.GetPosition(positionName)
	if err != nil {
		return sdk.DecCoins{}, err
	}

	remainingRewards, dust, err := accum.ClaimRewards(positionName)
	if err != nil {
		return sdk.DecCoins{}, err
	}

	accum.store.Delete(FormatPositionPrefixKey(accum.name, positionName))

	totalShares, err := accum.GetTotalShares()
	if err != nil {
		return sdk.DecCoins{}, err
	}
	accum.totalShares = totalShares.Sub(position.NumShares)
	setAccumulator(accum, accum.value, accum.totalShares)

	return sdk.NewDecCoinsFromCoins(remainingRewards...).Add(dust...), nil
}

// deletePosition deletes the position with the given name from state.
func (accum AccumulatorObject) deletePosition(positionName string) {
	accum.store.Delete(FormatPositionPrefixKey(accum.name, positionName))
}

// GetPositionSize returns the number of shares the position corresponding to `addr`
// in accumulator `accum` has, or an error if no position exists.
func (accum AccumulatorObject) GetPositionSize(name string) (sdk.Dec, error) {
	position, err := GetPosition(accum, name)
	if err != nil {
		return sdk.Dec{}, err
	}

	return position.NumShares, nil
}

// HasPosition returns true if a position with the given name exists,
// false otherwise. Returns error if internal database error occurs.
func (accum AccumulatorObject) HasPosition(name string) (bool, error) {
	_, err := GetPosition(accum, name)

	if err != nil {
		isNoPositionError := errors.Is(err, NoPositionError{Name: name})
		if isNoPositionError {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// GetValue returns the current value of the accumulator.
func (accum AccumulatorObject) GetName() string {
	return accum.name
}

// GetValue returns the current value of the accumulator.
func (accum AccumulatorObject) GetValue() sdk.DecCoins {
	return accum.value
}

// ClaimRewards claims the rewards for the given address, and returns the amount of rewards claimed.
// Upon claiming the rewards, the position at the current address is reset to have no
// unclaimed rewards. The position's accumulator is also set to the current accumulator value.
//
// Returns error if
// - no position exists for the given address
// - any database errors occur.
func (accum AccumulatorObject) ClaimRewards(positionName string) (sdk.Coins, sdk.DecCoins, error) {
	position, err := GetPosition(accum, positionName)
	if err != nil {
		return sdk.Coins{}, sdk.DecCoins{}, NoPositionError{positionName}
	}

	totalRewards := GetTotalRewards(accum, position)

	// Return the integer coins to the user
	// The remaining change is thrown away.
	// This is acceptable because we round in favor of the protocol.
	truncatedRewards, dust := totalRewards.TruncateDecimal()

	if position.NumShares.Equal(sdk.ZeroDec()) {
		// remove the position from state entirely if numShares = zero
		accum.deletePosition(positionName)
	} else {
		// else, update the position with no rewards
		initOrUpdatePosition(accum, accum.value, positionName, position.NumShares, sdk.NewDecCoins(), position.Options)
	}

	return truncatedRewards, dust, nil
}

// GetTotalShares returns the total number of shares in the accumulator
func (accum AccumulatorObject) GetTotalShares() (sdk.Dec, error) {
	accum, err := GetAccumulator(accum.store, accum.name)
	return accum.totalShares, err
}

// AddToUnclaimedRewards adds the given amount of rewards to the unclaimed rewards
// for the given position. Returns error if no position exists for the given address.
// Returns error if any database errors occur.
func (accum AccumulatorObject) AddToUnclaimedRewards(positionName string, rewards sdk.DecCoins) error {
	position, err := GetPosition(accum, positionName)
	if err != nil {
		return err
	}

	if rewards.IsAnyNegative() {
		return NegativeRewardsAdditionError{PositionName: positionName, AccumName: accum.name}
	}

	// Update the user's position with the new unclaimed rewards. The accumulator, options, and
	// the number of shares stays the same as in the original position.
	initOrUpdatePosition(accum, position.InitAccumValue, positionName, position.NumShares, position.UnclaimedRewards.Add(rewards...), position.Options)

	return nil
}
