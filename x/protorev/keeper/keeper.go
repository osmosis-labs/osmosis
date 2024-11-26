package keeper

import (
	"fmt"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"

	"cosmossdk.io/log"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/osmosis-labs/osmosis/v27/x/protorev/types"
)

type (
	Keeper struct {
		cdc          codec.BinaryCodec
		storeKey     storetypes.StoreKey
		transientKey *storetypes.TransientStoreKey
		paramstore   paramtypes.Subspace

		accountKeeper               types.AccountKeeper
		bankKeeper                  types.BankKeeper
		gammKeeper                  types.GAMMKeeper
		epochKeeper                 types.EpochKeeper
		poolmanagerKeeper           types.PoolManagerKeeper
		concentratedLiquidityKeeper types.ConcentratedLiquidityKeeper
		distributionKeeper          types.DistributionKeeper
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	transientKey *storetypes.TransientStoreKey,
	ps paramtypes.Subspace,
	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
	gammKeeper types.GAMMKeeper,
	epochKeeper types.EpochKeeper,
	poolmanagerKeeper types.PoolManagerKeeper,
	concentratedLiquidityKeeper types.ConcentratedLiquidityKeeper,
	distributionKeeper types.DistributionKeeper,
) Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		cdc:                         cdc,
		storeKey:                    storeKey,
		transientKey:                transientKey,
		paramstore:                  ps,
		accountKeeper:               accountKeeper,
		bankKeeper:                  bankKeeper,
		gammKeeper:                  gammKeeper,
		epochKeeper:                 epochKeeper,
		poolmanagerKeeper:           poolmanagerKeeper,
		concentratedLiquidityKeeper: concentratedLiquidityKeeper,
		distributionKeeper:          distributionKeeper,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
