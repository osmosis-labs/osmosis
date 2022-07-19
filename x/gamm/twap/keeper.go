package twap

import sdk "github.com/cosmos/cosmos-sdk/types"

type twapkeeper struct {
	storeKey     sdk.StoreKey
	transientKey sdk.TransientStoreKey
}
