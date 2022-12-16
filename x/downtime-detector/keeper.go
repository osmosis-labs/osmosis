package downtimedetector

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Keeper struct {
	storeKey sdk.StoreKey
}

func NewKeeper(storeKey sdk.StoreKey) *Keeper {
	return &Keeper{storeKey: storeKey}
}
