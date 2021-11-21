package keeper

import (
	"fmt"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/x/tokenfactory/types"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

type (
	Keeper struct {
		cdc      codec.Codec
		storeKey sdk.StoreKey

		accountKeeper types.AccountKeeper
		bankKeeper    types.BankKeeper
	}
)

func NewKeeper(
	cdc codec.Codec,
	storeKey sdk.StoreKey,
	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
) *Keeper {
	return &Keeper{
		cdc:      cdc,
		storeKey: storeKey,

		accountKeeper: accountKeeper,
		bankKeeper:    bankKeeper,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) GetDenomPrefixStore(ctx sdk.Context, denom string) sdk.KVStore {
	store := ctx.KVStore(k.storeKey)
	return prefix.NewStore(store, types.GetDenomPrefixStore(denom))
}

func (k Keeper) GetCreatorPrefixStore(ctx sdk.Context, creator string) sdk.KVStore {
	store := ctx.KVStore(k.storeKey)
	return prefix.NewStore(store, types.GetCreatorPrefix(creator))
}

func (k Keeper) GetCreatorsPrefixStore(ctx sdk.Context) sdk.KVStore {
	store := ctx.KVStore(k.storeKey)
	return prefix.NewStore(store, types.GetCreatorsPrefix())
}

// func (k Keeper) GetAdminPrefixStore(ctx sdk.Context, admin string) sdk.KVStore {
// 	store := ctx.KVStore(k.storeKey)
// 	return prefix.NewStore(store, types.GetAdminPrefix(admin))
// }

// CreateModuleAccount creates a module account with minting and burning capabilities
func (k Keeper) CreateModuleAccount(ctx sdk.Context) {
	moduleAcc := authtypes.NewEmptyModuleAccount(types.ModuleName, authtypes.Minter, authtypes.Burner)
	k.accountKeeper.SetModuleAccount(ctx, moduleAcc)
}
