package accum

import (
	"errors"

	"github.com/gogo/protobuf/proto"

	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmoutils"
)

// GetAllPositions returns all positions associated with the receiver accumulator.
// Returns error if any database errors occur.
// This function is currently used for testing purposes only.
// If there is a need to use this function in production, it
// can be moved to a non-test file.
func (accum AccumulatorObject) GetAllPositions() ([]Record, error) {
	return osmoutils.GatherValuesFromStorePrefix(accum.store, FormatPositionPrefixKey(accum.name, ""), parseRecordFromBz)
}

// Creates an accumulator object for testing purposes
func MakeTestAccumulator(store store.KVStore, name string, value sdk.DecCoins, totalShares sdk.Dec) *AccumulatorObject {
	// We store an accumulator object in state even if unused in tests
	// because position operations still require GetAccumulator to work
	_ = MakeAccumulator(store, name)
	acc := AccumulatorObject{
		store:         store,
		name:          name,
		valuePerShare: value,
		totalShares:   totalShares,
	}
	setAccumulator(&acc, value, totalShares)
	return &acc
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

// WithPosition is a decorator test function to append a position with the given name to the given accumulator.
func WithPosition(accum *AccumulatorObject, name string, position Record) *AccumulatorObject {
	osmoutils.MustSet(accum.store, FormatPositionPrefixKey(accum.name, name), &position)
	return accum
}

// SetValue is a convinience test helper for updatung the value of an accumulator object
// in tests.
func (accum *AccumulatorObject) SetValue(value sdk.DecCoins) {
	accum.valuePerShare = value
}

func (o *Options) Validate() error {
	return o.validate()
}

// WARNING: only used in tests to make sure that receiver is mutated.
// Do not move out of export_test.go and do not use in production code.
func (accum AccumulatorObject) GetTotalShareField() sdk.Dec {
	return accum.totalShares
}

// WARNING: only used in tests to make sure that receiver is mutated.
// Do not move out of export_test.go and do not use in production code.
func (accum AccumulatorObject) GetValueField() sdk.DecCoins {
	return accum.valuePerShare
}

func InitOrUpdatePosition(accum *AccumulatorObject, accumulatorValue sdk.DecCoins, index string, numShareUnits sdk.Dec, unclaimedRewards sdk.DecCoins, options *Options) {
	initOrUpdatePosition(accum, accumulatorValue, index, numShareUnits, unclaimedRewards, options)
}
