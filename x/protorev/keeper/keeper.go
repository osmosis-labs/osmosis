package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/osmosis-labs/osmosis/v15/x/protorev/types"
)

type (
	Keeper struct {
		cdc        codec.BinaryCodec
		storeKey   storetypes.StoreKey
		paramstore paramtypes.Subspace

		accountKeeper     types.AccountKeeper
		bankKeeper        types.BankKeeper
		gammKeeper        types.GAMMKeeper
		epochKeeper       types.EpochKeeper
		poolmanagerKeeper types.PoolManagerKeeper
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey sdk.StoreKey,
	ps paramtypes.Subspace,
	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
	gammKeeper types.GAMMKeeper,
	epochKeeper types.EpochKeeper,
	poolmanagerKeeper types.PoolManagerKeeper,
) Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		cdc:               cdc,
		storeKey:          storeKey,
		paramstore:        ps,
		accountKeeper:     accountKeeper,
		bankKeeper:        bankKeeper,
		gammKeeper:        gammKeeper,
		epochKeeper:       epochKeeper,
		poolmanagerKeeper: poolmanagerKeeper,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
