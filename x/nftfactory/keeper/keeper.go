package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/gogo/protobuf/proto"
	"github.com/osmosis-labs/osmosis/v12/x/nftfactory/types"
	"github.com/tendermint/tendermint/libs/log"
)

type (
	Keeper struct {
		storeKey      sdk.StoreKey
		paramSpace    paramtypes.Subspace
		accountKeeper types.AccountKeeper
		bankKeeper    types.BankKeeper
	}
)

// NewKeeper returns a new instance of the x/tokenfactory keeper
func NewKeeper(
	storeKey sdk.StoreKey,
	paramSpace paramtypes.Subspace,
	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
) Keeper {
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		storeKey:   storeKey,
		paramSpace: paramSpace,

		accountKeeper: accountKeeper,
		bankKeeper:    bankKeeper,
	}
}

// Logger returns a logger for the x/tokenfactory module
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// HasDenomID returns whether the specified denom ID exists
func (k Keeper) HasDenomID(ctx sdk.Context, id string) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(types.KeyDenomID(id))
}

// SetDenom is responsible for saving the definition of denom
func (k Keeper) SetDenom(ctx sdk.Context, denom types.Denom) error {
	if k.HasDenomID(ctx, denom.Id) {
		return fmt.Errorf("denomID %s has already exists", denom.Id)
	}

	store := ctx.KVStore(k.storeKey)
	bz, err := proto.Marshal(&denom)
	if err != nil {
		return err
	}

	store.Set(types.KeyDenomID(denom.Id), bz)
	store.Set(types.KeyDenomName(denom.DenomName), []byte(denom.Id))
	return nil
}
