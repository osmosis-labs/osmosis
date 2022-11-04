package keeper

import (
	"fmt"

	"github.com/osmosis-labs/osmosis/v12/x/gamm/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

func permContains(perms []string, perm string) bool {
	for _, v := range perms {
		if v == perm {
			return true
		}
	}

	return false
}

type Keeper struct {
	storeKey sdk.StoreKey
	cdc      codec.BinaryCodec

	hooks types.GammHooks

	// keepers
	accountKeeper       types.AccountKeeper
	bankKeeper          types.BankKeeper
	communityPoolKeeper types.CommunityPoolKeeper
}

func NewKeeper(cdc codec.BinaryCodec, storeKey sdk.StoreKey, accountKeeper types.AccountKeeper, bankKeeper types.BankKeeper, communityPoolKeeper types.CommunityPoolKeeper) Keeper {
	// Ensure that the module account are set.
	moduleAddr, perms := accountKeeper.GetModuleAddressAndPermissions(types.ModuleName)
	if moduleAddr == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleName))
	}
	if !permContains(perms, authtypes.Minter) {
		panic(fmt.Sprintf("%s module account should have the minter permission", types.ModuleName))
	}
	if !permContains(perms, authtypes.Burner) {
		panic(fmt.Sprintf("%s module account should have the burner permission", types.ModuleName))
	}
	return Keeper{
		storeKey: storeKey,
		cdc:      cdc,
		// keepers
		accountKeeper:       accountKeeper,
		bankKeeper:          bankKeeper,
		communityPoolKeeper: communityPoolKeeper,
	}
}

// Set the gamm hooks.
func (k *Keeper) SetHooks(gh types.GammHooks) *Keeper {
	if k.hooks != nil {
		panic("cannot set gamm hooks twice")
	}

	k.hooks = gh

	return k
}
