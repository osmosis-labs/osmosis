package concentrated_liquidity

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	swaproutertypes "github.com/osmosis-labs/osmosis/v13/x/swaprouter/types"
)

type Keeper struct {
	storeKey sdk.StoreKey
	cdc      codec.BinaryCodec
}

func NewKeeper(cdc codec.BinaryCodec, storeKey sdk.StoreKey) *Keeper {
	return &Keeper{
		storeKey: storeKey,
		cdc:      cdc,
	}
}

// TODO: spec, tests, implementation
func (k Keeper) InitializePool(ctx sdk.Context, pool swaproutertypes.PoolI, creatorAddress sdk.AccAddress) error {
	panic("not implemented")
}
