package downtimedetector

import (
	storetypes "cosmossdk.io/store/types"
)

type Keeper struct {
	storeKey storetypes.StoreKey
}

func NewKeeper(storeKey storetypes.StoreKey) *Keeper {
	return &Keeper{storeKey: storeKey}
}
