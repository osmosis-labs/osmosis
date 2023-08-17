package keeper

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/osmosis-labs/osmosis/v17/x/authenticator/types"
)

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

type Keeper struct {
	storeKey   sdk.StoreKey
	cdc        codec.BinaryCodec
	paramSpace paramtypes.Subspace
}

func NewKeeper(cdc codec.BinaryCodec, storeKey sdk.StoreKey, ps paramtypes.Subspace) Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		storeKey:   storeKey,
		cdc:        cdc,
		paramSpace: ps,
	}
}

func (k Keeper) UnmarshalAuthenticator(bz []byte) (types.Authenticator[any], error) {
	var authenticator types.Authenticator[any]
	// ToDo: register interfaces and concrete implementations
	return authenticator, k.cdc.UnmarshalInterface(bz, &authenticator)
}

func (k Keeper) GetAuthenticatorsForAccount(ctx sdk.Context, account sdk.AccAddress) ([]types.Authenticator[any], error) {
	return osmoutils.GatherValuesFromStorePrefix(
		ctx.KVStore(k.storeKey),
		types.KeyAccount(account),
		func(bz []byte) (types.Authenticator[any], error) {
			authenticator, err := k.UnmarshalAuthenticator(bz)
			if err != nil {
				return nil, err
			}

			return authenticator, nil
		})

}

// Add an authenticator to an account
func (k Keeper) AddAuthenticator(ctx sdk.Context, account sdk.AccAddress, authenticator types.Authenticator[any]) error {
	store := ctx.KVStore(k.storeKey)
	key := types.KeyAccount(account)
	// We probably need to create a proto for the authenticator. What should this contain? just type? (id and owner would be the keys)
	bz, err := k.cdc.MarshalInterface(authenticator)
	if err != nil {
		return err
	}
	store.Set(key, bz)
	return nil
}
