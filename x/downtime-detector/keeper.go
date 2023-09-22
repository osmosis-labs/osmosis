package downtimedetector

import (
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
)

type Keeper struct {
	storeKey storetypes.StoreKey
}

func NewKeeper(storeKey storetypes.StoreKey) *Keeper {
	return &Keeper{storeKey: storeKey}
}
