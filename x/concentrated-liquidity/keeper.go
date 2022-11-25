package concentrated_liquidity

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	gammtypes "github.com/osmosis-labs/osmosis/v13/x/gamm/types"
)

type Keeper struct {
	storeKey sdk.StoreKey
	cdc      codec.BinaryCodec

	// keepers
	bankKeeper gammtypes.BankKeeper
}

func NewKeeper(cdc codec.BinaryCodec, storeKey sdk.StoreKey, bankKeeper gammtypes.BankKeeper) *Keeper {
	return &Keeper{
		storeKey:   storeKey,
		cdc:        cdc,
		bankKeeper: bankKeeper,
	}
}

// TODO: spec, tests, implementation
func (k Keeper) InitializePool(ctx sdk.Context, pool gammtypes.PoolI, creatorAddress sdk.AccAddress) error {
	panic("not implemented")
}
