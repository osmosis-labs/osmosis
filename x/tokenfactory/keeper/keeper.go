package keeper

import (
	"fmt"

	"cosmossdk.io/log"

	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v27/x/tokenfactory/types"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

type (
	Keeper struct {
		storeKey    storetypes.StoreKey
		permAddrs   map[string]authtypes.PermissionsForAddress
		permAddrMap map[string]bool

		paramSpace paramtypes.Subspace

		accountKeeper  types.AccountKeeper
		bankKeeper     types.BankKeeper
		contractKeeper types.ContractKeeper

		communityPoolKeeper types.CommunityPoolKeeper
	}
)

// NewKeeper returns a new instance of the x/tokenfactory keeper
func NewKeeper(
	storeKey storetypes.StoreKey,
	paramSpace paramtypes.Subspace,
	maccPerms map[string][]string,
	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
	communityPoolKeeper types.CommunityPoolKeeper,
) Keeper {
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	permAddrs := make(map[string]authtypes.PermissionsForAddress)
	permAddrMap := make(map[string]bool)
	for name, perms := range maccPerms {
		permsForAddr := authtypes.NewPermissionsForAddress(name, perms)
		permAddrs[name] = permsForAddr
		permAddrMap[permsForAddr.GetAddress().String()] = true
	}

	return Keeper{
		storeKey:   storeKey,
		paramSpace: paramSpace,
		permAddrs:  permAddrs,

		accountKeeper:       accountKeeper,
		bankKeeper:          bankKeeper,
		communityPoolKeeper: communityPoolKeeper,
	}
}

// Logger returns a logger for the x/tokenfactory module
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetDenomPrefixStore returns the substore for a specific denom
func (k Keeper) GetDenomPrefixStore(ctx sdk.Context, denom string) storetypes.KVStore {
	store := ctx.KVStore(k.storeKey)
	return prefix.NewStore(store, types.GetDenomPrefixStore(denom))
}

// GetCreatorPrefixStore returns the substore for a specific creator address
func (k Keeper) GetCreatorPrefixStore(ctx sdk.Context, creator string) storetypes.KVStore {
	store := ctx.KVStore(k.storeKey)
	return prefix.NewStore(store, types.GetCreatorPrefix(creator))
}

// GetCreatorsPrefixStore returns the substore that contains a list of creators
func (k Keeper) GetCreatorsPrefixStore(ctx sdk.Context) storetypes.KVStore {
	store := ctx.KVStore(k.storeKey)
	return prefix.NewStore(store, types.GetCreatorsPrefix())
}

// Set the wasm keeper.
func (k *Keeper) SetContractKeeper(contractKeeper types.ContractKeeper) {
	k.contractKeeper = contractKeeper
}

// CreateModuleAccount creates a module account with minting and burning capabilities
// This account isn't intended to store any coins,
// it purely mints and burns them on behalf of the admin of respective denoms,
// and sends to the relevant address.
func (k Keeper) CreateModuleAccount(ctx sdk.Context) {
	k.accountKeeper.GetModuleAccount(ctx, types.ModuleName)
}
