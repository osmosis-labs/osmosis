package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/osmosis-labs/osmosis/v12/x/protorev/types"
)

type (
	Keeper struct {
		cdc          codec.BinaryCodec
		storeKey     storetypes.StoreKey
		transientKey *sdk.TransientStoreKey
		memKey       *sdk.MemoryStoreKey
		paramstore   paramtypes.Subspace

		accountKeeper types.AccountKeeper
		bankKeeper    types.BankKeeper
		gammKeeper    types.GAMMKeeper
		epochKeeper   types.EpochKeeper
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey sdk.StoreKey,
	transientKey *sdk.TransientStoreKey,
	memKey *sdk.MemoryStoreKey,
	ps paramtypes.Subspace,
	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
	gammKeeper types.GAMMKeeper,
	epochKeeper types.EpochKeeper,

) Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{

		cdc:           cdc,
		storeKey:      storeKey,
		transientKey:  transientKey,
		memKey:        memKey,
		paramstore:    ps,
		accountKeeper: accountKeeper,
		bankKeeper:    bankKeeper,
		gammKeeper:    gammKeeper,
		epochKeeper:   epochKeeper,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
