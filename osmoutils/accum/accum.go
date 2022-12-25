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
	value sdk.Dec
}

// Makes a new accumulator at store/accum/{accumName}
// returns an error if already exists / theres some overlapping keys
func MakeAccumulator(accumStore store.KVStore, accumName string) error {
	keybz := []byte(accumName)
	if accumStore.Has(keybz) {
		return errors.New("Accumulator with given name already exists in store")
	}

	// New accumulator values start out at zero
	// TODO: consider whether this should be a parameter instead of always zero
	initAccumValue := sdk.ZeroDec()

	newAccum := AccumulatorObject{accumStore, accumName, initAccumValue}

	// Stores accumulator in state
	setAccumulator(newAccum, initAccumValue)

	return nil
}
 
// Gets the current value of the accumulator corresponding to accumName in accumStore
func GetAccumulator(accumStore store.KVStore, accumName string) (AccumulatorObject, error) {
	keybz := []byte(accumName)

	accumContent := AccumulatorContent{}
	found, err := osmoutils.Get(accumStore, keybz, &accumContent)
	if err != nil {
		return AccumulatorObject{}, err
	}
	if !found {
		return AccumulatorObject{}, errors.New(fmt.Sprintf("Accumulator name %s does not exist in store", accumName))
	}

	accum := AccumulatorObject{accumStore, accumContent.AccumName, accumContent.AccumValue}

	return accum, nil
}

func setAccumulator(accum AccumulatorObject, amt sdk.Dec) error {
	keybz := []byte(accum.name)

	// TODO: consider removing name as as a field from AccumulatorContent (doesn't need to be stored in state)
	newAccum := AccumulatorContent{accum.name, amt}

	osmoutils.MustSet(accum.store, keybz, &newAccum)

	return nil
}

// TODO: consider making this increment the accumulator's value instead of overwriting it
func (accum AccumulatorObject) UpdateAccumulator(amt sdk.Dec) {
	setAccumulator(accum, amt)
}

// func (accum AccumulatorObject) NewPosition(addr, num_units, positionOptions) error

// func (accum AccumulatorObject) AddToPosition(addr, num_units) error

// func (accum AccumulatorObject) RemovePosition(addr, num_units) error

// func (accum AccumulatorObject) GetPositionSize(addr) (num_units, error)

// func (accum AccumulatorObject) ClaimRewards(sendKeeper, addr) (amt AccumType, error)
