package keeper

import (
	"fmt"

	"github.com/osmosis-labs/osmosis/x/osmolbp"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Keeper struct {
	storeKey storetypes.StoreKey
	cdc      codec.BinaryMarshaler // .BinaryCodec
	bank     BankKeeper
}

// NewKeeper constructs a message authorization Keeper
func NewKeeper(storeKey storetypes.StoreKey, cdc codec.BinaryMarshaler, bank BankKeeper) Keeper {
	return Keeper{
		storeKey: storeKey,
		cdc:      cdc,
		bank:     bank,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", osmolbp.ModuleName))
}
