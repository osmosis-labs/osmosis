package accum

import (
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Creates an accumulator object for testing purposes
func CreateRawAccumObject(store store.KVStore, name string, value sdk.Dec) AccumulatorObject {
	return AccumulatorObject{
		store: store,
		name: name,
		value: value,
	}
}