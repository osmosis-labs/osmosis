package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	gogotypes "github.com/gogo/protobuf/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/osmosis-labs/osmosis/osmoutils"

	"github.com/osmosis-labs/osmosis/v19/x/authenticator/authenticator"
	"github.com/osmosis-labs/osmosis/v19/x/authenticator/types"
)

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

type Keeper struct {
	storeKey   sdk.StoreKey
	cdc        codec.BinaryCodec
	paramSpace paramtypes.Subspace

	AuthenticatorManager *authenticator.AuthenticatorManager
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey sdk.StoreKey,
	ps paramtypes.Subspace,
	authenticatorManager *authenticator.AuthenticatorManager,
) Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		storeKey:             storeKey,
		cdc:                  cdc,
		paramSpace:           ps,
		AuthenticatorManager: authenticatorManager,
	}
}

func (k Keeper) GetAuthenticatorDataForAccount(
	ctx sdk.Context,
	account sdk.AccAddress,
) ([]*types.AccountAuthenticator, error) {
	accountAuthenticators, err := osmoutils.GatherValuesFromStorePrefix(
		ctx.KVStore(k.storeKey),
		types.KeyAccount(account),
		func(bz []byte) (*types.AccountAuthenticator, error) {
			// unmarshall the authenticator
			var authenticator types.AccountAuthenticator
			err := k.cdc.Unmarshal(bz, &authenticator)
			if err != nil {
				return &types.AccountAuthenticator{}, err
			}

			return &authenticator, nil
		})
	if err != nil {
		return nil, err
	}

	return accountAuthenticators, nil
}

func (k Keeper) GetAuthenticatorsForAccount(
	ctx sdk.Context,
	account sdk.AccAddress,
) ([]authenticator.Authenticator, error) {
	authenticatorData, err := k.GetAuthenticatorDataForAccount(ctx, account)
	if err != nil {
		return nil, err
	}
	authenticators := make([]authenticator.Authenticator, len(authenticatorData))
	for i, authenticator := range authenticatorData {
		authenticators[i] = authenticator.AsAuthenticator(k.AuthenticatorManager)
	}
	return authenticators, nil
}

// GetNextAuthenticatorId returns the next authenticator id.
func (k Keeper) GetNextAuthenticatorId(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	nextAuthenticatorId := gogotypes.UInt64Value{}
	found, err := osmoutils.Get(store, types.KeyNextAccountAuthenticatorId(), &nextAuthenticatorId)
	if err != nil {
		panic(err)
	}
	if !found {
		k.SetNextAuthenticatorId(ctx, 0)
		return 0
	}
	return nextAuthenticatorId.Value
}

// SetNextAuthenticatorId sets next authenticator id.
func (k Keeper) SetNextAuthenticatorId(ctx sdk.Context, authenticatorId uint64) {
	store := ctx.KVStore(k.storeKey)
	osmoutils.MustSet(store, types.KeyNextAccountAuthenticatorId(), &gogotypes.UInt64Value{Value: authenticatorId})
}

// GetNextAuthenticatorIdAndIncrement returns the next authenticator id and increments it.
func (k Keeper) GetNextAuthenticatorIdAndIncrement(ctx sdk.Context) uint64 {
	nextAuthenticatorId := k.GetNextAuthenticatorId(ctx)
	k.SetNextAuthenticatorId(ctx, nextAuthenticatorId+1)
	return nextAuthenticatorId
}

// AddAuthenticator adds an authenticator to an account
func (k Keeper) AddAuthenticator(ctx sdk.Context, account sdk.AccAddress, authenticatorType string, data []byte) error {
	if !k.AuthenticatorManager.IsAuthenticatorTypeRegistered(authenticatorType) {
		return fmt.Errorf("authenticator type %s is not registered", authenticatorType)
	}
	nextId := k.GetNextAuthenticatorIdAndIncrement(ctx)
	osmoutils.MustSet(ctx.KVStore(k.storeKey),
		types.KeyAccountId(account, nextId), // ToDo: will this lead to any concurrency issues?
		&types.AccountAuthenticator{
			Id:   nextId,
			Type: authenticatorType,
			Data: data,
		})
	return nil
}

// RemoveAuthenticator removes an authenticator from an account
func (k Keeper) RemoveAuthenticator(ctx sdk.Context, account sdk.AccAddress, authenticatorId uint64) error {
	store := ctx.KVStore(k.storeKey)
	key := types.KeyAccountId(account, authenticatorId)
	// check that the key exists
	if !store.Has(key) {
		return fmt.Errorf("authenticator with id %d does not exist for account %s", authenticatorId, account)
	}
	store.Delete(key)
	return nil
}

// ToDo: Open questions:
//  * Do we care about authenticator ordering?
