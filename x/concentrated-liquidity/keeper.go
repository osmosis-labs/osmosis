package concentrated_liquidity

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Keeper struct {
	storeKey     sdk.StoreKey
	transientKey *sdk.TransientStoreKey
}

func NewKeeper(storeKey sdk.StoreKey, transientKey *sdk.TransientStoreKey) *Keeper {
	return &Keeper{storeKey: storeKey, transientKey: transientKey}
}
