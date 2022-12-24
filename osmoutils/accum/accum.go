package accum

import (
	"errors"
	"fmt"

	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"
)

// We keep this object as a way to interface with the methods, even though
// only the Accumulator inside is stored in state
type AccumulatorObject struct {
	// Store where accumulator is stored
	Store store.KVStore

	// Accumulator's name (pulled from AccumulatorContent)
	Name string

	// Accumulator's current value (pulled from AccumulatorContent)
	Value sdk.Dec
}

// Makes a new accumulator at store/accum/{accumulator_name}
// returns an error if already exists / theres some overlapping keys
func MakeAccumulator(accum_store store.KVStore, accum_name string) error {
	keybz := []byte(accum_name)
	if accum_store.Has(keybz) {
		return errors.New("Accumulator with given name already exists in store")
	}

	// New accumulator values start out at zero
	// TODO: consider whether this should be a parameter instead of always zero
	init_accum_value := sdk.ZeroDec()

	var new_accum AccumulatorObject
	new_accum.Store = accum_store
	new_accum.Name = accum_name
	new_accum.Value = init_accum_value

	// Stores accumulator in state
	setAccumulator(accum_store, new_accum, init_accum_value)

	return nil
}
 
// Gets the current value of the accumulator corresponding to accum_name in accum_store
func GetAccumulator(accum_store store.KVStore, accum_name string) (AccumulatorObject, error) {
	keybz := []byte(accum_name)
	if !accum_store.Has(keybz) {
		return AccumulatorObject{}, errors.New(fmt.Sprintf("Accumulator name %s does not exist in store", accum_name))
	}

	var accum_content AccumulatorContent
	bz := accum_store.Get(keybz)
	err := proto.Unmarshal(bz, &accum_content)
	if err != nil {
		return AccumulatorObject{}, err
	}

	accum := AccumulatorObject{accum_store, accum_content.AccumName, accum_content.AccumValue}

	return accum, nil
}

func setAccumulator(accum_store store.KVStore, accum AccumulatorObject, amt sdk.Dec) error {
	keybz := []byte(accum.Name)

	// TODO: consider removing name as as a field from AccumulatorContent (doesn't need to be stored in state)
	var new_accum AccumulatorContent
	new_accum.AccumName = accum.Name
	new_accum.AccumValue = amt

	bz, err := proto.Marshal(&new_accum)
	if err != nil {
		return err
	}

	accum_store.Set(keybz, bz)

	return nil
}

// TODO: consider making this increment the accumulator's value instead of overwriting it
func (accum AccumulatorObject) UpdateAccumulator(amt sdk.Dec) {
	setAccumulator(accum.Store, accum, amt)
}

// func (accum AccumulatorObject) NewPosition(addr, num_units, positionOptions) error

// func (accum AccumulatorObject) AddToPosition(addr, num_units) error

// func (accum AccumulatorObject) RemovePosition(addr, num_units) error

// func (accum AccumulatorObject) GetPositionSize(addr) (num_units, error)

// func (accum AccumulatorObject) ClaimRewards(sendKeeper, addr) (amt AccumType, error)
