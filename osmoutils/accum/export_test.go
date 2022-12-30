package accum

import (
	"errors"

	"github.com/gogo/protobuf/proto"

	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmoutils"
)

// GetPosition returns the position associated with the given address.
// This function is currently used for testing purposes only.
// If there is a need to use this function in production, it
// can be moved to a non-test file.
func (accum AccumulatorObject) GetPosition(addr sdk.AccAddress) Record {
	position := Record{}
	osmoutils.MustGet(accum.store, formatPositionPrefixKey(accum.name, addr.String()), &position)
	return position
}

// GetAllPositions returns all positions associated with the receiver accumulator.
// Returns error if any database errors occur.
// This function is currently used for testing purposes only.
// If there is a need to use this function in production, it
// can be moved to a non-test file.
func (accum AccumulatorObject) GetAllPositions() ([]Record, error) {
	return osmoutils.GatherValuesFromStorePrefix(accum.store, formatPositionPrefixKey(accum.name, ""), parseRecordFromBz)
}

// Creates an accumulator object for testing purposes
func CreateRawAccumObject(store store.KVStore, name string, value sdk.DecCoins) AccumulatorObject {
	return AccumulatorObject{
		store: store,
		name:  name,
		value: value,
	}
}

func CreateRawPosition(accum AccumulatorObject, addr sdk.AccAddress, numShareUnits sdk.Dec, unclaimedRewards sdk.DecCoins, options *Options) {
	createNewPosition(accum, addr, numShareUnits, unclaimedRewards, options)
}

func GetPosition(accum AccumulatorObject, addr sdk.AccAddress) (Record, error) {
	return getPosition(accum, addr)
}

// Gets store from accumulator for testing purposes
func GetStore(accum AccumulatorObject) store.KVStore {
	return accum.store
}

// parseRecordFromBz parses a record from a byte slice.
// Returns error if fails to unmarshal or if the given bytes slice
// is empty.
func parseRecordFromBz(bz []byte) (record Record, err error) {
	if len(bz) == 0 {
		return Record{}, errors.New("record not found")
	}
	err = proto.Unmarshal(bz, &record)
	if err != nil {
		return Record{}, err
	}
	return record, nil
}

// WithPosition is a decorator test function to append a position at given address to the given accumulator.
func WithPosition(accum AccumulatorObject, addr sdk.Address, position Record) AccumulatorObject {
	osmoutils.MustSet(accum.store, formatPositionPrefixKey(accum.name, addr.String()), &position)
	return accum
}

// SetValue is a convinience test helper for updatung the value of an accumulator object
// in tests.
func (accum *AccumulatorObject) SetValue(value sdk.DecCoins) {
	accum.value = value
}

func (o *Options) Validate() error {
	return o.validate()
}
