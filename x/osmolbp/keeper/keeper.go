package keeper

import (
	"fmt"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/osmosis-labs/osmosis/x/osmolbp/api"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Keeper struct {
	storeKey storetypes.StoreKey
	cdc      codec.BinaryCodec
	bank     BankKeeper
	paramSpace paramtypes.Subspace
}

// NewKeeper constructs a new osmolbp Keeper
func NewKeeper(storeKey storetypes.StoreKey, cdc codec.BinaryCodec, bank BankKeeper) Keeper {
	return Keeper{
		storeKey: storeKey,
		cdc:      cdc,
		bank:     bank,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", api.ModuleName))
}
