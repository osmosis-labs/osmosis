package keeper

import (
	"fmt"

	"github.com/osmosis-labs/osmosis/v21/x/authenticator/iface"

	"github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	gogotypes "github.com/gogo/protobuf/types"

	"github.com/osmosis-labs/osmosis/osmoutils"

	"github.com/osmosis-labs/osmosis/v21/x/authenticator/authenticator"
	"github.com/osmosis-labs/osmosis/v21/x/authenticator/types"
)

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

type Keeper struct {
	storeKey   storetypes.StoreKey
	cdc        codec.BinaryCodec
	paramSpace paramtypes.Subspace

	AuthenticatorManager *authenticator.AuthenticatorManager
	TransientStore       *authenticator.TransientStore
}

func NewKeeper(
	cdc codec.BinaryCodec,
	managerStoreKey storetypes.StoreKey,
	authenticatorStoreKey storetypes.StoreKey,
	ps paramtypes.Subspace,
	authenticatorManager *authenticator.AuthenticatorManager,
) Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		storeKey:             managerStoreKey,
		cdc:                  cdc,
		paramSpace:           ps,
		AuthenticatorManager: authenticatorManager,
		TransientStore:       authenticator.NewTransientStore(authenticatorStoreKey, sdk.Context{}),
	}
}

// GetAuthenticatorDataForAccount gets all authenticators AccAddressFromBech32 with an account
// from the store, the data is  prefixed by 2|<accAddr|
func (k Keeper) GetAuthenticatorDataForAccount(
	ctx sdk.Context,
	account sdk.AccAddress,
) ([]*types.AccountAuthenticator, error) {
	// unmarshalFn is used to unmarshal the AccountAuthenticator from the store
	unmarshalFn := func(bz []byte) (*types.AccountAuthenticator, error) {
		var authenticator types.AccountAuthenticator
		err := k.cdc.Unmarshal(bz, &authenticator)
		if err != nil {
			return &types.AccountAuthenticator{}, err
		}
		return &authenticator, nil
	}

	accountAuthenticators, err := osmoutils.GatherValuesFromStorePrefix(
		ctx.KVStore(k.storeKey),
		types.KeyAccount(account),
		unmarshalFn,
	)
	if err != nil {
		return nil, err
	}

	return accountAuthenticators, nil
}

// GetAuthenticatorsForAccount returns all the authenticators for the account
// this function relies in GetAuthenticationDataForAccount
func (k Keeper) GetAuthenticatorsForAccount(
	ctx sdk.Context,
	account sdk.AccAddress,
) ([]iface.Authenticator, error) {
	authenticatorData, err := k.GetAuthenticatorDataForAccount(ctx, account)
	if err != nil {
		return nil, err
	}

	authenticators := make([]iface.Authenticator, len(authenticatorData))
	for i, accountAuthenticator := range authenticatorData {
		authenticators[i] = accountAuthenticator.AsAuthenticator(k.AuthenticatorManager)
		if authenticators[i] == nil {
			return nil, fmt.Errorf("authenticator %d failed to initialize", accountAuthenticator.Id)
		}
	}
	return authenticators, nil
}

// GetAuthenticatorsForAccountOrDefault returns the authenticators for the account if there allRecords
// authenticators in the store, or the default if there is no authenticator associated with an account,
// this would be the case if there is an account with authenticators
// This function relies in GetAuthenticationsForAccount
func (k Keeper) GetAuthenticatorsForAccountOrDefault(
	ctx sdk.Context,
	account sdk.AccAddress,
) ([]iface.Authenticator, error) {
	authenticators, err := k.GetAuthenticatorsForAccount(ctx, account)
	if err != nil {
		return nil, err
	}

	if len(authenticators) == 0 {
		authenticators = append(authenticators, k.AuthenticatorManager.GetDefaultAuthenticator())
	}

	return authenticators, nil
}

// GetNextAuthenticatorId returns the next authenticator id
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

// AddAuthenticator adds an authenticator to an account, this function is used to add multiple
// authenticators such as SignatureVerificationAuthenticators and AllOfAuthenticators
func (k Keeper) AddAuthenticator(ctx sdk.Context, account sdk.AccAddress, authenticatorType string, data []byte) error {
	impl := k.AuthenticatorManager.GetAuthenticatorByType(authenticatorType)
	if impl == nil {
		return fmt.Errorf("authenticator type %s is not registered", authenticatorType)
	}

	err := impl.OnAuthenticatorAdded(ctx, account, data)

	if err != nil {
		return err
	}
	nextId := k.GetNextAuthenticatorIdAndIncrement(ctx)
	osmoutils.MustSet(ctx.KVStore(k.storeKey),
		types.KeyAccountId(account, nextId),
		&types.AccountAuthenticator{
			Id:   nextId,
			Type: authenticatorType,
			Data: data,
			// set this to false to skip the `ConfirmExecution` call on MsgAddAuthenticator for itself
			// it will be ready after `ConfirmExecution` on the aforementioned message is called.
			IsReady: false,
		})
	return nil
}

// MarkAuthenticatorAsReady sets an authenticator to be ready
func (k Keeper) MarkAuthenticatorAsReady(ctx sdk.Context, keyAccountId []byte) {
	store := ctx.KVStore(k.storeKey)

	var auth types.AccountAuthenticator
	osmoutils.MustGet(store, keyAccountId, &auth)

	auth.IsReady = true
	osmoutils.MustSet(store, keyAccountId, &auth)
}

// RemoveAuthenticator removes an authenticator from an account
func (k Keeper) RemoveAuthenticator(ctx sdk.Context, account sdk.AccAddress, authenticatorId uint64) error {
	store := ctx.KVStore(k.storeKey)
	key := types.KeyAccountId(account, authenticatorId)

	var existing types.AccountAuthenticator
	found, err := osmoutils.Get(store, key, &existing)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("authenticator with id %d does not exist for account %s", authenticatorId, account)
	}
	impl := k.AuthenticatorManager.GetAuthenticatorByType(existing.Type)
	if impl == nil {
		return fmt.Errorf("authenticator type %s is not registered", existing.Type)
	}

	// Authenticators can prevent removal. This should be used sparingly
	err = impl.OnAuthenticatorRemoved(ctx, account, existing.Data)
	if err != nil {
		return err
	}

	store.Delete(key)
	return nil
}

// GetAuthenticatorExtension unpacks the extension for the transaction, this is used with transactions specify
// an authenticator to use
func (k Keeper) GetAuthenticatorExtension(exts []*codectypes.Any) types.AuthenticatorTxOptions {
	var authExtension types.AuthenticatorTxOptions
	for _, ext := range exts {
		err := k.cdc.UnpackAny(ext, &authExtension)
		if err == nil {
			return authExtension
		}
	}
	return nil
}
