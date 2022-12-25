package accum

import (
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Creates an accumulator object for testing purposes
func CreateRawAccumObject(store store.KVStore, name string, value sdk.DecCoins) AccumulatorObject {
	return AccumulatorObject{
		store: store,
		name:  name,
		value: value,
	}
}

func CreateRawPosition(accum AccumulatorObject, addr sdk.AccAddress, numShareUnits sdk.Dec, unclaimedRewards sdk.DecCoins, options PositionOptions) {
	createNewPosition(accum, addr, numShareUnits, unclaimedRewards, options)
}

func GetPosition(store store.KVStore, addr sdk.AccAddress) (Record, error) {
	return getPosition(store, addr)
}

// Gets store from accumulator for testing purposes
func GetStore(accum AccumulatorObject) store.KVStore {
	return accum.store
}